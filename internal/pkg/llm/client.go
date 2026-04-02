package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type Model string

const (
	ModelGemini15Pro  Model = "gemini-1.5-pro"
	ModelGPT4o        Model = "gpt-4o"
	ModelGPT4oMini    Model = "gpt-4o-mini"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Model       Model
	SystemPrompt string
	Messages    []Message
	Temperature float32
	MaxTokens   int
}

type Response struct {
	Content string
	Model   Model
	Usage   struct {
		InputTokens  int
		OutputTokens int
	}
}

type Client struct {
	httpClient *http.Client
	geminiKey  string
	openaiKey  string
}

func New() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 60 * time.Second},
		geminiKey:  os.Getenv("GEMINI_API_KEY"),
		openaiKey:  os.Getenv("OPENAI_API_KEY"),
	}
}

func (c *Client) Complete(ctx context.Context, req Request) (*Response, error) {
	switch req.Model {
	case ModelGemini15Pro:
		return c.callGemini(ctx, req)
	case ModelGPT4o, ModelGPT4oMini:
		return c.callOpenAI(ctx, req)
	default:
		return c.callGemini(ctx, req)
	}
}

func (c *Client) CompleteJSON(ctx context.Context, req Request, target any) error {
	resp, err := c.Complete(ctx, req)
	if err != nil {
		return fmt.Errorf("llm.CompleteJSON: %w", err)
	}
	if err := json.Unmarshal([]byte(resp.Content), target); err != nil {
		return fmt.Errorf("llm.CompleteJSON unmarshal: %w\nraw: %s", err, resp.Content)
	}
	return nil
}

// ── Gemini ────────────────────────────────────────────────────────────────

func (c *Client) callGemini(ctx context.Context, req Request) (*Response, error) {
	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		req.Model, c.geminiKey,
	)

	type part struct {
		Text string `json:"text"`
	}
	type content struct {
		Role  string `json:"role"`
		Parts []part `json:"parts"`
	}
	type generationConfig struct {
		Temperature     float32 `json:"temperature"`
		MaxOutputTokens int     `json:"maxOutputTokens"`
	}
	type payload struct {
		SystemInstruction *content         `json:"systemInstruction,omitempty"`
		Contents          []content        `json:"contents"`
		GenerationConfig  generationConfig `json:"generationConfig"`
	}

	p := payload{
		GenerationConfig: generationConfig{
			Temperature:     req.Temperature,
			MaxOutputTokens: req.MaxTokens,
		},
	}

	if req.SystemPrompt != "" {
		p.SystemInstruction = &content{Parts: []part{{Text: req.SystemPrompt}}}
	}

	for _, m := range req.Messages {
		role := m.Role
		if role == "assistant" {
			role = "model"
		}
		p.Contents = append(p.Contents, content{
			Role:  role,
			Parts: []part{{Text: m.Content}},
		})
	}

	body, _ := json.Marshal(p)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("gemini http: %w", err)
	}
	defer httpResp.Body.Close()

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
		} `json:"usageMetadata"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("gemini decode: %w", err)
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("gemini: empty response")
	}

	return &Response{
		Content: result.Candidates[0].Content.Parts[0].Text,
		Model:   req.Model,
		Usage: struct {
			InputTokens  int
			OutputTokens int
		}{
			InputTokens:  result.UsageMetadata.PromptTokenCount,
			OutputTokens: result.UsageMetadata.CandidatesTokenCount,
		},
	}, nil
}

// ── OpenAI ────────────────────────────────────────────────────────────────

func (c *Client) callOpenAI(ctx context.Context, req Request) (*Response, error) {
	type oaiMessage struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type payload struct {
		Model       string       `json:"model"`
		Messages    []oaiMessage `json:"messages"`
		Temperature float32      `json:"temperature"`
		MaxTokens   int          `json:"max_tokens"`
	}

	var msgs []oaiMessage
	if req.SystemPrompt != "" {
		msgs = append(msgs, oaiMessage{Role: "system", Content: req.SystemPrompt})
	}
	for _, m := range req.Messages {
		msgs = append(msgs, oaiMessage{Role: m.Role, Content: m.Content})
	}

	p := payload{
		Model:       string(req.Model),
		Messages:    msgs,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
	}

	body, _ := json.Marshal(p)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.openaiKey)

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("openai http: %w", err)
	}
	defer httpResp.Body.Close()

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("openai decode: %w", err)
	}
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("openai: empty choices")
	}

	return &Response{
		Content: result.Choices[0].Message.Content,
		Model:   req.Model,
		Usage: struct {
			InputTokens  int
			OutputTokens int
		}{
			InputTokens:  result.Usage.PromptTokens,
			OutputTokens: result.Usage.CompletionTokens,
		},
	}, nil
}

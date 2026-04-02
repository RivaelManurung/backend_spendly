package ai

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiClient struct {
	client *genai.Client
}

func NewGeminiClient(ctx context.Context, apiKey string) (*GeminiClient, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is required")
	}
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize gemini client: %w", err)
	}

	return &GeminiClient{client: client}, nil
}

func (g *GeminiClient) Close() {
	if g.client != nil {
		g.client.Close()
	}
}

// ExecutePrompt reads a .prompt file, replaces placeholders, and executes it.
func (g *GeminiClient) ExecutePrompt(ctx context.Context, promptName string, variables map[string]string, forceJSON bool) (string, error) {
	promptPath := filepath.Join(".github", "agents", "prompts", promptName+".prompt")
	promptBytes, err := os.ReadFile(promptPath)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt file %s: %w", promptPath, err)
	}

	promptText := string(promptBytes)

	// Replace placeholders like {{variable}}
	for k, v := range variables {
		placeholder := "{{" + k + "}}"
		promptText = strings.ReplaceAll(promptText, placeholder, v)
		// Also support {{variable | json}} for backward compatibility or future use
		placeholderJSON := "{{" + k + " | json}}"
		promptText = strings.ReplaceAll(promptText, placeholderJSON, v)
	}

	model := g.client.GenerativeModel("gemini-1.5-pro")
	if forceJSON {
		model.ResponseMIMEType = "application/json"
	}
	model.SetTemperature(0.1)

	resp, err := model.GenerateContent(ctx, genai.Text(promptText))
	if err != nil {
		return "", fmt.Errorf("gemini generation error for %s: %w", promptName, err)
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no response candidates for %s", promptName)
	}

	var result string
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			result += string(txt)
		}
	}

	// Clean up markdown code blocks if the model included them despite forceJSON
	result = strings.TrimPrefix(strings.TrimSpace(result), "```json")
	result = strings.TrimSuffix(result, "```")

	return result, nil
}

// Keep AutoCategorize for specific use case or refactor to use ExecutePrompt
func (g *GeminiClient) AutoCategorize(ctx context.Context, merchant, description string, amount float64, categoriesJSON string) (string, error) {
	vars := map[string]string{
		"merchant":    merchant,
		"description": description,
		"amount":      fmt.Sprintf("%v", amount),
		"categories":  categoriesJSON,
	}
	return g.ExecutePrompt(ctx, "categorize", vars, true)
}

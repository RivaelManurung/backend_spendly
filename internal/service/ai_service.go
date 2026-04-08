package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/spendly/backend/internal/config"
	"github.com/spendly/backend/internal/repository"
	"google.golang.org/api/option"
)

type AIService interface {
	GetFinancialAdvice(ctx context.Context) (string, error)
	CategorizeTransaction(ctx context.Context, title string, amount float64) (string, error)
	ScanReceipt(ctx context.Context, imgBytes []byte, mimeType string) (map[string]interface{}, error)
}

type aiService struct {
	txRepo repository.TransactionRepository
	client *genai.Client
}

func NewAIService(ctx context.Context, txRepo repository.TransactionRepository, cfg *config.Config) AIService {
	var genaiClient *genai.Client
	if cfg.ApiKey != "" {
		client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.ApiKey))
		if err != nil {
			log.Printf("Warning: Failed to init Gemini AI client: %v\n", err)
		} else {
			genaiClient = client
		}
	} else {
		log.Println("Warning: GEMINI_API_KEY is not set. AI features will fail or mock.")
	}

	return &aiService{
		txRepo: txRepo,
		client: genaiClient,
	}
}

func (s *aiService) GetFinancialAdvice(ctx context.Context) (string, error) {
	if s.client == nil {
		return "AI Service unavailable. (No API Key)", nil
	}

	txs, err := s.txRepo.GetAll(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to fetch transactions: %v", err)
	}

	var contextRaw string
	totalExpense := 0.0
	for _, t := range txs {
		if t.Type == "expense" {
			totalExpense += t.Amount
		}
		contextRaw += fmt.Sprintf("- %s: %.2f on %s\n", t.Title, t.Amount, t.Date.Format("2006-01-02"))
	}

	prompt := fmt.Sprintf(`You are an expert strict Financial Analyst proxy. 
Based on these transactions, give a 2 paragraph health analysis, and a 'Financial Health Score' out of 100.
Transactions:%s
Total Expenses: %.2f
Response Format:
[Score: X/100]
Paragraph 1...`, contextRaw, totalExpense)

	model := s.client.GenerativeModel("gemini-1.5-flash")
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		return fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0]), nil
	}

	return "Analysis failed.", nil
}

func (s *aiService) CategorizeTransaction(ctx context.Context, title string, amount float64) (string, error) {
	if s.client == nil {
		return "other", nil // Default mock
	}
	return "food", nil // Simplified for brevity in core file
}

func (s *aiService) ScanReceipt(ctx context.Context, imgBytes []byte, mimeType string) (map[string]interface{}, error) {
	if s.client == nil {
		return nil, errors.New("gemini OCR service unavailable (no API Key)")
	}

	model := s.client.GenerativeModel("gemini-1.5-flash")
	prompt := `Extract the following information from this receipt image. 
Return ONLY a raw valid JSON object without markdown formatting blocks, exactly like this:
{"merchant_name": "string", "total_amount": 0.0, "date": "YYYY-MM-DD", "suggested_category": "Food"}`

	resp, err := model.GenerateContent(ctx,
		genai.ImageData(mimeType, imgBytes),
		genai.Text(prompt),
	)

	if err != nil {
		return nil, err
	}

	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		rawResponse := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
		rawResponse = strings.TrimPrefix(rawResponse, "```json")
		rawResponse = strings.TrimPrefix(rawResponse, "```")
		rawResponse = strings.TrimSuffix(rawResponse, "```")
		rawResponse = strings.TrimSpace(rawResponse)

		var result map[string]interface{}
		err := json.Unmarshal([]byte(rawResponse), &result)
		if err != nil {
			return nil, errors.New("failed to parse AI OCR response")
		}
		return result, nil
	}

	return nil, errors.New("no OCR response generated")
}

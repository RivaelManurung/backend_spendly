package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spendly/backend/internal/ai"
	"github.com/spendly/backend/internal/domain"
)

type AiCategorizationService struct {
	geminiClient *ai.GeminiClient
}

func NewAiCategorizationService(gemini *ai.GeminiClient) *AiCategorizationService {
	return &AiCategorizationService{
		geminiClient: gemini,
	}
}

type CategorizeResponse struct {
	CategoryID   int     `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Confidence   float32 `json:"confidence"`
	Reasoning    string  `json:"reasoning"`
}

func (s *AiCategorizationService) AutoCategorize(ctx context.Context, txn *domain.Transaction, availableCategories []domain.Category) error {
	catData, _ := json.Marshal(availableCategories)

	var merchantStr, descStr string
	if txn.Merchant != nil {
		merchantStr = *txn.Merchant
	}
	if txn.Description != nil {
		descStr = *txn.Description
	}

	amountInBaseFloat, _ := txn.AmountInBase.Float64()
	resStr, err := s.geminiClient.AutoCategorize(ctx, merchantStr, descStr, amountInBaseFloat, string(catData))
	if err != nil {
		return err
	}

	var resp CategorizeResponse
	if err := json.Unmarshal([]byte(resStr), &resp); err != nil {
		return fmt.Errorf("failed to parse AI response: %w", err)
	}

	txn.AICategorySuggestion = &resp.CategoryName
	txn.AIConfidenceScore = &resp.Confidence

	if resp.Confidence >= 0.90 && resp.CategoryID != 0 {
		catID := int64(resp.CategoryID)
		txn.CategoryID = &catID
	}

	return nil
}

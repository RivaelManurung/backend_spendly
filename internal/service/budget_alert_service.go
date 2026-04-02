package service

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/spendly/backend/internal/ai"
	"github.com/spendly/backend/internal/domain"
)

// 7. Budget Alert (budget_alert_pipeline)
type BudgetPipeline struct {
	gemini *ai.GeminiClient
}

func NewBudgetPipeline(g *ai.GeminiClient) *BudgetPipeline {
	return &BudgetPipeline{gemini: g}
}

// GenerateAlert me-run budget_alert.prompt di .agent ketika warning 80% / 100% threshold terpenuhi
func (b *BudgetPipeline) GenerateAlert(ctx context.Context, userID uuid.UUID, budget domain.Budget, currentSpent string) (string, error) {
	budgetBytes, _ := json.Marshal(budget)
	
	res, err := b.gemini.ExecutePrompt(ctx, "budget_alert", map[string]string{
		"budget_info": string(budgetBytes),
		"current_spent": currentSpent,
	}, false)

	return res, err // Biasanya langsung di-emit via push notification
}

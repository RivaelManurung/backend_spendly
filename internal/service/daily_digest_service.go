package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/spendly/backend/internal/ai"
	"github.com/spendly/backend/internal/domain"
	"github.com/spendly/backend/internal/repository"
)

// DailyDigestService generates a daily spending summary using AI.
// Triggered every morning at 07:00 via scheduler.
type DailyDigestService struct {
	gemini      *ai.GeminiClient
	repoTxn     repository.TransactionRepository
	repoInsight repository.InsightRepository
	repoBudget  repository.BudgetRepository
}

func NewDailyDigestService(
	gemini *ai.GeminiClient,
	repoTxn repository.TransactionRepository,
	repoInsight repository.InsightRepository,
	repoBudget repository.BudgetRepository,
) *DailyDigestService {
	return &DailyDigestService{
		gemini:      gemini,
		repoTxn:     repoTxn,
		repoInsight: repoInsight,
		repoBudget:  repoBudget,
	}
}

// RunForUser generates and saves a daily digest insight for a single user.
func (s *DailyDigestService) RunForUser(ctx context.Context, userID uuid.UUID, currency string) error {
	// 1. Fetch yesterday's transactions
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	startOfYesterday := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location())
	endOfYesterday := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 23, 59, 59, 0, yesterday.Location())

	txns, err := s.repoTxn.GetByDateRange(ctx, userID, startOfYesterday, endOfYesterday)
	if err != nil {
		return fmt.Errorf("DailyDigestService.RunForUser fetch txns: %w", err)
	}

	if len(txns) == 0 {
		// No transactions yesterday — no insight needed
		return nil
	}

	// 2. Build per-category summary for budget status
	spentByCategory := make(map[string]decimal.Decimal)
	for _, t := range txns {
		if t.CategoryID != nil && !t.Amount.IsPositive() {
			key := fmt.Sprintf("%d", *t.CategoryID)
			spentByCategory[key] = spentByCategory[key].Add(t.AmountInBase.Abs())
		}
	}

	// 3. Build variables for prompt
	txnsJSON, _ := json.Marshal(txns)
	spentJSON, _ := json.Marshal(spentByCategory)

	vars := map[string]string{
		"transactions":     string(txnsJSON),
		"accounts_balance": `{}`, // TODO: inject account balances once AccountRepository is wired
		"budget_status":    string(spentJSON),
		"upcoming_recurring": `[]`, // TODO: inject from RecurringRepository
		"user_currency":    currency,
	}

	// 4. Execute AI prompt
	result, err := s.gemini.ExecutePrompt(ctx, "daily_digest", vars, true)
	if err != nil {
		return fmt.Errorf("DailyDigestService.RunForUser AI prompt: %w", err)
	}

	// 5. Parse AI response to extract title for display
	var parsed struct {
		Summary string `json:"summary"`
		Warning string `json:"warning"`
	}
	_ = json.Unmarshal([]byte(result), &parsed)

	title := "Ringkasan Harian"
	if parsed.Summary != "" {
		title = parsed.Summary
	}

	// 6. Save insight
	insight := domain.AIInsight{
		UserID:   userID,
		Type:     "DAILY_DIGEST",
		Title:    title,
		Content:  result,
		Priority: 1, // Lower priority than anomalies
	}

	return s.repoInsight.SaveInsight(ctx, &insight)
}

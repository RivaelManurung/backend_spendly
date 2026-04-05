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

// NetWorthService handles asset graph computation and AI net worth analysis.
// Responsible for:
// 1. Taking a monthly snapshot of all account balances (for Asset Graph)
// 2. Running the net_worth_analysis.prompt for wealth trajectory insight
type NetWorthService struct {
	gemini      *ai.GeminiClient
	repoInsight repository.InsightRepository
	repoUser    repository.UserRepository
}

func NewNetWorthService(
	gemini *ai.GeminiClient,
	repoInsight repository.InsightRepository,
	repoUser repository.UserRepository,
) *NetWorthService {
	return &NetWorthService{
		gemini:      gemini,
		repoInsight: repoInsight,
		repoUser:    repoUser,
	}
}

// AccountBalanceSummary is a lightweight view of account balances passed to the AI.
type AccountBalanceSummary struct {
	AccountName string          `json:"account_name"`
	Type        string          `json:"type"`
	Balance     decimal.Decimal `json:"balance"`
	Currency    string          `json:"currency"`
	IsDebt      bool            `json:"is_debt"` // true for credit/loan accounts
}

// NetWorthResult is computed in-memory from account balances.
type NetWorthResult struct {
	TotalAssets decimal.Decimal         `json:"total_assets"`
	TotalDebt   decimal.Decimal         `json:"total_debt"`
	NetWorth    decimal.Decimal         `json:"net_worth"`
	Accounts    []AccountBalanceSummary `json:"accounts"`
	RecordedAt  time.Time               `json:"recorded_at"`
}

// ComputeNetWorth calculates net worth from a slice of account summaries.
// This can be called by a scheduler or an HTTP handler.
func (s *NetWorthService) ComputeNetWorth(accounts []AccountBalanceSummary) NetWorthResult {
	var totalAssets, totalDebt decimal.Decimal
	for _, a := range accounts {
		if a.IsDebt {
			totalDebt = totalDebt.Add(a.Balance)
		} else {
			totalAssets = totalAssets.Add(a.Balance)
		}
	}
	return NetWorthResult{
		TotalAssets: totalAssets,
		TotalDebt:   totalDebt,
		NetWorth:    totalAssets.Sub(totalDebt),
		Accounts:    accounts,
		RecordedAt:  time.Now(),
	}
}

// RunAnalysisForUser fetches account data, computes net worth, runs AI, and saves insight.
// In a full implementation, accounts would come from AccountRepository.
// For now we accept a pre-built slice for testability.
func (s *NetWorthService) RunAnalysisForUser(
	ctx context.Context,
	userID uuid.UUID,
	currentAccounts []AccountBalanceSummary,
	historicalSnapshots []NetWorthResult,
) error {
	if len(currentAccounts) == 0 {
		return nil
	}

	// 1. Compute current net worth
	current := s.ComputeNetWorth(currentAccounts)

	// 2. Fetch user profile for personalization
	user, err := s.repoUser.GetByID(ctx, userID)
	if err != nil || user == nil {
		return fmt.Errorf("NetWorthService.RunAnalysisForUser: user not found: %w", err)
	}

	// 3. Build prompt variables
	currentJSON, _ := json.Marshal(current)
	historyJSON, _ := json.Marshal(historicalSnapshots)
	accountsJSON, _ := json.Marshal(currentAccounts)
	goalsJSON, _ := json.Marshal(user.FinancialGoals)

	vars := map[string]string{
		"net_worth_current":  string(currentJSON),
		"net_worth_history":  string(historyJSON),
		"accounts_detail":    string(accountsJSON),
		"user.risk_profile":  user.RiskProfile,
		"user.financial_goals": string(goalsJSON),
		"user.currency":      user.CurrencyPreference,
	}

	// 4. Run AI net_worth_analysis prompt
	result, err := s.gemini.ExecutePrompt(ctx, "net_worth_analysis", vars, true)
	if err != nil {
		return fmt.Errorf("NetWorthService.RunAnalysisForUser AI: %w", err)
	}

	// 5. Parse response for priority decision
	var parsed struct {
		Warning string  `json:"warning"`
		Score   float64 `json:"supporting_data.financial_health_score"`
	}
	_ = json.Unmarshal([]byte(result), &parsed)

	priority := 2
	title := "📊 Analisa Kekayaan Bersih"
	if parsed.Warning != "" {
		priority = 3
		title = "⚠️ Peringatan Kondisi Keuangan"
	}

	// 6. Save insight to database
	insight := domain.AIInsight{
		UserID:   userID,
		Type:     "NET_WORTH_ANALYSIS",
		Title:    title,
		Content:  result,
		Priority: priority,
	}

	return s.repoInsight.SaveInsight(ctx, &insight)
}

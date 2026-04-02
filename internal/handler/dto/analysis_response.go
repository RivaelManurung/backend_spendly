package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/spendly/backend/internal/domain"
)

type AnalysisSnapshotResponse struct {
	ID                 int64                      `json:"id"`
	UserID             uuid.UUID                  `json:"user_id"`
	PeriodType         string                     `json:"period_type"`
	PeriodStart        time.Time                  `json:"period_start"`
	PeriodEnd          time.Time                  `json:"period_end"`
	TotalIncome        decimal.Decimal            `json:"total_income"`
	TotalExpense       decimal.Decimal            `json:"total_expense"`
	NetCashflow        decimal.Decimal            `json:"net_cashflow"`
	TopExpenseCategory *string                    `json:"top_expense_category,omitempty"`
	CategoryBreakdown  map[string]decimal.Decimal `json:"category_breakdown"`
	CreatedAt          time.Time                  `json:"created_at"`
}

func FromAnalysisSnapshot(s *domain.AnalysisSnapshot) *AnalysisSnapshotResponse {
	if s == nil {
		return nil
	}
	return &AnalysisSnapshotResponse{
		ID:                 s.ID,
		UserID:             s.UserID,
		PeriodType:         s.PeriodType,
		PeriodStart:        s.PeriodStart,
		PeriodEnd:          s.PeriodEnd,
		TotalIncome:        s.TotalIncome,
		TotalExpense:       s.TotalExpense,
		NetCashflow:        s.NetCashflow,
		TopExpenseCategory: s.TopExpenseCategory,
		CategoryBreakdown:  s.CategoryBreakdown,
		CreatedAt:          s.CreatedAt,
	}
}

type AIInsightResponse struct {
	ID        int64     `json:"id"`
	Type      string    `json:"type"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

func FromAIInsight(i *domain.AIInsight) *AIInsightResponse {
	if i == nil {
		return nil
	}
	return &AIInsightResponse{
		ID:        i.ID,
		Type:      i.Type,
		Content:   i.Content,
		CreatedAt: i.CreatedAt,
	}
}

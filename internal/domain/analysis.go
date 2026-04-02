package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type AnalysisSnapshot struct {
	ID                 int64                      `db:"id" json:"id"`
	UserID             uuid.UUID                  `db:"user_id" json:"user_id"`
	PeriodType         string                     `db:"period_type" json:"period_type"` // e.g. "MONTHLY"
	PeriodValue        string                     `db:"period_value" json:"period_value"` // e.g. "2025-01"
	PeriodStart        time.Time                  `db:"period_start" json:"period_start"`
	PeriodEnd          time.Time                  `db:"period_end" json:"period_end"`
	TotalIncome        decimal.Decimal            `db:"total_income" json:"total_income"`
	TotalExpense       decimal.Decimal            `db:"total_expense" json:"total_expense"`
	NetSavings         decimal.Decimal            `db:"net_savings" json:"net_savings"`
	NetCashflow        decimal.Decimal            `db:"net_cashflow" json:"net_cashflow"`
	TransactionCount   int                        `db:"transaction_count" json:"transaction_count"`
	TopExpenseCategory *string                    `db:"top_expense_category" json:"top_expense_category"`
	TopCategories      json.RawMessage            `db:"top_categories" json:"top_categories"`
	CategoryBreakdown  map[string]decimal.Decimal `db:"category_breakdown" json:"category_breakdown"`
	MerchantBreakdown  map[string]decimal.Decimal `db:"merchant_breakdown" json:"merchant_breakdown"`
	DailyTrend        map[string]decimal.Decimal `db:"daily_trend" json:"daily_trend"`
	Status            string                     `db:"status" json:"status"` // e.g. "PENDING", "COMPLETED"
	Metadata          json.RawMessage            `db:"metadata" json:"metadata"`
	CreatedAt         time.Time                  `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time                  `db:"updated_at" json:"updated_at"`
}

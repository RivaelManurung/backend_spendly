package dto

import "github.com/shopspring/decimal"

// BudgetCreateRequest is used for creating a new budget.
type BudgetCreateRequest struct {
	CategoryID  int64           `json:"category_id" binding:"required"`
	Period      string          `json:"period" binding:"required" example:"2025-01"`
	LimitAmount decimal.Decimal `json:"limit_amount" binding:"required"`
	Currency    string          `json:"currency" binding:"required"`
}

// BudgetUpdateRequest is used for updating an existing budget.
type BudgetUpdateRequest struct {
	LimitAmount *decimal.Decimal `json:"limit_amount"`
	IsActive    *bool            `json:"is_active"`
}

// BudgetResponse is a data structure for sending budget data back to the API.
type BudgetResponse struct {
	ID          int64           `json:"id"`
	CategoryID  int64           `json:"category_id"`
	Period      string          `json:"period"`
	LimitAmount decimal.Decimal `json:"limit_amount"`
	SpentAmount float64         `json:"spent_amount"`
	Currency    string          `json:"currency"`
	IsActive    bool            `json:"is_active"`
	UsagePercent float64        `json:"usage_percent"`
}

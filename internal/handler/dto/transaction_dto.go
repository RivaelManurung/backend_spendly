package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

// TransactionCreateRequest is used for creating a new transaction.
type TransactionCreateRequest struct {
	CategoryID      *int64          `json:"category_id"`
	Amount          decimal.Decimal `json:"amount" binding:"required"`
	Currency        string          `json:"currency" binding:"required"`
	Description     *string         `json:"description"`
	Merchant        *string         `json:"merchant"`
	Source          string          `json:"source" binding:"required"` // e.g. "MANUAL", "BANK_SYNC"
	TransactionDate time.Time       `json:"transaction_date" binding:"required"`
	Metadata        interface{}     `json:"metadata"`
}

// TransactionUpdateRequest is used for updating an existing transaction.
type TransactionUpdateRequest struct {
	CategoryID      *int64          `json:"category_id"`
	Amount          *decimal.Decimal `json:"amount"`
	Currency        *string         `json:"currency"`
	Description     *string         `json:"description"`
	Merchant        *string         `json:"merchant"`
	TransactionDate *time.Time       `json:"transaction_date"`
}

// TransactionResponse is a data structure for sending transaction data back to the API.
type TransactionResponse struct {
	ID                   string          `json:"id"`
	CategoryID           *int64          `json:"category_id"`
	Amount               decimal.Decimal `json:"amount"`
	Currency             string          `json:"currency"`
	AmountInBase         decimal.Decimal `json:"amount_in_base"`
	Description          *string         `json:"description"`
	Merchant             *string         `json:"merchant"`
	Source               string          `json:"source"`
	TransactionDate      time.Time       `json:"transaction_date"`
	AICategorySuggestion *string         `json:"ai_category_suggestion"`
	AIConfidenceScore    *float32        `json:"ai_confidence_score"`
	CreatedAt            time.Time       `json:"created_at"`
}

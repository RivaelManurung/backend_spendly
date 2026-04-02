package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Transaction struct {
	ID                   uuid.UUID       `db:"id" json:"id"`
	UserID               uuid.UUID       `db:"user_id" json:"user_id"`
	CategoryID           *int64          `db:"category_id" json:"category_id,omitempty"`
	Amount               decimal.Decimal `db:"amount" json:"amount"`
	Currency             string          `db:"currency" json:"currency"`
	AmountInBase         decimal.Decimal `db:"amount_in_base" json:"amount_in_base"`
	Description          *string         `db:"description" json:"description,omitempty"`
	Merchant             *string         `db:"merchant" json:"merchant,omitempty"`
	Source               string          `db:"source" json:"source"`
	TransactionDate      time.Time       `db:"transaction_date" json:"transaction_date"`
	AICategorySuggestion *string         `db:"ai_category_suggestion" json:"ai_category_suggestion,omitempty"`
	AIConfidenceScore    *float32        `db:"ai_confidence_score" json:"ai_confidence_score,omitempty"`
	Metadata             json.RawMessage `db:"metadata" json:"metadata,omitempty" swaggertype:"string"`
	CreatedAt            time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time       `db:"updated_at" json:"updated_at"`
	DeletedAt            *time.Time      `db:"deleted_at" json:"deleted_at,omitempty"`
}

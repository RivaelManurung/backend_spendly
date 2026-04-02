package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Budget struct {
	ID          int64           `db:"id" json:"id"`
	UserID      uuid.UUID       `db:"user_id" json:"user_id"`
	CategoryID  *int64          `db:"category_id" json:"category_id"`
	Period      string          `db:"period" json:"period"` // e.g. "2025-01"
	LimitAmount decimal.Decimal `db:"limit_amount" json:"limit_amount"`
	Currency    string          `db:"currency" json:"currency"`
	IsActive    bool            `db:"is_active" json:"is_active"`
	CreatedAt   time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at" json:"updated_at"`
}

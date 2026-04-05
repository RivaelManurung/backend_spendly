package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// RecurringFrequency defines how often a transaction repeats
type RecurringFrequency string

const (
	FrequencyDaily   RecurringFrequency = "daily"
	FrequencyWeekly  RecurringFrequency = "weekly"
	FrequencyMonthly RecurringFrequency = "monthly"
	FrequencyYearly  RecurringFrequency = "yearly"
)

// RecurringTransaction is a template/bookmark for automatically repeating transactions
// Inspired by Money Manager's "Recurring Transactions" feature
type RecurringTransaction struct {
	ID          int64              `db:"id" json:"id"`
	UserID      uuid.UUID          `db:"user_id" json:"user_id"`
	AccountID   *int64             `db:"account_id" json:"account_id,omitempty"`
	CategoryID  *int64             `db:"category_id" json:"category_id,omitempty"`
	Title       string             `db:"title" json:"title"`         // e.g. "Bayar Netflix", "Cicilan KPR"
	Amount      decimal.Decimal    `db:"amount" json:"amount"`
	Currency    string             `db:"currency" json:"currency"`
	Type        string             `db:"type" json:"type"`           // "income" | "expense"
	Frequency   RecurringFrequency `db:"frequency" json:"frequency"`
	StartDate   time.Time          `db:"start_date" json:"start_date"`
	EndDate     *time.Time         `db:"end_date" json:"end_date,omitempty"` // nil = recurring forever
	NextDueDate time.Time          `db:"next_due_date" json:"next_due_date"`
	LastRunAt   *time.Time         `db:"last_run_at" json:"last_run_at,omitempty"`
	IsActive    bool               `db:"is_active" json:"is_active"`
	AutoPost    bool               `db:"auto_post" json:"auto_post"` // true = post otomatis, false = remind user
	Notes       *string            `db:"notes" json:"notes,omitempty"`
	Metadata    json.RawMessage    `db:"metadata" json:"metadata,omitempty"`
	CreatedAt   time.Time          `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `db:"updated_at" json:"updated_at"`
}

package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// AccountType represents the type of financial account
type AccountType string

const (
	AccountTypeCash        AccountType = "cash"
	AccountTypeDebit       AccountType = "debit"
	AccountTypeCredit      AccountType = "credit"
	AccountTypeSavings     AccountType = "savings"
	AccountTypeInvestment  AccountType = "investment"
	AccountTypeLoan        AccountType = "loan"
	AccountTypeEWallet     AccountType = "e_wallet"
)

// Account represents a user's financial account (like Money Manager accounts)
type Account struct {
	ID             int64           `db:"id" json:"id"`
	UserID         uuid.UUID       `db:"user_id" json:"user_id"`
	Name           string          `db:"name" json:"name"`                     // e.g. "BCA Tabungan", "Gopay", "Dompet Tunai"
	Type           AccountType     `db:"type" json:"type"`
	InitialBalance decimal.Decimal `db:"initial_balance" json:"initial_balance"` // Saldo awal saat akun dibuat
	CurrentBalance decimal.Decimal `db:"current_balance" json:"current_balance"` // Saldo real-time (dihitung dari transaksi)
	Currency       string          `db:"currency" json:"currency"`
	Color          string          `db:"color" json:"color"`     // Warna di UI (hex)
	Icon           string          `db:"icon" json:"icon"`       // Icon identifier
	IsActive       bool            `db:"is_active" json:"is_active"`
	ExcludeFromTotal bool          `db:"exclude_from_total" json:"exclude_from_total"` // Untuk akun pinjaman yg tidak ingin dihitung
	CreditLimit    *decimal.Decimal `db:"credit_limit" json:"credit_limit,omitempty"` // Hanya untuk credit card
	PaymentDueDay  *int            `db:"payment_due_day" json:"payment_due_day,omitempty"` // Tanggal jatuh tempo kartu kredit
	Notes          *string         `db:"notes" json:"notes,omitempty"`
	SortOrder      int             `db:"sort_order" json:"sort_order"`
	CreatedAt      time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time       `db:"updated_at" json:"updated_at"`
	DeletedAt      *time.Time      `db:"deleted_at" json:"deleted_at,omitempty"`
}

// AccountTransfer represents an internal transfer between two accounts (Double-entry)
type AccountTransfer struct {
	ID            int64           `db:"id" json:"id"`
	UserID        uuid.UUID       `db:"user_id" json:"user_id"`
	FromAccountID int64           `db:"from_account_id" json:"from_account_id"`
	ToAccountID   int64           `db:"to_account_id" json:"to_account_id"`
	Amount        decimal.Decimal `db:"amount" json:"amount"`
	Fee           decimal.Decimal `db:"fee" json:"fee"`   // Transfer fee (e.g. BI-Fast 2500)
	Currency      string          `db:"currency" json:"currency"`
	Notes         *string         `db:"notes" json:"notes,omitempty"`
	TransferDate  time.Time       `db:"transfer_date" json:"transfer_date"`
	CreatedAt     time.Time       `db:"created_at" json:"created_at"`

	// Joined for display
	FromAccountName *string `db:"-" json:"from_account_name,omitempty"`
	ToAccountName   *string `db:"-" json:"to_account_name,omitempty"`
}

// AccountSnapshot represents a point-in-time balance snapshot for Asset Graph
type AccountSnapshot struct {
	ID         int64           `db:"id" json:"id"`
	AccountID  int64           `db:"account_id" json:"account_id"`
	UserID     uuid.UUID       `db:"user_id" json:"user_id"`
	Balance    decimal.Decimal `db:"balance" json:"balance"`
	RecordedAt time.Time       `db:"recorded_at" json:"recorded_at"` // End of day/month snapshot
}

// NetWorthSnapshot represents total asset across all accounts at a point in time
type NetWorthSnapshot struct {
	UserID      uuid.UUID       `db:"user_id" json:"user_id"`
	TotalAssets decimal.Decimal `db:"total_assets" json:"total_assets"` // Cash + Bank + Investment + E-Wallet
	TotalDebt   decimal.Decimal `db:"total_debt" json:"total_debt"`     // Credit + Loan
	NetWorth    decimal.Decimal `db:"net_worth" json:"net_worth"`       // Assets - Debt
	RecordedAt  time.Time       `db:"recorded_at" json:"recorded_at"`
}

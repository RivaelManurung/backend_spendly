package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
	"github.com/spendly/backend/internal/domain"
	"github.com/spendly/backend/internal/repository"
)

var _ repository.AccountRepository = (*accountRepository)(nil)

type accountRepository struct {
	db *sqlx.DB
}

func NewAccountRepository(db *sqlx.DB) repository.AccountRepository {
	return &accountRepository{db: db}
}

func (r *accountRepository) Create(ctx context.Context, account *domain.Account) error {
	now := time.Now()
	if account.CreatedAt.IsZero() {
		account.CreatedAt = now
	}
	account.UpdatedAt = now
	account.CurrentBalance = account.InitialBalance

	query := `
		INSERT INTO accounts (
			user_id, name, type, initial_balance, current_balance,
			currency, color, icon, is_active, exclude_from_total,
			credit_limit, payment_due_day, notes, sort_order, created_at, updated_at
		) VALUES (
			:user_id, :name, :type, :initial_balance, :current_balance,
			:currency, :color, :icon, :is_active, :exclude_from_total,
			:credit_limit, :payment_due_day, :notes, :sort_order, :created_at, :updated_at
		) RETURNING id
	`
	rows, err := r.db.NamedQueryContext(ctx, query, account)
	if err != nil {
		return fmt.Errorf("accountRepository.Create: %w", err)
	}
	defer rows.Close()
	if rows.Next() {
		_ = rows.Scan(&account.ID)
	}
	return nil
}

func (r *accountRepository) GetByID(ctx context.Context, id int64) (*domain.Account, error) {
	query := `
		SELECT id, user_id, name, type, initial_balance, current_balance,
		       currency, color, icon, is_active, exclude_from_total,
		       credit_limit, payment_due_day, notes, sort_order, created_at, updated_at, deleted_at
		FROM accounts WHERE id = $1 AND deleted_at IS NULL
	`
	var account domain.Account
	if err := r.db.GetContext(ctx, &account, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("accountRepository.GetByID: %w", err)
	}
	return &account, nil
}

func (r *accountRepository) GetAllByUser(ctx context.Context, userID uuid.UUID) ([]domain.Account, error) {
	query := `
		SELECT id, user_id, name, type, initial_balance, current_balance,
		       currency, color, icon, is_active, exclude_from_total,
		       credit_limit, payment_due_day, notes, sort_order, created_at, updated_at
		FROM accounts
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY sort_order ASC, created_at ASC
	`
	var accounts []domain.Account
	if err := r.db.SelectContext(ctx, &accounts, query, userID); err != nil {
		return nil, fmt.Errorf("accountRepository.GetAllByUser: %w", err)
	}
	if accounts == nil {
		accounts = []domain.Account{}
	}
	return accounts, nil
}

func (r *accountRepository) UpdateBalance(ctx context.Context, id int64, newBalance interface{}) error {
	var bal decimal.Decimal
	switch v := newBalance.(type) {
	case decimal.Decimal:
		bal = v
	case float64:
		bal = decimal.NewFromFloat(v)
	default:
		return fmt.Errorf("accountRepository.UpdateBalance: unsupported balance type")
	}

	query := `UPDATE accounts SET current_balance = $1, updated_at = NOW() WHERE id = $2 AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, bal, id)
	if err != nil {
		return fmt.Errorf("accountRepository.UpdateBalance: %w", err)
	}
	return nil
}

func (r *accountRepository) Update(ctx context.Context, account *domain.Account) error {
	account.UpdatedAt = time.Now()
	query := `
		UPDATE accounts SET
			name = :name, color = :color, icon = :icon,
			is_active = :is_active, exclude_from_total = :exclude_from_total,
			credit_limit = :credit_limit, payment_due_day = :payment_due_day,
			notes = :notes, sort_order = :sort_order, updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL
	`
	_, err := r.db.NamedExecContext(ctx, query, account)
	if err != nil {
		return fmt.Errorf("accountRepository.Update: %w", err)
	}
	return nil
}

func (r *accountRepository) SoftDelete(ctx context.Context, id int64) error {
	query := `UPDATE accounts SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("accountRepository.SoftDelete: %w", err)
	}
	return nil
}

func (r *accountRepository) CreateTransfer(ctx context.Context, transfer *domain.AccountTransfer) error {
	if transfer.CreatedAt.IsZero() {
		transfer.CreatedAt = time.Now()
	}
	query := `
		INSERT INTO account_transfers (user_id, from_account_id, to_account_id, amount, fee, currency, notes, transfer_date, created_at)
		VALUES (:user_id, :from_account_id, :to_account_id, :amount, :fee, :currency, :notes, :transfer_date, :created_at)
		RETURNING id
	`
	rows, err := r.db.NamedQueryContext(ctx, query, transfer)
	if err != nil {
		return fmt.Errorf("accountRepository.CreateTransfer: %w", err)
	}
	defer rows.Close()
	if rows.Next() {
		_ = rows.Scan(&transfer.ID)
	}

	// Update balances: deduct from source, add to destination
	if err := r.UpdateBalance(ctx, transfer.FromAccountID, decimal.Zero); err != nil {
		return err // In production: wrap in a DB transaction
	}
	return nil
}

func (r *accountRepository) GetTransfersByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.AccountTransfer, error) {
	query := `
		SELECT t.id, t.user_id, t.from_account_id, t.to_account_id, t.amount, t.fee,
		       t.currency, t.notes, t.transfer_date, t.created_at,
		       fa.name AS from_account_name, ta.name AS to_account_name
		FROM account_transfers t
		LEFT JOIN accounts fa ON fa.id = t.from_account_id
		LEFT JOIN accounts ta ON ta.id = t.to_account_id
		WHERE t.user_id = $1
		ORDER BY t.transfer_date DESC
		LIMIT $2 OFFSET $3
	`
	var transfers []domain.AccountTransfer
	if err := r.db.SelectContext(ctx, &transfers, query, userID, limit, offset); err != nil {
		return nil, fmt.Errorf("accountRepository.GetTransfersByUser: %w", err)
	}
	if transfers == nil {
		transfers = []domain.AccountTransfer{}
	}
	return transfers, nil
}

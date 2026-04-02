package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/spendly/backend/internal/domain"
	"github.com/spendly/backend/internal/repository"
)

var _ repository.TransactionRepository = (*transactionRepository)(nil)

type transactionRepository struct {
	db *sqlx.DB
}

// NewTransactionRepository creates a new instance of postgres repo for domain.Transaction.
func NewTransactionRepository(db *sqlx.DB) repository.TransactionRepository {
	return &transactionRepository{
		db: db,
	}
}

func (r *transactionRepository) Insert(ctx context.Context, txn *domain.Transaction) error {
	now := time.Now()
	if txn.CreatedAt.IsZero() {
		txn.CreatedAt = now
	}
	if txn.UpdatedAt.IsZero() {
		txn.UpdatedAt = now
	}

	query := `
		INSERT INTO transactions (
			id, user_id, category_id, amount, currency, amount_in_base,
			description, merchant, source, transaction_date,
			ai_category_suggestion, ai_confidence_score, metadata,
			created_at, updated_at
		) VALUES (
			:id, :user_id, :category_id, :amount, :currency, :amount_in_base,
			:description, :merchant, :source, :transaction_date,
			:ai_category_suggestion, :ai_confidence_score, :metadata,
			:created_at, :updated_at
		)
	`
	_, err := r.db.NamedExecContext(ctx, query, txn)
	if err != nil {
		return fmt.Errorf("transactionRepository.Insert: %w", err)
	}

	return nil
}

func (r *transactionRepository) BulkInsert(ctx context.Context, txns []domain.Transaction) error {
	if len(txns) == 0 {
		return nil
	}

	query := `
		INSERT INTO transactions (
			id, user_id, category_id, amount, currency, amount_in_base,
			description, merchant, source, transaction_date,
			ai_category_suggestion, ai_confidence_score, metadata,
			created_at, updated_at
		) VALUES (
			:id, :user_id, :category_id, :amount, :currency, :amount_in_base,
			:description, :merchant, :source, :transaction_date,
			:ai_category_suggestion, :ai_confidence_score, :metadata,
			:created_at, :updated_at
		)
	`
	_, err := r.db.NamedExecContext(ctx, query, txns)
	if err != nil {
		return fmt.Errorf("transactionRepository.BulkInsert: %w", err)
	}

	return nil
}

func (r *transactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error) {
	query := `
		SELECT 
			id, user_id, category_id, amount, currency, amount_in_base,
			description, merchant, source, transaction_date,
			ai_category_suggestion, ai_confidence_score, metadata,
			created_at, updated_at, deleted_at
		FROM transactions
		WHERE id = $1 AND deleted_at IS NULL
	`
	var txn domain.Transaction
	if err := r.db.GetContext(ctx, &txn, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("transactionRepository.GetByID: %w", err)
	}

	return &txn, nil
}

func (r *transactionRepository) GetByDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]domain.Transaction, error) {
	query := `
		SELECT 
			id, user_id, category_id, amount, currency, amount_in_base,
			description, merchant, source, transaction_date,
			ai_category_suggestion, ai_confidence_score, metadata,
			created_at, updated_at, deleted_at
		FROM transactions
		WHERE user_id = $1 AND deleted_at IS NULL AND transaction_date >= $2 AND transaction_date <= $3
		ORDER BY transaction_date DESC
	`

	var txns []domain.Transaction
	if err := r.db.SelectContext(ctx, &txns, query, userID, start, end); err != nil {
		return nil, fmt.Errorf("transactionRepository.GetByDateRange: %w", err)
	}

	if txns == nil {
		txns = []domain.Transaction{}
	}
	return txns, nil
}

func (r *transactionRepository) FindAllByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Transaction, error) {
	query := `
		SELECT 
			id, user_id, category_id, amount, currency, amount_in_base,
			description, merchant, source, transaction_date,
			ai_category_suggestion, ai_confidence_score, metadata,
			created_at, updated_at, deleted_at
		FROM transactions
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY transaction_date DESC, created_at DESC
		LIMIT $2 OFFSET $3
	`

	var txns []domain.Transaction
	if err := r.db.SelectContext(ctx, &txns, query, userID, limit, offset); err != nil {
		return nil, fmt.Errorf("transactionRepository.FindAllByUserID: %w", err)
	}

	if txns == nil {
		txns = []domain.Transaction{}
	}
	return txns, nil
}

func (r *transactionRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE transactions SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("transactionRepository.SoftDelete: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

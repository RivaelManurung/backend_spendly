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

var _ repository.RecurringRepository = (*recurringRepository)(nil)

type recurringRepository struct {
	db *sqlx.DB
}

func NewRecurringRepository(db *sqlx.DB) repository.RecurringRepository {
	return &recurringRepository{db: db}
}

func (r *recurringRepository) Create(ctx context.Context, rec *domain.RecurringTransaction) error {
	now := time.Now()
	if rec.CreatedAt.IsZero() {
		rec.CreatedAt = now
	}
	rec.UpdatedAt = now

	query := `
		INSERT INTO recurring_transactions (
			user_id, account_id, category_id, title, amount, currency, type,
			frequency, start_date, end_date, next_due_date, is_active, auto_post, notes, metadata,
			created_at, updated_at
		) VALUES (
			:user_id, :account_id, :category_id, :title, :amount, :currency, :type,
			:frequency, :start_date, :end_date, :next_due_date, :is_active, :auto_post, :notes, :metadata,
			:created_at, :updated_at
		) RETURNING id
	`
	rows, err := r.db.NamedQueryContext(ctx, query, rec)
	if err != nil {
		return fmt.Errorf("recurringRepository.Create: %w", err)
	}
	defer rows.Close()
	if rows.Next() {
		_ = rows.Scan(&rec.ID)
	}
	return nil
}

func (r *recurringRepository) GetByID(ctx context.Context, id int64) (*domain.RecurringTransaction, error) {
	query := `
		SELECT id, user_id, account_id, category_id, title, amount, currency, type,
		       frequency, start_date, end_date, next_due_date, last_run_at,
		       is_active, auto_post, notes, metadata, created_at, updated_at
		FROM recurring_transactions WHERE id = $1 AND is_active = TRUE
	`
	var rec domain.RecurringTransaction
	if err := r.db.GetContext(ctx, &rec, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("recurringRepository.GetByID: %w", err)
	}
	return &rec, nil
}

func (r *recurringRepository) GetActiveByUser(ctx context.Context, userID uuid.UUID) ([]domain.RecurringTransaction, error) {
	query := `
		SELECT id, user_id, account_id, category_id, title, amount, currency, type,
		       frequency, start_date, end_date, next_due_date, last_run_at,
		       is_active, auto_post, notes, metadata, created_at, updated_at
		FROM recurring_transactions
		WHERE user_id = $1 AND is_active = TRUE
		ORDER BY next_due_date ASC
	`
	var recs []domain.RecurringTransaction
	if err := r.db.SelectContext(ctx, &recs, query, userID); err != nil {
		return nil, fmt.Errorf("recurringRepository.GetActiveByUser: %w", err)
	}
	if recs == nil {
		recs = []domain.RecurringTransaction{}
	}
	return recs, nil
}

func (r *recurringRepository) GetDueBy(ctx context.Context, dueDate time.Time) ([]domain.RecurringTransaction, error) {
	query := `
		SELECT id, user_id, account_id, category_id, title, amount, currency, type,
		       frequency, start_date, end_date, next_due_date, last_run_at,
		       is_active, auto_post, notes, metadata, created_at, updated_at
		FROM recurring_transactions
		WHERE is_active = TRUE AND next_due_date <= $1
		ORDER BY next_due_date ASC
	`
	var recs []domain.RecurringTransaction
	if err := r.db.SelectContext(ctx, &recs, query, dueDate); err != nil {
		return nil, fmt.Errorf("recurringRepository.GetDueBy: %w", err)
	}
	if recs == nil {
		recs = []domain.RecurringTransaction{}
	}
	return recs, nil
}

func (r *recurringRepository) UpdateNextDueDate(ctx context.Context, id int64, nextDue time.Time) error {
	query := `
		UPDATE recurring_transactions
		SET next_due_date = $1, last_run_at = NOW(), updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, nextDue, id)
	if err != nil {
		return fmt.Errorf("recurringRepository.UpdateNextDueDate: %w", err)
	}
	return nil
}

func (r *recurringRepository) SetActive(ctx context.Context, id int64, active bool) error {
	query := `UPDATE recurring_transactions SET is_active = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, active, id)
	if err != nil {
		return fmt.Errorf("recurringRepository.SetActive: %w", err)
	}
	return nil
}

func (r *recurringRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM recurring_transactions WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("recurringRepository.Delete: %w", err)
	}
	return nil
}

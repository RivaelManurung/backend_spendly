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

var _ repository.BudgetRepository = (*budgetRepository)(nil)

type budgetRepository struct {
	db *sqlx.DB
}

// NewBudgetRepository creates a new instance of postgres repo for domain.Budget.
func NewBudgetRepository(db *sqlx.DB) repository.BudgetRepository {
	return &budgetRepository{
		db: db,
	}
}

func (r *budgetRepository) Create(ctx context.Context, budget *domain.Budget) error {
	now := time.Now()
	if budget.CreatedAt.IsZero() {
		budget.CreatedAt = now
	}
	if budget.UpdatedAt.IsZero() {
		budget.UpdatedAt = now
	}

	query := `
		INSERT INTO budgets (
			user_id, category_id, period, limit_amount, currency,
			is_active, created_at, updated_at
		) VALUES (
			:user_id, :category_id, :period, :limit_amount, :currency,
			:is_active, :created_at, :updated_at
		) RETURNING id
	`
	// Since NamedExecContext doesn't support RETURNING into a field, 
	// we use QueryRowxContext with NamedQuery if we need the ID.
	// But let's keep it simple for now or use PrepareNamed.
	
	rows, err := r.db.NamedQueryContext(ctx, query, budget)
	if err != nil {
		return fmt.Errorf("budgetRepository.Create: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&budget.ID); err != nil {
			return fmt.Errorf("budgetRepository.Create scan id: %w", err)
		}
	}

	return nil
}

func (r *budgetRepository) GetByID(ctx context.Context, id int64) (*domain.Budget, error) {
	query := `
		SELECT 
			id, user_id, category_id, period, limit_amount, currency,
			is_active, created_at, updated_at
		FROM budgets
		WHERE id = $1
	`
	var budget domain.Budget
	if err := r.db.GetContext(ctx, &budget, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("budgetRepository.GetByID: %w", err)
	}

	return &budget, nil
}

func (r *budgetRepository) GetActiveBudgetsByUser(ctx context.Context, userID uuid.UUID) ([]domain.Budget, error) {
	query := `
		SELECT 
			id, user_id, category_id, period, limit_amount, currency,
			is_active, created_at, updated_at
		FROM budgets
		WHERE user_id = $1 AND is_active = true
	`
	var budgets []domain.Budget
	if err := r.db.SelectContext(ctx, &budgets, query, userID); err != nil {
		return nil, fmt.Errorf("budgetRepository.GetActiveBudgetsByUser: %w", err)
	}

	if budgets == nil {
		budgets = []domain.Budget{}
	}
	return budgets, nil
}

func (r *budgetRepository) GetSpentAmount(ctx context.Context, budgetID int64) (float64, error) {
	// Query the v_budget_status view or calculate explicitly.
	// The user mentioned querying v_budget_status view.
	query := `SELECT spent_amount FROM v_budget_status WHERE budget_id = $1`
	var spentAmount float64
	err := r.db.GetContext(ctx, &spentAmount, query, budgetID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("budgetRepository.GetSpentAmount: %w", err)
	}
	return spentAmount, nil
}

func (r *budgetRepository) Update(ctx context.Context, budget *domain.Budget) error {
	budget.UpdatedAt = time.Now()
	query := `
		UPDATE budgets
		SET limit_amount = :limit_amount,
			is_active = :is_active,
			updated_at = :updated_at
		WHERE id = :id
	`
	_, err := r.db.NamedExecContext(ctx, query, budget)
	if err != nil {
		return fmt.Errorf("budgetRepository.Update: %w", err)
	}
	return nil
}

func (r *budgetRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM budgets WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("budgetRepository.Delete: %w", err)
	}
	return nil
}

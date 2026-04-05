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

var _ repository.AnalysisRepository = (*analysisRepository)(nil)

type analysisRepository struct {
	db *sqlx.DB
}

// NewAnalysisRepository creates a new instance of postgres repo for analysis snapshots.
func NewAnalysisRepository(db *sqlx.DB) repository.AnalysisRepository {
	return &analysisRepository{
		db: db,
	}
}

func (r *analysisRepository) UpsertSnapshot(ctx context.Context, snapshot *domain.AnalysisSnapshot) error {
	now := time.Now()
	if snapshot.CreatedAt.IsZero() {
		snapshot.CreatedAt = now
	}
	snapshot.UpdatedAt = now

	query := `
		INSERT INTO analysis_snapshots (
			user_id, period_type, period_value, total_income, total_expense,
			net_savings, top_categories, category_breakdown, 
			merchant_breakdown, daily_trend, forecast_end_balance, forecast_confidence,
			status, metadata, created_at, updated_at
		) VALUES (
			:user_id, :period_type, :period_value, :total_income, :total_expense,
			:net_savings, :top_categories, :category_breakdown,
			:merchant_breakdown, :daily_trend, :forecast_end_balance, :forecast_confidence,
			:status, :metadata, :created_at, :updated_at
		)
		ON CONFLICT (user_id, period_type, period_value) DO UPDATE SET
			total_income = EXCLUDED.total_income,
			total_expense = EXCLUDED.total_expense,
			net_savings = EXCLUDED.net_savings,
			top_categories = EXCLUDED.top_categories,
			category_breakdown = EXCLUDED.category_breakdown,
			merchant_breakdown = EXCLUDED.merchant_breakdown,
			daily_trend = EXCLUDED.daily_trend,
			forecast_end_balance = EXCLUDED.forecast_end_balance,
			forecast_confidence = EXCLUDED.forecast_confidence,
			status = EXCLUDED.status,
			metadata = EXCLUDED.metadata,
			updated_at = EXCLUDED.updated_at
		RETURNING id
	`

	rows, err := r.db.NamedQueryContext(ctx, query, snapshot)
	if err != nil {
		return fmt.Errorf("analysisRepository.UpsertSnapshot: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&snapshot.ID); err != nil {
			return fmt.Errorf("analysisRepository.UpsertSnapshot scan id: %w", err)
		}
	}

	return nil
}

func (r *analysisRepository) GetByPeriod(ctx context.Context, userID uuid.UUID, period string) (*domain.AnalysisSnapshot, error) {
	query := `
		SELECT 
			id, user_id, period_type, period_value, total_income, total_expense,
			net_savings, top_categories, category_breakdown, 
			merchant_breakdown, daily_trend, forecast_end_balance, forecast_confidence,
			status, metadata, created_at, updated_at
		FROM analysis_snapshots
		WHERE user_id = $1 AND period_value = $2
	`
	var snapshot domain.AnalysisSnapshot
	if err := r.db.GetContext(ctx, &snapshot, query, userID, period); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("analysisRepository.GetByPeriod: %w", err)
	}

	return &snapshot, nil
}

func (r *analysisRepository) GetLatestByUserID(ctx context.Context, userID uuid.UUID, periodType string) (*domain.AnalysisSnapshot, error) {
	query := `
		SELECT 
			id, user_id, period_type, period_value, total_income, total_expense,
			net_savings, top_categories, category_breakdown, 
			merchant_breakdown, daily_trend, status, metadata,
			created_at, updated_at
		FROM analysis_snapshots
		WHERE user_id = $1 AND period_type = $2 AND status = 'COMPLETED'
		ORDER BY period_value DESC
		LIMIT 1
	`
	var snapshot domain.AnalysisSnapshot
	if err := r.db.GetContext(ctx, &snapshot, query, userID, periodType); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("analysisRepository.GetLatestByUserID: %w", err)
	}

	return &snapshot, nil
}

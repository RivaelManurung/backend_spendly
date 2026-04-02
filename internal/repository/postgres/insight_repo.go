package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/spendly/backend/internal/domain"
	"github.com/spendly/backend/internal/repository"
)

var _ repository.InsightRepository = (*insightRepository)(nil)

type insightRepository struct {
	db *sqlx.DB
}

// NewInsightRepository creates a new instance of postgres repo for AI insights.
func NewInsightRepository(db *sqlx.DB) repository.InsightRepository {
	return &insightRepository{
		db: db,
	}
}

func (r *insightRepository) SaveInsight(ctx context.Context, insight *domain.AIInsight) error {
	now := time.Now()
	if insight.CreatedAt.IsZero() {
		insight.CreatedAt = now
	}
	insight.UpdatedAt = now

	query := `
		INSERT INTO ai_insights (
			user_id, snapshot_id, type, title, content, 
			priority, is_read, metadata, created_at, updated_at
		) VALUES (
			:user_id, :snapshot_id, :type, :title, :content,
			:priority, :is_read, :metadata, :created_at, :updated_at
		) RETURNING id
	`

	rows, err := r.db.NamedQueryContext(ctx, query, insight)
	if err != nil {
		return fmt.Errorf("insightRepository.SaveInsight: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&insight.ID); err != nil {
			return fmt.Errorf("insightRepository.SaveInsight scan id: %w", err)
		}
	}

	return nil
}

func (r *insightRepository) SaveBatch(ctx context.Context, insights []domain.AIInsight) error {
	if len(insights) == 0 {
		return nil
	}
	// Simplified batch insert using NamedExecContext if sqlx supports slice.
	// Note: NamedExec with a slice of structs works in recent sqlx versions.
	query := `
		INSERT INTO ai_insights (
			user_id, snapshot_id, type, title, content, 
			priority, is_read, metadata, created_at, updated_at
		) VALUES (
			:user_id, :snapshot_id, :type, :title, :content,
			:priority, :is_read, :metadata, :created_at, :updated_at
		)
	`
	_, err := r.db.NamedExecContext(ctx, query, insights)
	if err != nil {
		return fmt.Errorf("insightRepository.SaveBatch: %w", err)
	}
	return nil
}

func (r *insightRepository) ListUnread(ctx context.Context, userID uuid.UUID) ([]domain.AIInsight, error) {
	query := `
		SELECT 
			id, user_id, snapshot_id, type, title, content, 
			priority, is_read, metadata, created_at, updated_at
		FROM ai_insights
		WHERE user_id = $1 AND is_read = false
		ORDER BY priority DESC, created_at DESC
	`
	var insights []domain.AIInsight
	if err := r.db.SelectContext(ctx, &insights, query, userID); err != nil {
		return nil, fmt.Errorf("insightRepository.ListUnread: %w", err)
	}

	if insights == nil {
		insights = []domain.AIInsight{}
	}
	return insights, nil
}

func (r *insightRepository) MarkRead(ctx context.Context, id int64) error {
	query := `UPDATE ai_insights SET is_read = true, updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("insightRepository.MarkRead: %w", err)
	}
	return nil
}

func (r *insightRepository) SetHelpful(ctx context.Context, id int64, helpful bool) error {
	query := `
		UPDATE ai_insights 
		SET metadata = jsonb_set(COALESCE(metadata, '{}'::jsonb), '{helpful}', $1::jsonb), 
			updated_at = NOW() 
		WHERE id = $2
	`
	helpfulStr := "false"
	if helpful {
		helpfulStr = "true"
	}
	_, err := r.db.ExecContext(ctx, query, helpfulStr, id)
	if err != nil {
		return fmt.Errorf("insightRepository.SetHelpful: %w", err)
	}
	return nil
}

func (r *insightRepository) GetLatestByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]domain.AIInsight, error) {
	query := `
		SELECT 
			id, user_id, snapshot_id, type, title, content, 
			priority, is_read, metadata, created_at, updated_at
		FROM ai_insights
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	var insights []domain.AIInsight
	if err := r.db.SelectContext(ctx, &insights, query, userID, limit); err != nil {
		return nil, fmt.Errorf("insightRepository.GetLatestByUserID: %w", err)
	}

	if insights == nil {
		insights = []domain.AIInsight{}
	}
	return insights, nil
}

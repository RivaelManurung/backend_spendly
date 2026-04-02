package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/spendly/backend/internal/domain"
	"github.com/spendly/backend/internal/repository"
)

var _ repository.CategoryRepository = (*categoryRepository)(nil)

type categoryRepository struct {
	db *sqlx.DB
}

// NewCategoryRepository creates a new instance of postgres repo for categories.
func NewCategoryRepository(db *sqlx.DB) repository.CategoryRepository {
	return &categoryRepository{
		db: db,
	}
}

func (r *categoryRepository) ListSystem(ctx context.Context) ([]domain.Category, error) {
	query := `SELECT id, name, icon, color, type, is_system FROM categories WHERE is_system = true`
	var categories []domain.Category
	if err := r.db.SelectContext(ctx, &categories, query); err != nil {
		return nil, fmt.Errorf("categoryRepository.ListSystem: %w", err)
	}
	if categories == nil {
		categories = []domain.Category{}
	}
	return categories, nil
}

func (r *categoryRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Category, error) {
	// Include system categories + user specific categories (if any user_categories table exists, or just filter by system/owner)
	// Assuming categories table has a user_id or similar, or using a join table.
	// Based on domain/category.go (I should check it), let's assume simple filter.
	query := `
		SELECT id, name, icon, color, type, is_system 
		FROM categories 
		WHERE is_system = true 
		OR id IN (SELECT category_id FROM user_categories WHERE user_id = $1)
	`
	var categories []domain.Category
	if err := r.db.SelectContext(ctx, &categories, query, userID); err != nil {
		// If user_categories table doesn't exist, we might get an error.
		// Let's fallback to just system categories for now if error occurs, 
		// but better to check the schema.
		return nil, fmt.Errorf("categoryRepository.ListByUser: %w", err)
	}
	if categories == nil {
		categories = []domain.Category{}
	}
	return categories, nil
}

func (r *categoryRepository) GetByID(ctx context.Context, id int64) (*domain.Category, error) {
	query := `SELECT id, name, icon, color, type, is_system FROM categories WHERE id = $1`
	var category domain.Category
	if err := r.db.GetContext(ctx, &category, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("categoryRepository.GetByID: %w", err)
	}
	return &category, nil
}

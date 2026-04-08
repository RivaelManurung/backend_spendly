package repository

import (
	"context"

	"github.com/spendly/backend/internal/domain"
	"gorm.io/gorm"
)

type BudgetRepository interface {
	Create(ctx context.Context, budget *domain.Budget) error
	GetAll(ctx context.Context) ([]domain.Budget, error)
	GetActiveByCategoryID(ctx context.Context, categoryID string) (*domain.Budget, error)
}

type budgetRepository struct {
	db *gorm.DB
}

func NewBudgetRepository(db *gorm.DB) BudgetRepository {
	return &budgetRepository{db: db}
}

func (r *budgetRepository) Create(ctx context.Context, budget *domain.Budget) error {
	return r.db.WithContext(ctx).Create(budget).Error
}

func (r *budgetRepository) GetAll(ctx context.Context) ([]domain.Budget, error) {
	var budgets []domain.Budget
	// Preload category detail for UI viewing
	err := r.db.WithContext(ctx).Preload("Category").Find(&budgets).Error
	return budgets, err
}

func (r *budgetRepository) GetActiveByCategoryID(ctx context.Context, categoryID string) (*domain.Budget, error) {
	var budget domain.Budget
	// Note: Currently fetching the latest budget for this category
	err := r.db.WithContext(ctx).
		Where("category_id = ?", categoryID).
		Order("start_date desc").First(&budget).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No active budget found
		}
		return nil, err
	}
	return &budget, nil
}

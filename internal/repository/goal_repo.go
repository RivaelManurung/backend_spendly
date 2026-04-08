package repository

import (
	"context"

	"github.com/spendly/backend/internal/domain"
	"gorm.io/gorm"
)

type GoalRepository interface {
	Create(ctx context.Context, goal *domain.Goal) error
	GetAll(ctx context.Context) ([]domain.Goal, error)
	FindByID(ctx context.Context, goalID string) (*domain.Goal, error)
	Update(ctx context.Context, goal *domain.Goal) error
	AddContribution(ctx context.Context, contribution *domain.GoalContribution, tx *domain.Transaction) error
}

type goalRepository struct {
	db *gorm.DB
}

func NewGoalRepository(db *gorm.DB) GoalRepository {
	return &goalRepository{db: db}
}

func (r *goalRepository) Create(ctx context.Context, goal *domain.Goal) error {
	return r.db.WithContext(ctx).Create(goal).Error
}

func (r *goalRepository) GetAll(ctx context.Context) ([]domain.Goal, error) {
	var goals []domain.Goal
	err := r.db.WithContext(ctx).Order("created_at desc").Find(&goals).Error
	return goals, err
}

func (r *goalRepository) FindByID(ctx context.Context, goalID string) (*domain.Goal, error) {
	var goal domain.Goal
	err := r.db.WithContext(ctx).First(&goal, "id = ?", goalID).Error
	return &goal, err
}

func (r *goalRepository) Update(ctx context.Context, goal *domain.Goal) error {
	return r.db.WithContext(ctx).Save(goal).Error
}

func (r *goalRepository) AddContribution(ctx context.Context, contribution *domain.GoalContribution, tx *domain.Transaction) error {
	// Atomic transaction: create goal contribution, create related transaction, update goal current amount
	return r.db.WithContext(ctx).Transaction(func(dbTx *gorm.DB) error {
		// Create transaction first
		if err := dbTx.Create(tx).Error; err != nil {
			return err
		}

		// Set transaction ID to contribution
		contribution.TransactionID = tx.ID
		if err := dbTx.Create(contribution).Error; err != nil {
			return err
		}

		// Update goal amount
		if err := dbTx.Model(&domain.Goal{}).
			Where("id = ?", contribution.GoalID).
			UpdateColumn("current_amount", gorm.Expr("current_amount + ?", contribution.Amount)).Error; err != nil {
			return err
		}

		return nil
	})
}

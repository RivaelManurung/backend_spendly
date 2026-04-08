package repository

import (
	"context"
	"github.com/spendly/backend/internal/domain"
	"gorm.io/gorm"
)

type CategoryRepository interface {
	Create(ctx context.Context, cat *domain.Category) error
	GetAll(ctx context.Context) ([]domain.Category, error)
	Upsert(ctx context.Context, cat *domain.Category) error
}

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(ctx context.Context, cat *domain.Category) error {
	return r.db.WithContext(ctx).Create(cat).Error
}

func (r *categoryRepository) GetAll(ctx context.Context) ([]domain.Category, error) {
	var cats []domain.Category
	err := r.db.WithContext(ctx).Find(&cats).Error
	return cats, err
}

func (r *categoryRepository) Upsert(ctx context.Context, cat *domain.Category) error {
	return r.db.WithContext(ctx).Save(cat).Error
}

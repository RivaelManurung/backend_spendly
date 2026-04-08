package service

import (
	"context"
	"github.com/spendly/backend/internal/domain"
	"github.com/spendly/backend/internal/repository"
)

type CategoryService interface {
	SyncCategory(ctx context.Context, id, label, icon, color, catType string) (*domain.Category, error)
	GetAllCategories(ctx context.Context) ([]domain.Category, error)
}

type categoryService struct {
	repo repository.CategoryRepository
}

func NewCategoryService(repo repository.CategoryRepository) CategoryService {
	return &categoryService{repo: repo}
}

func (s *categoryService) SyncCategory(ctx context.Context, id, label, icon, color, catType string) (*domain.Category, error) {
	cat := &domain.Category{
		Base:  domain.Base{ID: id},
		Label: label,
		Icon:  icon,
		Color: color,
		Type:  catType,
	}
	err := s.repo.Upsert(ctx, cat)
	return cat, err
}

func (s *categoryService) GetAllCategories(ctx context.Context) ([]domain.Category, error) {
	return s.repo.GetAll(ctx)
}

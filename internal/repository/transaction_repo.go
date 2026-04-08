package repository

import (
	"context"
	"time"

	"github.com/spendly/backend/internal/domain"
	"gorm.io/gorm"
)

type TransactionRepository interface {
	Create(ctx context.Context, tx *domain.Transaction) error
	GetAll(ctx context.Context) ([]domain.Transaction, error)
	GetUpdatedSince(ctx context.Context, since time.Time) ([]domain.Transaction, error)
	GetDeletedSince(ctx context.Context, since time.Time) ([]string, error)
}

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(ctx context.Context, tx *domain.Transaction) error {
	// Use Save instead of Create to handle Upsert/Update if ID already exists
	return r.db.WithContext(ctx).Save(tx).Error
}

func (r *transactionRepository) GetAll(ctx context.Context) ([]domain.Transaction, error) {
	var txs []domain.Transaction
	// Preload Category so we get category details
	err := r.db.WithContext(ctx).Preload("Category").Order("date desc").Find(&txs).Error
	return txs, err
}

func (r *transactionRepository) GetUpdatedSince(ctx context.Context, since time.Time) ([]domain.Transaction, error) {
	var txs []domain.Transaction
	// Find all transactions either created or updated after 'since'
	err := r.db.WithContext(ctx).Preload("Category").
		Where("updated_at >= ?", since).
		Find(&txs).Error
	return txs, err
}

func (r *transactionRepository) GetDeletedSince(ctx context.Context, since time.Time) ([]string, error) {
	var txs []domain.Transaction
	// Unscoped allows fetching soft-deleted records!
	err := r.db.WithContext(ctx).Unscoped().
		Where("deleted_at >= ?", since).
		Find(&txs).Error

	var ids []string
	for _, tx := range txs {
		ids = append(ids, tx.ID)
	}
	return ids, err
}

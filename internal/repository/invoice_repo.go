package repository

import (
	"context"

	"github.com/spendly/backend/internal/domain"
	"gorm.io/gorm"
)

type InvoiceRepository interface {
	Create(ctx context.Context, invoice *domain.Invoice) error
	GetAll(ctx context.Context) ([]domain.Invoice, error)
	UpdateStatus(ctx context.Context, id string, status string) error
}

type invoiceRepository struct {
	db *gorm.DB
}

func NewInvoiceRepository(db *gorm.DB) InvoiceRepository {
	return &invoiceRepository{db: db}
}

func (r *invoiceRepository) Create(ctx context.Context, invoice *domain.Invoice) error {
	return r.db.WithContext(ctx).Create(invoice).Error
}

func (r *invoiceRepository) GetAll(ctx context.Context) ([]domain.Invoice, error) {
	var invoices []domain.Invoice
	err := r.db.WithContext(ctx).Order("created_at desc").Find(&invoices).Error
	return invoices, err
}

func (r *invoiceRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	return r.db.WithContext(ctx).Model(&domain.Invoice{}).
		Where("id = ?", id).
		Update("status", status).Error
}

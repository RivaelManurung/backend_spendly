package service

import (
	"context"
	"errors"
	"time"

	"github.com/spendly/backend/internal/domain"
	"github.com/spendly/backend/internal/repository"
)

type TransactionService interface {
	SyncTransaction(ctx context.Context, id string, title string, amount float64, catID string, txType string, note string, isRecurring bool, date time.Time, deviceID string) (*domain.Transaction, error)
	CreateTransaction(ctx context.Context, title string, amount float64, catID string, txType string, note string, isRecurring bool) (*domain.Transaction, error)
	GetAllTransactions(ctx context.Context) ([]domain.Transaction, error)
}

type transactionService struct {
	txRepo repository.TransactionRepository
}

func NewTransactionService(txRepo repository.TransactionRepository) TransactionService {
	return &transactionService{
		txRepo: txRepo,
	}
}

func (s *transactionService) SyncTransaction(ctx context.Context, id string, title string, amount float64, catID string, txType string, note string, isRecurring bool, date time.Time, deviceID string) (*domain.Transaction, error) {
	tx := &domain.Transaction{
		Base: domain.Base{
			ID: id,
		},
		Title:       title,
		Amount:      amount,
		CategoryID:  catID,
		Type:        txType,
		Note:        note,
		Date:        date,
		IsRecurring: isRecurring,
		DeviceID:    deviceID,
	}

	// Try to find if it exists first to avoid duplicates during sync
	// But for now, we just create. Repository might need to handle Upsert.
	if err := s.txRepo.Create(ctx, tx); err != nil {
		return nil, err
	}

	return tx, nil
}

func (s *transactionService) CreateTransaction(ctx context.Context, title string, amount float64, catID string, txType string, note string, isRecurring bool) (*domain.Transaction, error) {
	if txType != "income" && txType != "expense" && txType != "goal" {
		return nil, errors.New("invalid transaction type")
	}

	tx := &domain.Transaction{
		Title:       title,
		Amount:      amount,
		CategoryID:  catID,
		Type:        txType,
		Note:        note,
		Date:        time.Now(),
		IsRecurring: isRecurring,
	}

	if err := s.txRepo.Create(ctx, tx); err != nil {
		return nil, err
	}

	return tx, nil
}

func (s *transactionService) GetAllTransactions(ctx context.Context) ([]domain.Transaction, error) {
	return s.txRepo.GetAll(ctx)
}

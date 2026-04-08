package service

import (
	"context"
	"time"

	"github.com/spendly/backend/internal/domain"
	"github.com/spendly/backend/internal/repository"
)

type SyncResponse struct {
	Updated []domain.Transaction `json:"updated"`
	Deleted []string             `json:"deleted"` // IDs of soft-deleted records to wipe from mobile DB
	SyncAt  time.Time            `json:"sync_at"`
}

type SyncService interface {
	SyncTransactions(ctx context.Context, lastSyncTime time.Time) (*SyncResponse, error)
}

type syncService struct {
	txRepo repository.TransactionRepository
}

func NewSyncService(txRepo repository.TransactionRepository) SyncService {
	return &syncService{txRepo: txRepo}
}

func (s *syncService) SyncTransactions(ctx context.Context, lastSyncTime time.Time) (*SyncResponse, error) {
	updated, err := s.txRepo.GetUpdatedSince(ctx, lastSyncTime)
	if err != nil {
		return nil, err
	}

	deleted, err := s.txRepo.GetDeletedSince(ctx, lastSyncTime)
	if err != nil {
		return nil, err
	}

	return &SyncResponse{
		Updated: updated,
		Deleted: deleted,
		SyncAt:  time.Now(),
	}, nil
}

package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/spendly/backend/internal/domain"
	"github.com/spendly/backend/internal/repository"
)

// TransactionService mendefinisikan contract untuk interaksi dengan Transaksi.
// Termasuk auto-kategorisasi AI dan notifikasi Threshold Budget.
type TransactionService interface {
	CreateTransaction(ctx context.Context, txn *domain.Transaction, validateBudget bool) (*domain.Transaction, error)
	GetTransactionHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Transaction, error)
}

type transactionService struct {
	repo         repository.TransactionRepository
	aiCategorize *AiCategorizationService
	// Di sistem rill, kita mungkin butuh:
	// budgetSvc BudgetAlertService
	// notifierSvc NotificationService
}

func NewTransactionService(repo repository.TransactionRepository, aiCat *AiCategorizationService) TransactionService {
	return &transactionService{
		repo:         repo,
		aiCategorize: aiCat,
	}
}

// CreateTransaction mewakili pipeline `transaction_input_pipeline` di README:
// 1. Input/Validate
// 2. Currency Normalize
// 3. Auto Categorize (LLM)
// 4. Save + Threshold Emit Event
func (s *transactionService) CreateTransaction(ctx context.Context, txn *domain.Transaction, validateBudget bool) (*domain.Transaction, error) {
	if txn.UserID == uuid.Nil {
		return nil, errors.New("user_id cannot be empty")
	}

	// Step 1 & 2: Normalisasi dan Validasi input
	if txn.AmountInBase.IsZero() {
		// Mock logic: misal base currency user IDR.
		// Bila transaksi USD, kita kali rate (di skip sementara, disamakan dengan input).
		txn.AmountInBase = txn.Amount
	}
	
	if txn.TransactionDate.IsZero() {
		txn.TransactionDate = time.Now()
	}

	// Step 3: Trigger LLM Kategori jika ID kategori belum ditembak secara manual oleh user
	if txn.CategoryID == nil && s.aiCategorize != nil {
		// Kita butuh daftar category yang live, anggap kita mock atau sediakan di cache.
		// Dalam real project: categories := s.catRepo.FindAll(...)
		mockAvailableCategories := []domain.Category{
			{ID: 1, Name: "Makanan & Minuman", Type: "expense"},
			{ID: 2, Name: "Transportasi", Type: "expense"},
			{ID: 3, Name: "Belanja", Type: "expense"},
			{ID: 4, Name: "Hiburan / Langganan", Type: "expense"},
		}

		err := s.aiCategorize.AutoCategorize(ctx, txn, mockAvailableCategories)
		if err != nil {
			// Kita tak gagalkan transaksi walau AI gagal
			fmt.Printf("Warning: AI Categorization failed: %v\n", err)
		}
	}

	txn.ID = uuid.New()
	
	// Step 4: Simpan ke PostgreSQL Repo
	if err := s.repo.Insert(ctx, txn); err != nil {
		return nil, fmt.Errorf("failed creating transaction: %w", err)
	}

	// Step 5: (Optional Pipeline) - Pengecekan Limit Budget Event
	if validateBudget {
		// Sesuai README `.agent`, setelah save -> check budget -> emit event.
		// err := s.budgetSvc.CheckThresholds(ctx, txn) ...
		// s.eventEmitter.Emit("transaction.created", txn)
	}

	return txn, nil
}

func (s *transactionService) GetTransactionHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Transaction, error) {
	if limit <= 0 {
		limit = 10 // Paging default
	}

	txns, err := s.repo.FindAllByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user transactions: %w", err)
	}

	return txns, nil
}

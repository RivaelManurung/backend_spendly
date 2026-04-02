package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/spendly/backend/internal/domain"
	// "github.com/spendly/backend/internal/repository"
)

// BudgetService menyediakan interface logic khusus Manajemen Anggaran Bulanan
type BudgetService interface {
	CreateBudget(ctx context.Context, userID uuid.UUID, catID *int64, limit decimal.Decimal, pType string) (*domain.Budget, error)
	// Kita bisa juga menambahkan GetBudgets, GetBudgetRemaining
	CheckBudgetThreshold(ctx context.Context, txn *domain.Transaction) error
}

type budgetService struct {
	// budgetRepo repository.BudgetRepository
	budgetAlertSvc *BudgetPipeline
}

func NewBudgetService(alertSvc *BudgetPipeline) BudgetService {
	return &budgetService{
		// budgetRepo: repo,
		budgetAlertSvc: alertSvc,
	}
}

func (b *budgetService) CreateBudget(ctx context.Context, userID uuid.UUID, catID *int64, limit decimal.Decimal, pType string) (*domain.Budget, error) {
	if pType != "monthly" && pType != "weekly" && pType != "yearly" {
		return nil, errors.New("invalid period type")
	}

	budget := &domain.Budget{
		UserID:      userID,
		CategoryID:  catID,
		LimitAmount: limit,
		Currency:    "IDR", // Mock base currency
		Period:      pType,
		IsActive:    true,
	}

	now := time.Now()
	budget.CreatedAt = now
	budget.UpdatedAt = now

	// Panggil repo create disini:
	// err := b.budgetRepo.Create(ctx, budget)

	return budget, nil
}

// CheckBudgetThreshold dipanggil lewat emitter ketika Transaksi Baru Dibuat.
func (b *budgetService) CheckBudgetThreshold(ctx context.Context, txn *domain.Transaction) error {
	// 1. Ambil Data Budget di PostgreSQL sesuai Category ID milik transaksi
	// budget, err := b.budgetRepo.GetActiveBudget(ctx, txn.UserID, txn.CategoryID)
	
	// Mock implementasi:
	limit := decimal.NewFromFloat(5000000.0) // Budget 5jt
	spent := decimal.NewFromFloat(4200000.0) // Sudah terpakai 4.2jt
	
	currentSpent := spent.Add(txn.AmountInBase)
	
	// 2. Hitung persentase threshold
	ratio := currentSpent.Div(limit).InexactFloat64()
	
	// 3. Emit / panggil Gemini Agent Notification (`budget_alert_pipeline` dari Antigravity) di 80% & 100%
	if ratio >= 0.80 {
		
		mockModel := domain.Budget{LimitAmount: limit, CategoryID: txn.CategoryID}
		
		// Run pipeline .agent (budget_alert.prompt)
		msgAI, promptErr := b.budgetAlertSvc.GenerateAlert(ctx, txn.UserID, mockModel, currentSpent.String())
		if promptErr != nil {
			fmt.Printf("Warning: Failed generating AI Notification: %v\n", promptErr)
			return promptErr
		}

		// Kirim Push Notification dari hasil Text Gemini
		fmt.Printf("[BUDGET ALERT -> %s] : %s\n", txn.UserID.String(), msgAI)
	}

	return nil
}

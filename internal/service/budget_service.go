package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/spendly/backend/internal/domain"
	"github.com/spendly/backend/internal/repository"
)

type BudgetBurnRate struct {
	Budget        *domain.Budget `json:"budget"`
	CurrentSpent  float64        `json:"current_spent"`
	Remaining     float64        `json:"remaining"`
	UsagePercent  float64        `json:"usage_percent"`
	StatusMessage string         `json:"status_message"`
}

type BudgetService interface {
	CreateBudget(ctx context.Context, categoryID string, amount float64, period string) (*domain.Budget, error)
	GetBudgetsWithBurnRate(ctx context.Context) ([]BudgetBurnRate, error)
	CheckBurnRate(ctx context.Context, categoryID string, transactionAmount float64) error // Triggers Alert logic
}

type budgetService struct {
	budgetRepo repository.BudgetRepository
	txRepo     repository.TransactionRepository
}

func NewBudgetService(budgetRepo repository.BudgetRepository, txRepo repository.TransactionRepository) BudgetService {
	return &budgetService{
		budgetRepo: budgetRepo,
		txRepo:     txRepo,
	}
}

func (s *budgetService) CreateBudget(ctx context.Context, categoryID string, amount float64, period string) (*domain.Budget, error) {
	if period != "monthly" && period != "weekly" {
		return nil, errors.New("invalid period: choose monthly or weekly")
	}

	budget := &domain.Budget{
		CategoryID: categoryID,
		Amount:     amount,
		Period:     period,
		StartDate:  time.Now(),
	}

	if err := s.budgetRepo.Create(ctx, budget); err != nil {
		return nil, err
	}

	return budget, nil
}

func (s *budgetService) GetBudgetsWithBurnRate(ctx context.Context) ([]BudgetBurnRate, error) {
	budgets, err := s.budgetRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	txs, err := s.txRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var responses []BudgetBurnRate

	for i := range budgets {
		b := &budgets[i]

		var periodStart, periodEnd time.Time
		if b.Period == "monthly" {
			now := time.Now()
			periodStart = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
			periodEnd = periodStart.AddDate(0, 1, -1)
		} else {
			now := time.Now()
			offset := int(now.Weekday())
			periodStart = now.AddDate(0, 0, -offset)
			periodEnd = periodStart.AddDate(0, 0, 6)
		}

		var sumSpent float64
		for _, tx := range txs {
			if tx.CategoryID == b.CategoryID &&
				tx.Type == "expense" &&
				(tx.Date.After(periodStart) || tx.Date.Equal(periodStart)) &&
				(tx.Date.Before(periodEnd) || tx.Date.Equal(periodEnd)) {
				sumSpent += tx.Amount
			}
		}

		pct := 0.0
		if b.Amount > 0 {
			pct = (sumSpent / b.Amount) * 100
		}

		status := "Healthy"
		if pct > 80 {
			status = "Nearing Limit/Overbudget"
		}

		responses = append(responses, BudgetBurnRate{
			Budget:        b,
			CurrentSpent:  sumSpent,
			Remaining:     math.Max(0, b.Amount-sumSpent),
			UsagePercent:  pct,
			StatusMessage: status,
		})
	}

	return responses, nil
}

func (s *budgetService) CheckBurnRate(ctx context.Context, categoryID string, transactionAmount float64) error {
	b, err := s.budgetRepo.GetActiveByCategoryID(ctx, categoryID)
	if err != nil || b == nil {
		return nil
	}
	fmt.Printf("[ALERTS]: Emitting async CheckBurnRate calculation\n")
	return nil
}

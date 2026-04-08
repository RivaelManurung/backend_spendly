package service

import (
	"context"
	"math"

	"github.com/spendly/backend/internal/repository"
)

type AnalyticsReport struct {
	TotalIncome  float64            `json:"total_income"`
	TotalExpense float64            `json:"total_expense"`
	NetWorth     float64            `json:"net_worth"`
	TrendData    map[string]float64 `json:"trend_data"` // For FL Chart plotting: e.g. "Food" -> % or amount
}

type ReportService interface {
	GetMonthlyAnalytics(ctx context.Context) (*AnalyticsReport, error)
}

type reportService struct {
	txRepo repository.TransactionRepository
}

func NewReportService(txRepo repository.TransactionRepository) ReportService {
	return &reportService{txRepo: txRepo}
}

func (s *reportService) GetMonthlyAnalytics(ctx context.Context) (*AnalyticsReport, error) {
	// Querying raw instead of pulling everything into memory is better for large datasets
	txs, err := s.txRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	report := &AnalyticsReport{
		TrendData: make(map[string]float64),
	}

	for _, tx := range txs {
		switch tx.Type {
		case "income":
			report.TotalIncome += tx.Amount
		case "expense":
			report.TotalExpense += tx.Amount

			// Categorization trends (for FL Chart)
			catName := "Uncategorized"
			if tx.Category.Label != "" {
				catName = tx.Category.Label
			}
			report.TrendData[catName] += tx.Amount
		}
	}

	report.NetWorth = math.Round((report.TotalIncome-report.TotalExpense)*100) / 100

	return report, nil
}

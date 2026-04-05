package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/spendly/backend/internal/ai"
	"github.com/spendly/backend/internal/domain"
	"github.com/spendly/backend/internal/repository"
)

type AnalysisPipeline struct {
	gemini       *ai.GeminiClient
	repoTxn      repository.TransactionRepository
	repoSnap     repository.AnalysisRepository
	repoInsight  repository.InsightRepository
	repoUser     repository.UserRepository
}

func NewAnalysisPipeline(g *ai.GeminiClient, repoTxn repository.TransactionRepository, repoSnap repository.AnalysisRepository, repoInsight repository.InsightRepository, repoUser repository.UserRepository) *AnalysisPipeline {
	return &AnalysisPipeline{
		gemini:      g,
		repoTxn:     repoTxn,
		repoSnap:    repoSnap,
		repoInsight: repoInsight,
		repoUser:    repoUser,
	}
}

// RunBackgroundMonthlyJobs runs the full analysis pipeline for a user.
func (p *AnalysisPipeline) RunBackgroundMonthlyJobs(ctx context.Context, userID uuid.UUID, period string) error {
	// 1. Fetch transactions for the period
	// Assume we calculate start/end date from period "2025-01"
	start, _ := time.Parse("2006-01", period)
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)

	txns, err := p.repoTxn.GetByDateRange(ctx, userID, start, end)
	if err != nil {
		return fmt.Errorf("failed to fetch transactions: %w", err)
	}

	// 2. Build local snapshot
	snap := p.buildSnapshot(userID, start, end, txns)
	snap.PeriodValue = period
	snap.Status = "COMPLETED"

	// 2.1 Calculate Forecast
	user, _ := p.repoUser.GetByID(ctx, userID)
	if user != nil {
		p.enrichWithForecast(ctx, snap, user, txns)
	}

	// 3. Persist snapshot
	if err := p.repoSnap.UpsertSnapshot(ctx, snap); err != nil {
		return fmt.Errorf("failed to upsert snapshot: %w", err)
	}

	// 4. Fetch previous snapshot for comparison
	prevPeriod := start.AddDate(0, -1, 0).Format("2006-01")
	prevSnap, _ := p.repoSnap.GetByPeriod(ctx, userID, prevPeriod)

	// 5. Run 3 AI agents in parallel
	snapBytes, _ := json.Marshal(snap)
	prevSnapBytes := []byte("{}")
	if prevSnap != nil {
		prevSnapBytes, _ = json.Marshal(prevSnap)
	}

	var wg sync.WaitGroup
	insightsChan := make(chan domain.AIInsight, 3)

	type agentJob struct {
		name string
		vars map[string]string
		type_ string
	}

	jobs := []agentJob{
		{
			name: "monthly_analysis",
			type_: "MONTHLY_SUMMARY",
			vars: map[string]string{"snapshot_data": string(snapBytes)},
		},
		{
			name: "anomaly_detection",
			type_: "ANOMALY",
			vars: map[string]string{
				"current_month": string(snapBytes),
				"prev_month":    string(prevSnapBytes),
				"subscriptions": string(p.detectSubscriptions(txns)),
			},
		},
		{
			name: "saving_opportunities",
			type_: "SAVING_TIP",
			vars: map[string]string{"snapshot_data": string(snapBytes)},
		},
	}

	for _, job := range jobs {
		wg.Add(1)
		go func(j agentJob) {
			defer wg.Done()
			res, err := p.gemini.ExecutePrompt(ctx, j.name, j.vars, true)
			if err != nil {
				fmt.Printf("Error executing AI agent %s: %v\n", j.name, err)
				return
			}
			insightsChan <- domain.AIInsight{
				UserID:     userID,
				SnapshotID: &snap.ID,
				Type:       j.type_,
				Content:    res,
				Priority:   2,
			}
		}(job)
	}

	// Wait for all agents to finish
	go func() {
		wg.Wait()
		close(insightsChan)
	}()

	var allInsights []domain.AIInsight
	for ins := range insightsChan {
		allInsights = append(allInsights, ins)
	}

	// 6. Save insights batch
	if len(allInsights) > 0 {
		return p.repoInsight.SaveBatch(ctx, allInsights)
	}

	return nil
}

func (p *AnalysisPipeline) enrichWithForecast(ctx context.Context, snap *domain.AnalysisSnapshot, user *domain.User, txns []domain.Transaction) {
	now := time.Now()
	// If current period is in the past, no forecast needed
	if now.After(snap.PeriodEnd) {
		return
	}

	daysInMonth := snap.PeriodEnd.Sub(snap.PeriodStart).Hours() / 24
	daysElapsed := now.Sub(snap.PeriodStart).Hours() / 24
	if daysElapsed < 1 {
		daysElapsed = 1
	}
	daysRemaining := daysInMonth - daysElapsed

	// Simple Burn Rate Calculation
	dailyBurnRate := snap.TotalExpense.Div(decimal.NewFromFloat(daysElapsed))
	projectedExtraExpense := dailyBurnRate.Mul(decimal.NewFromFloat(daysRemaining))
	
	// Predicted End balance = Income - Current Expense - Projected Expense
	snap.ForecastEndBalance = snap.TotalIncome.Sub(snap.TotalExpense).Sub(projectedExtraExpense)
	
	// Calculate confidence based on data points
	confidence := float64(len(txns)) / (daysElapsed * 2) // Rough heuristic
	if confidence > 0.95 {
		confidence = 0.95
	}
	snap.ForecastConfidence = confidence
}

func (p *AnalysisPipeline) detectSubscriptions(txns []domain.Transaction) []byte {
	// Simple Logic: Group by merchant and check for recurring patterns
	// In a real production app, this would be much more complex.
	recurring := make(map[string]int)
	for _, t := range txns {
		if t.Merchant != nil {
			recurring[*t.Merchant]++
		}
	}

	var detected []string
	for merchant, count := range recurring {
		if count >= 1 { // If seen at least once per month
			detected = append(detected, merchant)
		}
	}
	res, _ := json.Marshal(detected)
	return res
}

func (p *AnalysisPipeline) buildSnapshot(userID uuid.UUID, start, end time.Time, txns []domain.Transaction) *domain.AnalysisSnapshot {
	snap := &domain.AnalysisSnapshot{
		UserID:            userID,
		PeriodType:        "MONTHLY",
		PeriodStart:       start,
		PeriodEnd:         end,
		TransactionCount:  len(txns),
		CategoryBreakdown: make(map[string]decimal.Decimal),
		MerchantBreakdown: make(map[string]decimal.Decimal),
		DailyTrend:        make(map[string]decimal.Decimal),
	}

	for _, t := range txns {
		if t.Amount.GreaterThan(decimal.Zero) {
			snap.TotalIncome = snap.TotalIncome.Add(t.AmountInBase)
		} else {
			snap.TotalExpense = snap.TotalExpense.Add(t.AmountInBase.Abs())
		}
		// ... more complex breakdown logic would go here ...
	}
	snap.NetSavings = snap.TotalIncome.Sub(snap.TotalExpense)
	snap.NetCashflow = snap.NetSavings
	return snap
}

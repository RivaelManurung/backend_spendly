package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/spendly/backend/internal/ai"
	"github.com/spendly/backend/internal/domain"
	"github.com/spendly/backend/internal/repository"
)

type AnalysisPipeline struct {
	gemini       *ai.GeminiClient
	repoTxn      repository.TransactionRepository
	repoSnap     repository.AnalysisRepository
	repoInsight  repository.InsightRepository
}

func NewAnalysisPipeline(g *ai.GeminiClient, repoTxn repository.TransactionRepository, repoSnap repository.AnalysisRepository, repoInsight repository.InsightRepository) *AnalysisPipeline {
	return &AnalysisPipeline{
		gemini:      g,
		repoTxn:     repoTxn,
		repoSnap:    repoSnap,
		repoInsight: repoInsight,
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

func (p *AnalysisPipeline) buildSnapshot(userID uuid.UUID, start, end time.Time, txns []domain.Transaction) *domain.AnalysisSnapshot {
	// Just a simple builder for now
	return &domain.AnalysisSnapshot{
		UserID:      userID,
		PeriodType:  "MONTHLY",
		PeriodStart: start,
		PeriodEnd:   end,
		// ... and other calculations ...
	}
}

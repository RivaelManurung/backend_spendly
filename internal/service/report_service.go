package service

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/spendly/backend/internal/ai"
	"github.com/spendly/backend/internal/domain"
)

// 8. Daily Insight Job (daily_insight_job)
type ReportPipeline struct {
	gemini *ai.GeminiClient
}

func NewReportPipeline(g *ai.GeminiClient) *ReportPipeline {
	return &ReportPipeline{gemini: g}
}

// GenerateDailyDigest me-run daily_digest.prompt pagi hari setiap jam 08:00 (Cron)
func (r *ReportPipeline) GenerateDailyDigest(ctx context.Context, userID uuid.UUID, date string, yesterdayTxns []domain.Transaction) (string, error) {
	txnsBytes, _ := json.Marshal(yesterdayTxns)
	
	res, err := r.gemini.ExecutePrompt(ctx, "daily_digest", map[string]string{
		"date":             date,
		"transactions":     string(txnsBytes),
	}, false)

	return res, err
}

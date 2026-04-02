package ai

import (
	"context"
)

// LLMClient adalah interface umum untuk berinteraksi dengan AI/LLM models.
type LLMClient interface {
	// ExecutePrompt menjalankan prompt template dengan variabel tertentu.
	// templateName adalah nama file (misal: "monthly_analysis")
	ExecutePrompt(ctx context.Context, templateName string, vars map[string]string, useJson bool) (string, error)
}

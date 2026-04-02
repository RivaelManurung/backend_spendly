package ai

import (
	"context"
	"fmt"
)

type GroqClient struct {
	apiKey string
}

func NewGroqClient(apiKey string) *GroqClient {
	return &GroqClient{apiKey: apiKey}
}

func (g *GroqClient) ExecutePrompt(ctx context.Context, templateName string, vars map[string]string, useJson bool) (string, error) {
	return "", fmt.Errorf("groq provider not fully implemented yet")
}

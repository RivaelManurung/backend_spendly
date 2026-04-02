package ai

import (
	"context"
	"fmt"
)

type OllamaClient struct {
	baseURL string
}

func NewOllamaClient(baseURL string) *OllamaClient {
	return &OllamaClient{baseURL: baseURL}
}

func (o *OllamaClient) ExecutePrompt(ctx context.Context, templateName string, vars map[string]string, useJson bool) (string, error) {
	return "", fmt.Errorf("ollama provider not fully implemented yet")
}

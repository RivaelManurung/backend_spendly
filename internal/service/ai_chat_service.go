package service

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/spendly/backend/internal/ai"
	"github.com/spendly/backend/internal/domain"
)

// 5. & 6. Chat Analyst & Intent Parser Pipeline (chat_analyst_pipeline)
type ChatPipeline struct {
	gemini *ai.GeminiClient
}

func NewChatPipeline(g *ai.GeminiClient) *ChatPipeline {
	return &ChatPipeline{gemini: g}
}

// Menjalankan intent_parser.prompt
func (c *ChatPipeline) ParseUserIntent(ctx context.Context, message string) (string, error) {
	res, err := c.gemini.ExecutePrompt(ctx, "intent_parser", map[string]string{
		"user_message": message,
	}, true)
	return res, err
}

// Menjalankan chat_analyst.prompt sebagai chatbot pintar
func (c *ChatPipeline) AskAnalyst(ctx context.Context, userID uuid.UUID, message string, contextData []domain.Transaction) (string, error) {
	ctxBytes, _ := json.Marshal(contextData)

	res, err := c.gemini.ExecutePrompt(ctx, "chat_analyst", map[string]string{
		"user_message": message,
		"context_data": string(ctxBytes),
	}, false) // Karena ini chat text balasan (Markdown biasa, bukan json codeblock)
	
	return res, err
}

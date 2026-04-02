package service

import (
	"context"

	"github.com/spendly/backend/internal/ai"
)

// OCRService menangani pipeline `ocr_scan_pipeline` dari README yang berkaitan dengan struk belanja.
type OCRService struct {
	gemini *ai.GeminiClient
}

func NewOCRService(g *ai.GeminiClient) *OCRService {
	return &OCRService{gemini: g}
}

// ExtractReceipt melakukan parsing struct text (contoh dari Google Cloud Vision) ke dalam format JSON 
// dengan menggunakan prompt `ocr_parse.prompt`.
func (o *OCRService) ExtractReceipt(ctx context.Context, rawOCRText string) (string, error) {
	// Menjalankan ocr_parse.prompt
	res, err := o.gemini.ExecutePrompt(ctx, "ocr_parse", map[string]string{
		"raw_text": rawOCRText,
	}, true)

	return res, err
}

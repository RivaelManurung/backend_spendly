package http

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spendly/backend/internal/service"
)

type OCRHandler struct {
	aiSvc service.AIService
}

func NewOCRHandler(aiSvc service.AIService) *OCRHandler {
	return &OCRHandler{aiSvc: aiSvc}
}

func (h *OCRHandler) ScanReceipt(c *gin.Context) {
	// Multipart form parsing
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no image file received"})
		return
	}
	defer file.Close()

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed reading image bytes"})
		return
	}

	// Make sure we pass correct mime type
	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "image/jpeg"
	}

	// Send to AIService / Gemini M-Multi OCR
	receiptData, err := h.aiSvc.ScanReceipt(c.Request.Context(), buf.Bytes(), mimeType)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Receipt processed successfully",
		"data":    receiptData,
	})
}

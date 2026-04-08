package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spendly/backend/internal/domain"
	"github.com/spendly/backend/internal/service"
)

type TransactionHandler struct {
	txSvc service.TransactionService
}

func NewTransactionHandler(txSvc service.TransactionService) *TransactionHandler {
	return &TransactionHandler{txSvc: txSvc}
}

func (h *TransactionHandler) Create(c *gin.Context) {
	var req struct {
		ID          string    `json:"id"`
		Title       string    `json:"title" binding:"required"`
		Amount      float64   `json:"amount" binding:"required"`
		CategoryID  string    `json:"category_id" binding:"required"`
		Type        string    `json:"type" binding:"required"` // income, expense, goal
		Note        string    `json:"note"`
		IsRecurring bool      `json:"is_recurring"`
		Date        time.Time `json:"date"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deviceID := c.GetHeader("X-Device-ID")

	var tx *domain.Transaction
	var err error

	if req.ID != "" {
		// This is likely a sync request with a pre-defined local ID
		if req.Date.IsZero() {
			req.Date = time.Now()
		}
		tx, err = h.txSvc.SyncTransaction(c.Request.Context(), req.ID, req.Title, req.Amount, req.CategoryID, req.Type, req.Note, req.IsRecurring, req.Date, deviceID)
	} else {
		tx, err = h.txSvc.CreateTransaction(c.Request.Context(), req.Title, req.Amount, req.CategoryID, req.Type, req.Note, req.IsRecurring)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, tx)
}

func (h *TransactionHandler) GetTransactions(c *gin.Context) {
	txs, err := h.txSvc.GetAllTransactions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": txs})
}

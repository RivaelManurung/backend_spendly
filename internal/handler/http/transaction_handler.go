package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spendly/backend/internal/handler/dto"
	"github.com/spendly/backend/internal/repository"
)

type TransactionHandler struct {
	repo repository.TransactionRepository
}

func NewTransactionHandler(repo repository.TransactionRepository) *TransactionHandler {
	return &TransactionHandler{repo: repo}
}

// ListTransactions returns paginated transactions for a user
// @Summary List Transactions
// @Description Get a list of transactions for current user with pagination
// @Tags transactions
// @Param limit query int false "Results per page"
// @Param offset query int false "Offset results"
// @Produce json
// @Success 200 {array} dto.TransactionResponse
// @Router /transactions [get]
func (h *TransactionHandler) ListTransactions(c *gin.Context) {
	userIDStr := c.Query("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	limitStr := c.Query("limit")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 20
	}

	offsetStr := c.Query("offset")
	offset, _ := strconv.Atoi(offsetStr)

	txns, err := h.repo.FindAllByUserID(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch transactions"})
		return
	}

	// Map to DTOs
	resp := make([]dto.TransactionResponse, len(txns))
	for i, t := range txns {
		resp[i] = dto.TransactionResponse{
			ID:                   t.ID.String(),
			CategoryID:           t.CategoryID,
			Amount:               t.Amount,
			Currency:             t.Currency,
			AmountInBase:         t.AmountInBase,
			Description:          t.Description,
			Merchant:             t.Merchant,
			Source:               t.Source,
			TransactionDate:      t.TransactionDate,
			AICategorySuggestion: t.AICategorySuggestion,
			AIConfidenceScore:    t.AIConfidenceScore,
			CreatedAt:            t.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, resp)
}

// GetByID returns a single transaction by ID
// @Summary Get Transaction
// @Description Get a single transaction by its UUID
// @Tags transactions
// @Param id path string true "Transaction UUID"
// @Produce json
// @Success 200 {object} dto.TransactionResponse
// @Router /transactions/:id [get]
func (h *TransactionHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction ID"})
		return
	}

	txn, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch transaction"})
		return
	}

	if txn == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
		return
	}

	c.JSON(http.StatusOK, dto.TransactionResponse{
		ID:                   txn.ID.String(),
		CategoryID:           txn.CategoryID,
		Amount:               txn.Amount,
		Currency:             txn.Currency,
		AmountInBase:         txn.AmountInBase,
		Description:          txn.Description,
		Merchant:             txn.Merchant,
		Source:               txn.Source,
		TransactionDate:      txn.TransactionDate,
		AICategorySuggestion: txn.AICategorySuggestion,
		AIConfidenceScore:    txn.AIConfidenceScore,
		CreatedAt:            txn.CreatedAt,
	})
}

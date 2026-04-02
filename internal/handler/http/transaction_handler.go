package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/spendly/backend/internal/handler/dto"
	"github.com/spendly/backend/internal/repository"
	"github.com/spendly/backend/pkg/apperror"
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
// @Security Bearer
func (h *TransactionHandler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	// In a real app, userID would come from JWT context.
	// For now, let's look for a query param or placeholder.
	userIDStr := r.URL.Query().Get("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		respondWithError(w, apperror.BadRequest("Invalid user_id", err))
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 20
	}

	offsetStr := r.URL.Query().Get("offset")
	offset, _ := strconv.Atoi(offsetStr)

	txns, err := h.repo.FindAllByUserID(r.Context(), userID, limit, offset)
	if err != nil {
		respondWithError(w, apperror.Internal("Failed to fetch transactions", err))
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

	respondWithJSON(w, http.StatusOK, resp)
}

// GetByID returns a single transaction by ID
// @Summary Get Transaction
// @Description Get a single transaction by its UUID
// @Tags transactions
// @Param id path string true "Transaction UUID"
// @Produce json
// @Success 200 {object} dto.TransactionResponse
// @Router /transactions/{id} [get]
// @Security Bearer
func (h *TransactionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, apperror.BadRequest("Invalid transaction ID", err))
		return
	}

	txn, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		respondWithError(w, apperror.Internal("Failed to fetch transaction", err))
		return
	}

	if txn == nil {
		respondWithError(w, apperror.NotFound("Transaction not found", nil))
		return
	}

	respondWithJSON(w, http.StatusOK, dto.TransactionResponse{
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

// Helpers for responding (Usually these would be in a base handler or util package)

func respondWithError(w http.ResponseWriter, err *apperror.AppError) {
	respondWithJSON(w, err.Code, map[string]string{"error": err.Message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

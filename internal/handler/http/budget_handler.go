package http

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/spendly/backend/internal/handler/dto"
	"github.com/spendly/backend/internal/repository"
	"github.com/spendly/backend/pkg/apperror"
)

type BudgetHandler struct {
	repo repository.BudgetRepository
}

func NewBudgetHandler(repo repository.BudgetRepository) *BudgetHandler {
	return &BudgetHandler{repo: repo}
}

// ListActiveBudgets returns active budgets for a user
// @Summary List Active Budgets
// @Description Get all active budgets for current user and their current progress
// @Tags budgets
// @Param user_id query string true "User UUID"
// @Produce json
// @Success 200 {array} dto.BudgetResponse
// @Router /budgets [get]
// @Security Bearer
func (h *BudgetHandler) ListActiveBudgets(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		respondWithError(w, apperror.BadRequest("Invalid user_id", err))
		return
	}

	budgets, err := h.repo.GetActiveBudgetsByUser(r.Context(), userID)
	if err != nil {
		respondWithError(w, apperror.Internal("Failed to fetch budgets", err))
		return
	}

	resp := make([]dto.BudgetResponse, len(budgets))
	for i, b := range budgets {
		spent, _ := h.repo.GetSpentAmount(r.Context(), b.ID)
		usagePercent := 0.0
		limitF, _ := b.LimitAmount.Float64()
		if limitF > 0 {
			usagePercent = (spent / limitF) * 100
		}

		catID := int64(0)
		if b.CategoryID != nil {
			catID = *b.CategoryID
		}

		resp[i] = dto.BudgetResponse{
			ID:           b.ID,
			CategoryID:   catID,
			Period:       b.Period,
			LimitAmount:  b.LimitAmount,
			SpentAmount:  spent,
			Currency:     b.Currency,
			IsActive:     b.IsActive,
			UsagePercent: usagePercent,
		}
	}

	respondWithJSON(w, http.StatusOK, resp)
}

func (h *BudgetHandler) CreateBudget(w http.ResponseWriter, r *http.Request) {
	c := map[string]string{"message": "Not implemented"}
	respondWithJSON(w, http.StatusNotImplemented, c)
}

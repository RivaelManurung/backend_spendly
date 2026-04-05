package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spendly/backend/internal/handler/dto"
	"github.com/spendly/backend/internal/repository"
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
func (h *BudgetHandler) ListActiveBudgets(c *gin.Context) {
	userIDStr := c.Query("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	budgets, err := h.repo.GetActiveBudgetsByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch budgets"})
		return
	}

	resp := make([]dto.BudgetResponse, len(budgets))
	for i, b := range budgets {
		spent, _ := h.repo.GetSpentAmount(c.Request.Context(), b.ID)
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

	c.JSON(http.StatusOK, resp)
}

func (h *BudgetHandler) CreateBudget(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented"})
}

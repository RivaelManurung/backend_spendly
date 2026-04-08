package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spendly/backend/internal/service"
)

type BudgetHandler struct {
	budgetSvc service.BudgetService
}

func NewBudgetHandler(budgetSvc service.BudgetService) *BudgetHandler {
	return &BudgetHandler{budgetSvc: budgetSvc}
}

func (h *BudgetHandler) Create(c *gin.Context) {
	var req struct {
		CategoryID string  `json:"category_id" binding:"required"`
		Amount     float64 `json:"amount" binding:"required"`
		Period     string  `json:"period" binding:"required"` // monthly, weekly
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	budget, err := h.budgetSvc.CreateBudget(c.Request.Context(), req.CategoryID, req.Amount, req.Period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, budget)
}

func (h *BudgetHandler) Status(c *gin.Context) {
	reports, err := h.budgetSvc.GetBudgetsWithBurnRate(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": reports,
	})
}

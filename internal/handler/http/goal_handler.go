package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spendly/backend/internal/service"
)

type GoalHandler struct {
	goalSvc service.GoalService
}

func NewGoalHandler(goalSvc service.GoalService) *GoalHandler {
	return &GoalHandler{goalSvc: goalSvc}
}

func (h *GoalHandler) Create(c *gin.Context) {
	var req struct {
		Title        string  `json:"title" binding:"required"`
		TargetAmount float64 `json:"target_amount" binding:"required"`
		TargetDate   string  `json:"target_date" binding:"required"` // Format: YYYY-MM-DD
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tTime, err := time.Parse("2006-01-02", req.TargetDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format. expected YYYY-MM-DD"})
		return
	}

	g, err := h.goalSvc.CreateGoal(c.Request.Context(), req.Title, req.TargetAmount, tTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, g)
}

func (h *GoalHandler) List(c *gin.Context) {
	goals, err := h.goalSvc.GetAllGoals(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": goals})
}

func (h *GoalHandler) AddContribution(c *gin.Context) {
	goalID := c.Param("id")

	var req struct {
		Amount float64 `json:"amount" binding:"required"`
		Note   string  `json:"note"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.goalSvc.ContributeToGoal(c.Request.Context(), goalID, req.Amount, req.Note)
	if err != nil {
		fmt.Println("Err contributing", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

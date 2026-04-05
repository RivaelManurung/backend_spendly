package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spendly/backend/internal/repository"
)

type UserHandler struct {
	repo repository.UserRepository
}

func NewUserHandler(repo repository.UserRepository) *UserHandler {
	return &UserHandler{repo: repo}
}

// GetProfile returns public user profile information
// @Summary Get User Profile
// @Description Fetch profile details for a given User UUID
// @Tags users
// @Param id path string true "User UUID"
// @Produce json
// @Success 200 {object} domain.User
// @Failure 404 {object} map[string]string
// @Router /users/{id} [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	user, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateProfile updates user financial and persona metadata
// @Summary Update User Profile
// @Description Updates user fields like salary_cycle_day, financial_goals, etc.
// @Tags users
// @Accept json
// @Produce json
// @Router /users/{id} [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var input struct {
		Name               *string  `json:"name"`
		CurrencyPreference string   `json:"currency_preference"`
		SalaryCycleDay     int      `json:"salary_cycle_day"`
		FinancialGoals     []string `json:"financial_goals"`
		RiskProfile        string   `json:"risk_profile"`
		AIAnalystPersona   string   `json:"ai_analyst_persona"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Update fields
	if input.Name != nil {
		user.Name = input.Name
	}
	if input.CurrencyPreference != "" {
		user.CurrencyPreference = input.CurrencyPreference
	}
	if input.SalaryCycleDay != 0 {
		user.SalaryCycleDay = input.SalaryCycleDay
	}
	user.FinancialGoals = input.FinancialGoals
	if input.RiskProfile != "" {
		user.RiskProfile = input.RiskProfile
	}
	if input.AIAnalystPersona != "" {
		user.AIAnalystPersona = input.AIAnalystPersona
	}

	if err := h.repo.Update(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

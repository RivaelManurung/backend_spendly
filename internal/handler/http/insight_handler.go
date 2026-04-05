package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spendly/backend/internal/repository"
)

type InsightHandler struct {
	repo repository.InsightRepository
}

func NewInsightHandler(repo repository.InsightRepository) *InsightHandler {
	return &InsightHandler{repo: repo}
}

// GetLatestInsights returns the most recent insights for a user
func (h *InsightHandler) GetLatestInsights(c *gin.Context) {
	h.ListUnread(c)
}

// ListUnread returns unread insights for a user (alias for GetLatestInsights)
func (h *InsightHandler) ListUnread(c *gin.Context) {
	userIDStr := c.Query("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, _ := strconv.Atoi(limitStr)

	insights, err := h.repo.GetLatestByUserID(c.Request.Context(), userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, insights)
}

// MarkAsRead marks an insight as read
// @Summary Mark Insight as Read
// @Description Update the 'is_read' status of an insight
// @Tags insights
// @Param id path int true "Insight ID"
// @Produce json
// @Success 200 {object} map[string]string
// @Router /insights/{id}/read [patch]
func (h *InsightHandler) MarkAsRead(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid insight ID"})
		return
	}

	if err := h.repo.MarkRead(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

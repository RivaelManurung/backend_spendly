package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spendly/backend/internal/repository"
)

type ReportHandler struct {
	repo repository.AnalysisRepository
}

func NewReportHandler(repo repository.AnalysisRepository) *ReportHandler {
	return &ReportHandler{repo: repo}
}

// GetLatestReport returns the latest analysis for current user
// @Summary Latest Analysis Report
// @Description Get current user's latest monthly analysis snapshot
// @Tags reports
// @Param user_id query string true "User UUID"
// @Produce json
// @Success 200 {object} domain.AnalysisSnapshot
// @Router /reports/latest [get]
func (h *ReportHandler) GetLatestReport(c *gin.Context) {
	userIDStr := c.Query("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	snapshot, err := h.repo.GetLatestByUserID(c.Request.Context(), userID, "MONTHLY")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch latest report"})
		return
	}

	if snapshot == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no analysis found"})
		return
	}

	c.JSON(http.StatusOK, snapshot)
}

func (h *ReportHandler) TriggerAnalysis(c *gin.Context) {
	// ... manually trigger a snapshot generation for testing
	c.JSON(http.StatusOK, gin.H{"message": "Analysis queued"})
}

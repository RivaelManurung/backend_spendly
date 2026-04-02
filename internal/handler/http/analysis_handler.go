package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spendly/backend/internal/handler/dto"
	"github.com/spendly/backend/internal/repository"
)

type AnalysisHandler struct {
	repoSnap repository.AnalysisRepository
}

func NewAnalysisHandler(repoSnap repository.AnalysisRepository) *AnalysisHandler {
	return &AnalysisHandler{
		repoSnap: repoSnap,
	}
}

// GetLatestSnapshot returns the most recent monthly analysis for a user
// @Summary Get latest analysis snapshot
// @Description Fetch the latest monthly analysis for a given user UUID
// @Tags analysis
// @Param user_id query string true "User UUID"
// @Produce json
// @Success 200 {object} dto.AnalysisSnapshotResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /analysis/latest [get]
func (h *AnalysisHandler) GetLatestSnapshot(c *gin.Context) {
	userIDStr := c.Query("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	snapshot, err := h.repoSnap.GetLatestByUserID(c.Request.Context(), userID, "monthly")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if snapshot == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "no snapshot found"})
		return
	}

	c.JSON(http.StatusOK, dto.FromAnalysisSnapshot(snapshot))
}

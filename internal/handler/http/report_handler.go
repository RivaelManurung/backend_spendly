package http

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/spendly/backend/internal/repository"
	"github.com/spendly/backend/pkg/apperror"
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
// @Security Bearer
func (h *ReportHandler) GetLatestReport(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		respondWithError(w, apperror.BadRequest("Invalid user_id", err))
		return
	}

	snapshot, err := h.repo.GetLatestByUserID(r.Context(), userID, "MONTHLY")
	if err != nil {
		respondWithError(w, apperror.Internal("Failed to fetch latest report", err))
		return
	}

	if snapshot == nil {
		respondWithError(w, apperror.NotFound("No analysis found", nil))
		return
	}

	respondWithJSON(w, http.StatusOK, snapshot)
}
func (h *ReportHandler) TriggerAnalysis(w http.ResponseWriter, r *http.Request) {
	// ... manually trigger a snapshot generation for testing
	c := map[string]string{"message": "Analysis queued"}
	respondWithJSON(w, http.StatusOK, c)
}

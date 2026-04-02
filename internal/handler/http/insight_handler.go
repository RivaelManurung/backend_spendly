package http

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/spendly/backend/internal/repository"
	"github.com/spendly/backend/pkg/apperror"
)

type InsightHandler struct {
	repo repository.InsightRepository
}

func NewInsightHandler(repo repository.InsightRepository) *InsightHandler {
	return &InsightHandler{repo: repo}
}

// ListUnread returns unread insights for a user
// @Summary List Unread Insights
// @Description Get current user's unread AI insights
// @Tags insights
// @Param user_id query string true "User UUID"
// @Produce json
// @Success 200 {array} domain.AIInsight
// @Router /insights/unread [get]
// @Security Bearer
func (h *InsightHandler) ListUnread(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		respondWithError(w, apperror.BadRequest("Invalid user_id", err))
		return
	}

	insights, err := h.repo.ListUnread(r.Context(), userID)
	if err != nil {
		respondWithError(w, apperror.Internal("Failed to fetch insights", err))
		return
	}

	respondWithJSON(w, http.StatusOK, insights)
}

// MarkAsRead marks an insight as read
// @Summary Mark Insight as Read
// @Description Update the 'is_read' status of an insight
// @Tags insights
// @Param id path int true "Insight ID"
// @Produce json
// @Success 200 {object} map[string]string
// @Router /insights/{id}/read [patch]
// @Security Bearer
func (h *InsightHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondWithError(w, apperror.BadRequest("Invalid insight ID", err))
		return
	}

	if err := h.repo.MarkRead(r.Context(), id); err != nil {
		respondWithError(w, apperror.Internal("Failed to mark insight as read", err))
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

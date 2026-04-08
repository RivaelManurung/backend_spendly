package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spendly/backend/internal/service"
)

type AnalyticsHandler struct {
	reportSvc service.ReportService
	syncSvc   service.SyncService
}

func NewAnalyticsHandler(rSvc service.ReportService, sSvc service.SyncService) *AnalyticsHandler {
	return &AnalyticsHandler{reportSvc: rSvc, syncSvc: sSvc}
}

func (h *AnalyticsHandler) GetMonthlyReport(c *gin.Context) {
	report, err := h.reportSvc.GetMonthlyAnalytics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": report})
}

func (h *AnalyticsHandler) SynchronizeChanges(c *gin.Context) {
	// Ex. query ?last_sync=2024-10-01T20:20:20Z
	lastSyncStr := c.Query("last_sync")
	var lastSync time.Time

	if lastSyncStr != "" {
		t, err := time.Parse(time.RFC3339, lastSyncStr)
		if err == nil {
			lastSync = t
		}
	}

	syncResp, err := h.syncSvc.SyncTransactions(c.Request.Context(), lastSync)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": syncResp})
}

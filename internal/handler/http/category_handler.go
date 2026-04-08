package http

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/spendly/backend/internal/service"
)

type CategoryHandler struct {
	svc service.CategoryService
}

func NewCategoryHandler(svc service.CategoryService) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

func (h *CategoryHandler) Create(c *gin.Context) {
	var req struct {
		ID    string `json:"id" binding:"required"`
		Label string `json:"label" binding:"required"`
		Icon  string `json:"icon"`
		Color string `json:"color"`
		Type  string `json:"type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cat, err := h.svc.SyncCategory(c.Request.Context(), req.ID, req.Label, req.Icon, req.Color, req.Type)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, cat)
}

func (h *CategoryHandler) List(c *gin.Context) {
	cats, err := h.svc.GetAllCategories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": cats})
}

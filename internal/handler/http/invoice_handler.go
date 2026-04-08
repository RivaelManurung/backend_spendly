package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spendly/backend/internal/service"
)

type InvoiceHandler struct {
	invSvc service.InvoiceService
}

func NewInvoiceHandler(invSvc service.InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{invSvc: invSvc}
}

func (h *InvoiceHandler) Create(c *gin.Context) {
	var req struct {
		ClientName  string                `json:"client_name" binding:"required"`
		ClientEmail string                `json:"client_email"`
		DueDate     string                `json:"due_date" binding:"required"`
		Items       []service.InvoiceItem `json:"items" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dTime, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid due_date. Use YYYY-MM-DD"})
		return
	}

	invoice, err := h.invSvc.CreateInvoice(c.Request.Context(), req.ClientName, req.ClientEmail, dTime, req.Items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, invoice)
}

func (h *InvoiceHandler) GeneratePDF(c *gin.Context) {
	invoiceID := c.Param("id")

	pdfBytes, err := h.invSvc.GenerateInvoicePDF(c.Request.Context(), invoiceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed generating pdf: " + err.Error()})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=invoice_"+invoiceID[:8]+".pdf")
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

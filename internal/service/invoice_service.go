package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/spendly/backend/internal/domain"
	"github.com/spendly/backend/internal/repository"
)

type InvoiceItem struct {
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Qty         int     `json:"qty"`
}

type InvoiceService interface {
	CreateInvoice(ctx context.Context, clientName, clientEmail string, dueDate time.Time, items []InvoiceItem) (*domain.Invoice, error)
	GenerateInvoicePDF(ctx context.Context, invoiceID string) ([]byte, error)
}

type invoiceService struct {
	invoiceRepo repository.InvoiceRepository
}

func NewInvoiceService(invRepo repository.InvoiceRepository) InvoiceService {
	return &invoiceService{invoiceRepo: invRepo}
}

func (s *invoiceService) CreateInvoice(ctx context.Context, clientName, clientEmail string, dueDate time.Time, items []InvoiceItem) (*domain.Invoice, error) {
	var total float64
	for _, item := range items {
		total += item.Price * float64(item.Qty)
	}

	// Calculate a simple tax estimation (e.g. 11% PPN)
	tax := total * 0.11
	finalAmount := total + tax

	itemsBytes, err := json.Marshal(items)
	if err != nil {
		return nil, err
	}

	invoice := &domain.Invoice{
		ClientName:  clientName,
		ClientEmail: clientEmail,
		Amount:      finalAmount, // Includes 11% tax
		DueDate:     dueDate,
		Status:      "draft",
		Items:       string(itemsBytes),
	}

	if err := s.invoiceRepo.Create(ctx, invoice); err != nil {
		return nil, err
	}

	return invoice, nil
}

func (s *invoiceService) GenerateInvoicePDF(ctx context.Context, invoiceID string) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, fmt.Sprintf("INVOICE #%s", invoiceID[:8]))

	pdf.Ln(12)
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, "Status: DRAFT")

	pdf.Ln(12)
	pdf.Cell(40, 10, "Thank you for doing business with us!")

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %v", err)
	}

	return buf.Bytes(), nil
}

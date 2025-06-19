package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iqsanfm/dashboard-pekerjaan-backend/internal/models"
	"github.com/iqsanfm/dashboard-pekerjaan-backend/internal/repositories"
	"github.com/iqsanfm/dashboard-pekerjaan-backend/pkg/auth" // Sesuaikan dengan lokasi package auth Anda
)

type InvoiceHandler struct {
	repo repositories.InvoiceRepository
}

func NewInvoiceHandler(repo repositories.InvoiceRepository) *InvoiceHandler {
	return &InvoiceHandler{repo: repo}
}

// CreateInvoice menangani permintaan untuk membuat invoice baru.
func (h *InvoiceHandler) CreateInvoice(c *gin.Context) {
	var invoice models.Invoice
	if err := c.ShouldBindJSON(&invoice); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// Di sini Anda bisa menambahkan validasi lebih lanjut jika perlu

	if err := h.repo.CreateInvoice(&invoice); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create invoice: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, invoice)
}

// GetAllInvoices menangani permintaan untuk mendapatkan semua invoice.
func (h *InvoiceHandler) GetAllInvoices(c *gin.Context) {
	claims, exists := c.Get("user_claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User claims not found"})
		return
	}

	userClaims := claims.(*auth.Claims) // Lakukan type assertion ke struct Claims Anda

	invoices, err := h.repo.GetAllInvoices(userClaims.StaffID, userClaims.IsAdmin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get invoices: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, invoices)
}

// GetInvoiceByID menangani permintaan untuk mendapatkan satu invoice berdasarkan ID.
func (h *InvoiceHandler) GetInvoiceByID(c *gin.Context) {
	invoiceID := c.Param("id")

	claims, exists := c.Get("user_claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User claims not found"})
		return
	}

	userClaims := claims.(*auth.Claims)

	invoice, err := h.repo.GetInvoiceByID(invoiceID, userClaims.StaffID, userClaims.IsAdmin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get invoice by ID: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, invoice)
}
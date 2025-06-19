package services

import (
	"fmt"
	"log"
	"time"

	"github.com/iqsanfm/dashboard-pekerjaan-backend/internal/models"
	"github.com/iqsanfm/dashboard-pekerjaan-backend/internal/repositories"
)

// InvoiceService mendefinisikan operasi-operasi untuk logika bisnis invoice.
type InvoiceService interface {
	CreateInvoiceFromJob(jobID, jobType, clientID, assignedStaffID string) (*models.Invoice, error)
}

// invoiceService adalah implementasi dari InvoiceService.
type invoiceService struct {
	invoiceRepo      repositories.InvoiceRepository
	monthlyJobRepo   repositories.MonthlyJobRepository
	annualJobRepo    repositories.AnnualJobRepository
	sp2dkJobRepo     repositories.Sp2dkJobRepository
	pemeriksaanJobRepo repositories.PemeriksaanJobRepository
}

// NewInvoiceService adalah constructor untuk invoiceService.
func NewInvoiceService(
	invRepo repositories.InvoiceRepository,
	mjRepo repositories.MonthlyJobRepository,
	ajRepo repositories.AnnualJobRepository,
	sjRepo repositories.Sp2dkJobRepository,
	pjRepo repositories.PemeriksaanJobRepository,
) InvoiceService {
	return &invoiceService{
		invoiceRepo:      invRepo,
		monthlyJobRepo:   mjRepo,
		annualJobRepo:    ajRepo,
		sp2dkJobRepo:     sjRepo,
		pemeriksaanJobRepo: pjRepo,
	}
}

// CreateInvoiceFromJob membuat invoice secara otomatis dari pekerjaan yang telah selesai.
// internal/services/invoice_service.go

// internal/services/invoice_service.go

func (s *invoiceService) CreateInvoiceFromJob(jobID, jobType, clientID, assignedStaffID string) (*models.Invoice, error) {
	var lineItems []models.InvoiceLineItem
	var totalAmount float64
	var description string

	// Tentukan detail invoice berdasarkan jenis pekerjaan.
	switch jobType {
	case "Pekerjaan Bulanan":
		description = fmt.Sprintf("Jasa Akuntansi dan Pajak Bulanan (Job ID: %s)", jobID)
		totalAmount = 1500000.00
	case "Pekerjaan Tahunan":
		description = fmt.Sprintf("Jasa Laporan Pajak Tahunan (Job ID: %s)", jobID)
		totalAmount = 5000000.00
	case "Pemeriksaan":
		description = fmt.Sprintf("Jasa Pendampingan Pemeriksaan Pajak (Job ID: %s)", jobID)
		totalAmount = 10000000.00
	case "SP2DK":
		description = fmt.Sprintf("Jasa Respon SP2DK (Job ID: %s)", jobID)
		totalAmount = 2500000.00
	default:
		log.Printf("Tipe pekerjaan tidak dikenal '%s', invoice tidak dibuat.", jobType)
		return nil, fmt.Errorf("tipe pekerjaan tidak dikenal: %s", jobType)
	}

	// ================== AWAL BLOK PERBAIKAN ==================

	// Membuat satu line item berdasarkan deskripsi dan total harga
	lineItems = append(lineItems, models.InvoiceLineItem{
		Description:    description,
		Quantity:       totalAmount, // Pastikan tipe data sesuai (asumsi float64)
		UnitPrice:      1,           // Atau sebaliknya, quantity=1, unitprice=totalamount
		Amount:         totalAmount,
		RelatedJobType: &jobType, // Gunakan & untuk mendapatkan pointer string
		RelatedJobID:   &jobID,   // Gunakan & untuk mendapatkan pointer string
	})

	// Membuat objek invoice utama
	// Handle nullable assignedStaffID
	var assignedStaffIDPtr *string
	if assignedStaffID != "" {
		assignedStaffIDPtr = &assignedStaffID
	}
    
    notes := "Invoice otomatis dibuat dari penyelesaian pekerjaan."

	invoice := &models.Invoice{
		ClientID:        clientID,
		AssignedStaffID: assignedStaffIDPtr, // Gunakan variabel pointer
		InvoiceDate:     models.CustomDate{Time: time.Now()},
		DueDate:         models.CustomDate{Time: time.Now().Add(30 * 24 * time.Hour)},
		Status:          "Pending",
		Notes:           &notes, // Gunakan & untuk mendapatkan pointer string
		LineItems:       lineItems,
	}

	// ================== AKHIR BLOK PERBAIKAN ==================

	// Memanggil repository untuk menyimpan invoice ke database
	err := s.invoiceRepo.CreateInvoice(invoice)
	if err != nil {
		return nil, fmt.Errorf("gagal menyimpan invoice dari service: %w", err)
	}

	log.Printf("Invoice %s berhasil dibuat untuk pekerjaan %s.", invoice.InvoiceNumber, jobID)
	return invoice, nil
}
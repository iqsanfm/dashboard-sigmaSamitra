package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iqsanfm/dashboard-pekerjaan-backend/internal/models"       // Pastikan ini modul Anda
	"github.com/iqsanfm/dashboard-pekerjaan-backend/internal/repositories" // Pastikan ini modul Anda
	"github.com/iqsanfm/dashboard-pekerjaan-backend/internal/services"
	"github.com/iqsanfm/dashboard-pekerjaan-backend/pkg/auth"
)

// Sp2dkJobHandler handles HTTP requests for SP2DK job operations
type Sp2dkJobHandler struct {
	Sp2dkJobRepo repositories.Sp2dkJobRepository
	ClientRepo   repositories.ClientRepository
	StaffRepo    repositories.StaffRepository
	InvoiceService services.InvoiceService
}

// NewSp2dkJobHandler creates a new Sp2dkJobHandler
func NewSp2dkJobHandler(sjRepo repositories.Sp2dkJobRepository, cRepo repositories.ClientRepository, sRepo repositories.StaffRepository, invService services.InvoiceService) *Sp2dkJobHandler {
	return &Sp2dkJobHandler{
		Sp2dkJobRepo: sjRepo,
		ClientRepo:   cRepo,
		StaffRepo:    sRepo,
		InvoiceService: invService,
	}
}

// CreateSp2dkJob handles the creation of a new SP2DK job
func (h *Sp2dkJobHandler) CreateSp2dkJob(c *gin.Context) {
	var req models.NewSp2dkJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate client_id existence
	client, err := h.ClientRepo.GetClientByID(req.ClientID, "", true) // Admin check for client existence
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate client ID: " + err.Error()})
		return
	}
	if client == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Client ID not found"})
		return
	}

	// Validate assigned_pic_staff_sigma_id existence
	if req.AssignedPicStaffSigmaID != "" {
		staff, err := h.StaffRepo.GetStaffByID(req.AssignedPicStaffSigmaID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate Assigned PIC Staff ID: " + err.Error()})
			return
		}
		if staff == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Assigned PIC Staff ID not found"})
			return
		}
	}

	sp2dkJob := &models.Sp2dkJob{
		ClientID:              req.ClientID,
		AssignedPicStaffSigmaID: req.AssignedPicStaffSigmaID,
		ContractNo:            req.ContractNo,
		ContractDate:          req.ContractDate,
		Sp2dkNo:               req.Sp2dkNo,
		Sp2dkDate:             req.Sp2dkDate,
		Bap2dkNo:              req.Bap2dkNo,
		Bap2dkDate:            req.Bap2dkDate,
		PaymentDate:           req.PaymentDate,
		ReportDate:            req.ReportDate,
		OverallStatus:    			req.OverallStatus,
	}

	if err := h.Sp2dkJobRepo.CreateSp2dkJob(sp2dkJob); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create SP2DK job: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, sp2dkJob)
}

// GetAllSp2dkJobs fetches all SP2DK jobs, filtered by PIC if not admin
func (h *Sp2dkJobHandler) GetAllSp2dkJobs(c *gin.Context) {
		claims, exists := c.Get("user_claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User claims not found in context"})
		return
	}
	userClaims, ok := claims.(*auth.Claims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user claims"})
		return
	}

	jobs, err := h.Sp2dkJobRepo.GetAllSp2dkJobs(userClaims.StaffID, userClaims.IsAdmin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve SP2DK jobs: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, jobs)
}

// GetSp2dkJobByID fetches a single SP2DK job by ID, filtered by PIC if not admin
func (h *Sp2dkJobHandler) GetSp2dkJobByID(c *gin.Context) {
	id := c.Param("id")

	claims, exists := c.Get("user_claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User claims not found in context"})
		return
	}
	userClaims, ok := claims.(*auth.Claims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user claims"})
		return
	}

	job, err := h.Sp2dkJobRepo.GetSp2dkJobByID(id, userClaims.StaffID, userClaims.IsAdmin)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "SP2DK job not found or access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve SP2DK job: " + err.Error()})
		return
	}
	if job == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "SP2DK job not found or access denied"})
		return
	}
	c.JSON(http.StatusOK, job)
}

func (h *Sp2dkJobHandler) UpdateSp2dkJob(c *gin.Context) {
	id := c.Param("id")

	claims, exists := c.Get("user_claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User claims not found in context"})
		return
	}
	userClaims, ok := claims.(*auth.Claims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user claims"})
		return
	}

	// 1. Ambil data pekerjaan yang ada dari database
	existingJob, err := h.Sp2dkJobRepo.GetSp2dkJobByID(id, userClaims.StaffID, userClaims.IsAdmin)
	if err != nil {
		// INI AKAN MENANGKAP ERROR "NOT FOUND" DARI REPOSITORY
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "SP2DK job not found or access denied"})
			return
		}
		// Menangani error database lainnya
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve job for update: " + err.Error()})
		return
	}

	// 2. Ambil semua data dari form yang dikirim
	overallStatusForm := c.PostForm("overall_status")
	assignedPicStaffSigmaIDForm := c.PostForm("assigned_pic_staff_sigma_id")
	contractNoForm := c.PostForm("contract_no")
	sp2dkNoForm := c.PostForm("sp2dk_no")
	bap2dkNoForm := c.PostForm("bap2dk_no")
	contractDateForm := c.PostForm("contract_date")
	sp2dkDateForm := c.PostForm("sp2dk_date")
	bap2dkDateForm := c.PostForm("bap2dk_date")
	paymentDateForm := c.PostForm("payment_date")
	reportDateForm := c.PostForm("report_date")

	// 3. Simpan status lama sebelum diubah
	statusSebelumnya := existingJob.OverallStatus

	// 4. Terapkan semua perubahan dari form ke objek 'existingJob'
	if overallStatusForm != "" {
		existingJob.OverallStatus = overallStatusForm
	}
	if assignedPicStaffSigmaIDForm != "" {
		staff, err := h.StaffRepo.GetStaffByID(assignedPicStaffSigmaIDForm)
		if err != nil || staff == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "New Assigned PIC Staff ID not found"})
			return
		}
		existingJob.AssignedPicStaffSigmaID = assignedPicStaffSigmaIDForm
	}
	if contractNoForm != "" { existingJob.ContractNo = contractNoForm }
	if sp2dkNoForm != "" { existingJob.Sp2dkNo = sp2dkNoForm }
	if bap2dkNoForm != "" { existingJob.Bap2dkNo = bap2dkNoForm }
	if contractDateForm != "" {
		parsedDate, err := time.Parse("2006-01-02", contractDateForm)
		if err == nil { existingJob.ContractDate = &parsedDate }
	}
	if sp2dkDateForm != "" {
		parsedDate, err := time.Parse("2006-01-02", sp2dkDateForm)
		if err == nil { existingJob.Sp2dkDate = &parsedDate }
	}
	if bap2dkDateForm != "" {
		parsedDate, err := time.Parse("2006-01-02", bap2dkDateForm)
		if err == nil { existingJob.Bap2dkDate = &parsedDate }
	}
	if paymentDateForm != "" {
		parsedDate, err := time.Parse("2006-01-02", paymentDateForm)
		if err == nil { existingJob.PaymentDate = &parsedDate }
	}
	if reportDateForm != "" {
		parsedDate, err := time.Parse("2006-01-02", reportDateForm)
		if err == nil { existingJob.ReportDate = &parsedDate }
	}

	// 5. Proses file upload (jika ada)
	file, err := c.FormFile("proof_of_work_pdf")
	if err == nil {
		filename := fmt.Sprintf("%s.pdf", existingJob.JobID)
		filePath := filepath.Join("uploads", filename)
		if err := c.SaveUploadedFile(file, filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save proof of work PDF: " + err.Error()})
			return
		}
		url := fmt.Sprintf("/uploads/%s", filename)
		existingJob.ProofOfWorkURL = &url
	}

	// 6. Panggil pemicu invoice JIKA status diubah menjadi "Selesai"
	if overallStatusForm == "Selesai" && statusSebelumnya != "Selesai" {
		log.Printf("INFO: Status pekerjaan SP2DK %s diubah menjadi Selesai. Memicu pembuatan invoice...", existingJob.JobID)
		_, err := h.InvoiceService.CreateInvoiceFromJob(
			existingJob.JobID, "SP2DK", existingJob.ClientID, existingJob.AssignedPicStaffSigmaID,
		)
		if err != nil {
			log.Printf("PERINGATAN: Gagal membuat invoice otomatis untuk pekerjaan SP2DK %s: %v", existingJob.JobID, err)
		} else {
			log.Printf("INFO: Invoice berhasil dibuat untuk pekerjaan SP2DK %s.", existingJob.JobID)
		}
	}
	
	// 7. Simpan SEMUA perubahan ke database
	log.Printf("DEBUG: Menyimpan final -> JobID: %s, Status: '%s'", existingJob.JobID, existingJob.OverallStatus)
	if err := h.Sp2dkJobRepo.UpdateSp2dkJob(existingJob); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update SP2DK job: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, existingJob)
}

// DeleteSp2dkJob handles deleting an SP2DK job by ID
func (h *Sp2dkJobHandler) DeleteSp2dkJob(c *gin.Context) {
	id := c.Param("id")

	claims, exists := c.Get("user_claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User claims not found in context"})
		return
	}
	userClaims, ok := claims.(*auth.Claims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user claims"})
		return
	}

	existingJob, err := h.Sp2dkJobRepo.GetSp2dkJobByID(id, userClaims.StaffID, userClaims.IsAdmin) // Filter by access
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "SP2DK job not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve job for delete: " + err.Error()})
		return
	}
	if existingJob == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "SP2DK job not found"})
		return
	}

	if err := h.Sp2dkJobRepo.DeleteSp2dkJob(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete SP2DK job: " + err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
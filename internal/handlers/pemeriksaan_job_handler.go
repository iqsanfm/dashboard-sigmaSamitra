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

// PemeriksaanJobHandler handles HTTP requests for Pemeriksaan job operations
type PemeriksaanJobHandler struct {
	PemeriksaanJobRepo repositories.PemeriksaanJobRepository
	ClientRepo         repositories.ClientRepository
	StaffRepo          repositories.StaffRepository
	InvoiceService 		 services.InvoiceService
}

// NewPemeriksaanJobHandler creates a new PemeriksaanJobHandler
func NewPemeriksaanJobHandler(pjRepo repositories.PemeriksaanJobRepository, cRepo repositories.ClientRepository, sRepo repositories.StaffRepository, invService services.InvoiceService) *PemeriksaanJobHandler {
	return &PemeriksaanJobHandler{
		PemeriksaanJobRepo: pjRepo,
		ClientRepo:         cRepo,
		StaffRepo:          sRepo,
		InvoiceService: invService,
	}
}

// CreatePemeriksaanJob handles the creation of a new Pemeriksaan job
func (h *PemeriksaanJobHandler) CreatePemeriksaanJob(c *gin.Context) {
	var req models.NewPemeriksaanJobRequest
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

	pemeriksaanJob := &models.PemeriksaanJob{
		ClientID:              req.ClientID,
		AssignedPicStaffSigmaID: req.AssignedPicStaffSigmaID,
		ContractNo:            req.ContractNo,
		ContractDate:          req.ContractDate,
		Sp2No:                 req.Sp2No,
		Sp2Date:               req.Sp2Date,
		SkpNo:                 req.SkpNo,
		SkpDate:               req.SkpDate,
		OverallStatus:         req.OverallStatus,
	}

	if err := h.PemeriksaanJobRepo.CreatePemeriksaanJob(pemeriksaanJob); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Pemeriksaan job: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, pemeriksaanJob)
}

// GetAllPemeriksaanJobs fetches all Pemeriksaan jobs, filtered by PIC if not admin
func (h *PemeriksaanJobHandler) GetAllPemeriksaanJobs(c *gin.Context) {

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
	jobs, err := h.PemeriksaanJobRepo.GetAllPemeriksaanJobs(userClaims.StaffID, userClaims.IsAdmin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve Pemeriksaan jobs: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, jobs)
}

// GetPemeriksaanJobByID fetches a single Pemeriksaan job by ID, filtered by PIC if not admin
func (h *PemeriksaanJobHandler) GetPemeriksaanJobByID(c *gin.Context) {
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

	job, err := h.PemeriksaanJobRepo.GetPemeriksaanJobByID(id, userClaims.StaffID, userClaims.IsAdmin)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Pemeriksaan job not found or access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve Pemeriksaan job: " + err.Error()})
		return
	}
	if job == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pemeriksaan job not found or access denied"})
		return
	}
	c.JSON(http.StatusOK, job)
}

// UpdatePemeriksaanJob handles partial updates to a Pemeriksaan job's main fields
// internal/handlers/pemeriksaan_job_handler.go

// internal/handlers/pemeriksaan_job_handler.go

func (h *PemeriksaanJobHandler) UpdatePemeriksaanJob(c *gin.Context) {
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
	existingJob, err := h.PemeriksaanJobRepo.GetPemeriksaanJobByID(id, userClaims.StaffID, userClaims.IsAdmin)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Pemeriksaan job not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve job for update: " + err.Error()})
		return
	}

	// 2. Ambil semua data dari form yang dikirim
	// ================== PERUBAHAN NAMA FIELD DIMULAI DARI SINI ==================
	overallStatusForm := c.PostForm("overall_status") // <-- Gunakan nama field standar
	assignedPicStaffSigmaIDForm := c.PostForm("assigned_pic_staff_sigma_id")
	contractNoForm := c.PostForm("contract_no")
	sp2NoForm := c.PostForm("sp2_no")
	skpNoForm := c.PostForm("skp_no")
	contractDateForm := c.PostForm("contract_date")
	sp2DateForm := c.PostForm("sp2_date")
	skpDateForm := c.PostForm("skp_date")

	// 3. Terapkan semua perubahan dari form ke objek 'existingJob'
	// Terapkan status baru ke field yang sudah distandarisasi
	if overallStatusForm != "" {
		existingJob.OverallStatus = overallStatusForm // <-- Gunakan field standar
	}

	// Terapkan update untuk field lain...
	if assignedPicStaffSigmaIDForm != "" {
		staff, err := h.StaffRepo.GetStaffByID(assignedPicStaffSigmaIDForm)
		if err != nil || staff == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "New Assigned PIC Staff ID not found"})
			return
		}
		existingJob.AssignedPicStaffSigmaID = assignedPicStaffSigmaIDForm
	}
	if contractNoForm != "" { existingJob.ContractNo = contractNoForm }
	if sp2NoForm != "" { existingJob.Sp2No = sp2NoForm }
	if skpNoForm != "" { existingJob.SkpNo = skpNoForm }

	// Terapkan tanggal-tanggal baru
	if contractDateForm != "" {
		parsedDate, err := time.Parse("2006-01-02", contractDateForm)
		if err == nil { existingJob.ContractDate = &parsedDate }
	}
	if sp2DateForm != "" {
		parsedDate, err := time.Parse("2006-01-02", sp2DateForm)
		if err == nil { existingJob.Sp2Date = &parsedDate }
	}
	if skpDateForm != "" {
		parsedDate, err := time.Parse("2006-01-02", skpDateForm)
		if err == nil { existingJob.SkpDate = &parsedDate }
	}
	
	// 4. Proses file upload (jika ada)
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

	// 5. Panggil pemicu invoice JIKA status diubah menjadi "Selesai"
	// Gunakan field dan variabel yang sudah distandarisasi
	if overallStatusForm == "Selesai" && existingJob.OverallStatus != "Selesai" {
		log.Printf("INFO: Status pekerjaan %s diubah menjadi Selesai. Memicu pembuatan invoice...", existingJob.JobID)
		_, err := h.InvoiceService.CreateInvoiceFromJob(
			existingJob.JobID, "Pemeriksaan", existingJob.ClientID, existingJob.AssignedPicStaffSigmaID,
		)
		if err != nil {
			log.Printf("PERINGATAN: Gagal membuat invoice otomatis untuk pekerjaan %s: %v", existingJob.JobID, err)
		} else {
			log.Printf("INFO: Invoice berhasil dibuat untuk pekerjaan %s.", existingJob.JobID)
		}
	}
	
	// 6. Simpan SEMUA perubahan ke database
	log.Printf("DEBUG: Menyimpan final -> JobID: %s, Status: '%s'", existingJob.JobID, existingJob.OverallStatus)
	if err := h.PemeriksaanJobRepo.UpdatePemeriksaanJob(existingJob); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update Pemeriksaan job: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, existingJob)
}

// DeletePemeriksaanJob handles deleting a Pemeriksaan job by ID
func (h *PemeriksaanJobHandler) DeletePemeriksaanJob(c *gin.Context) {
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

	existingJob, err := h.PemeriksaanJobRepo.GetPemeriksaanJobByID(id, userClaims.StaffID, userClaims.IsAdmin) // Filter by access
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Pemeriksaan job not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve job for delete: " + err.Error()})
		return
	}
	if existingJob == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pemeriksaan job not found"})
		return
	}

	if err := h.PemeriksaanJobRepo.DeletePemeriksaanJob(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete Pemeriksaan job: " + err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
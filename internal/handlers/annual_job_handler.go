package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/iqsanfm/dashboard-pekerjaan-backend/internal/models"       // Pastikan ini modul Anda
	"github.com/iqsanfm/dashboard-pekerjaan-backend/internal/repositories" // Pastikan ini modul Anda
)

// AnnualJobHandler handles HTTP requests for annual job operations
type AnnualJobHandler struct {
	AnnualJobRepo repositories.AnnualJobRepository
	ClientRepo    repositories.ClientRepository
	StaffRepo     repositories.StaffRepository
}

// NewAnnualJobHandler creates a new AnnualJobHandler
func NewAnnualJobHandler(ajRepo repositories.AnnualJobRepository, cRepo repositories.ClientRepository, sRepo repositories.StaffRepository) *AnnualJobHandler {
	return &AnnualJobHandler{
		AnnualJobRepo: ajRepo,
		ClientRepo:    cRepo,
		StaffRepo:     sRepo,
	}
}

// CreateAnnualJob handles the creation of a new annual job and its initial reports
func (h *AnnualJobHandler) CreateAnnualJob(c *gin.Context) {
	var req models.NewAnnualJobRequest
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

	annualJob := &models.AnnualJob{
		ClientID:              req.ClientID,
		JobYear:               req.JobYear,
		AssignedPicStaffSigmaID: req.AssignedPicStaffSigmaID,
		OverallStatus:         req.OverallStatus,
	}

	// Handle initial Annual Tax Report (if provided)
	if req.TaxReport != nil {
		annualJob.TaxReports = []models.AnnualTaxReport{
			{
				BillingCode:   req.TaxReport.BillingCode,
				PaymentDate:   req.TaxReport.PaymentDate,
				PaymentAmount: req.TaxReport.PaymentAmount,
				ReportDate:    req.TaxReport.ReportDate,
				ReportStatus:  req.TaxReport.ReportStatus,
			},
		}
	}

	// Handle initial Annual Dividend Report (if provided)
	if req.DividendReport != nil {
		// Validasi: Jika is_reported true, report_date tidak boleh null
		if req.DividendReport.IsReported && req.DividendReport.ReportDate == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ReportDate is required for reported dividend reports"})
			return
		}
		annualJob.DividendReports = []models.AnnualDividendReport{
			{
				IsReported:   req.DividendReport.IsReported,
				ReportDate:   req.DividendReport.ReportDate,
				ReportStatus: req.DividendReport.ReportStatus,
			},
		}
	}

	if err := h.AnnualJobRepo.CreateAnnualJob(annualJob); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create annual job: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, annualJob)
}

// GetAllAnnualJobs fetches all annual jobs, filtered by PIC if not admin
func (h *AnnualJobHandler) GetAllAnnualJobs(c *gin.Context) {
	staffID, exists := c.Get("staffID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Staff ID not found in context"})
		return
	}
	role, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Role not found in context"})
		return
	}

	isAdmin := (role.(string) == "admin")
	staffIDStr := staffID.(string)

	jobs, err := h.AnnualJobRepo.GetAllAnnualJobs(staffIDStr, isAdmin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve annual jobs: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, jobs)
}

// GetAnnualJobByID fetches a single annual job by ID, filtered by PIC if not admin
func (h *AnnualJobHandler) GetAnnualJobByID(c *gin.Context) {
	id := c.Param("id")

	staffID, exists := c.Get("staffID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Staff ID not found in context"})
		return
	}
	role, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Role not found in context"})
		return
	}

	isAdmin := (role.(string) == "admin")
	staffIDStr := staffID.(string)

	job, err := h.AnnualJobRepo.GetAnnualJobByID(id, staffIDStr, isAdmin)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Annual job not found or access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve annual job: " + err.Error()})
		return
	}
	if job == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Annual job not found or access denied"})
		return
	}
	c.JSON(http.StatusOK, job)
}

// UpdateAnnualJob handles partial updates to an annual job's main fields
func (h *AnnualJobHandler) UpdateAnnualJob(c *gin.Context) {
	id := c.Param("id")

	staffID, exists := c.Get("staffID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Staff ID not found in context"})
		return
	}
	role, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Role not found in context"})
		return
	}
	isAdmin := (role.(string) == "admin")
	staffIDStr := staffID.(string)

	existingJob, err := h.AnnualJobRepo.GetAnnualJobByID(id, staffIDStr, isAdmin) // Filter by access
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Annual job not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve job for update: " + err.Error()})
		return
	}
	if existingJob == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Annual job not found"})
		return
	}

	// --- PENANGANAN MULTIPART FORM-DATA (FILE DAN FIELD LAINNYA) ---
	// Mendapatkan nilai field dari form (bukan JSON)
	overallStatusForm := c.PostForm("overall_status")
	assignedPicStaffSigmaIDForm := c.PostForm("assigned_pic_staff_sigma_id")
	jobYearForm := c.PostForm("job_year") // Pastikan Anda parsing int jika digunakan

	// Mendapatkan file yang diupload (bukti PDF)
	file, err := c.FormFile("proof_of_work_pdf")
	fileReceived := (err == nil && file != nil) // True jika file diterima

	// Validasi: Jika status berubah menjadi "Selesai", file PDF harus ada
	if overallStatusForm != "" && overallStatusForm == "Selesai" {
		if !fileReceived {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Proof of work PDF is required when setting status to 'Selesai'"})
			return
		}
		// Validasi tipe file (opsional tapi disarankan)
		if file.Header.Get("Content-Type") != "application/pdf" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Uploaded file must be a PDF"})
			return
		}
	} else if overallStatusForm != "" && overallStatusForm != "Selesai" && fileReceived {
        // Jika status BUKAN Selesai, tapi ada file diupload, tolak atau ingatkan
        c.JSON(http.StatusBadRequest, gin.H{"error": "Proof of work PDF can only be uploaded when status is 'Selesai'"})
        return
    }

	// Memproses file upload jika ada
	var uploadedFilePath *string
	if fileReceived {
		// Buat nama file unik (misalnya JobID.pdf)
		filename := fmt.Sprintf("%s.pdf", existingJob.JobID)
		filePath := filepath.Join("uploads", filename) // Simpan di folder 'uploads'

		// Pastikan folder 'uploads' ada
		if _, err := os.Stat("uploads"); os.IsNotExist(err) {
			os.Mkdir("uploads", os.ModePerm)
		}

		// Simpan file
		if err := c.SaveUploadedFile(file, filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save proof of work PDF: " + err.Error()})
			return
		}
		// URL yang akan disimpan di DB dan dikembalikan ke client
		// Asumsi server diakses melalui localhost:8080 dan folder uploads dilayani di /uploads
		url := fmt.Sprintf("/uploads/%s", filename)
		uploadedFilePath = &url
	}

	// --- TERAPKAN UPDATE KE existingJob BERDASARKAN FORM FIELD ---
	if overallStatusForm != "" {
		existingJob.OverallStatus = overallStatusForm
	}
	if assignedPicStaffSigmaIDForm != "" {
		// Validasi PIC Staff ID baru jika disediakan
		staff, err := h.StaffRepo.GetStaffByID(assignedPicStaffSigmaIDForm)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate new Assigned PIC Staff ID: " + err.Error()})
			return
		}
		if staff == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "New Assigned PIC Staff ID not found"})
			return
		}
		existingJob.AssignedPicStaffSigmaID = assignedPicStaffSigmaIDForm
	}
	// Parsing job_year
	if jobYearForm != "" {
		year, err := strconv.Atoi(jobYearForm)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job_year format"})
			return
		}
		existingJob.JobYear = year
	}
	// Jika file diupload, update ProofOfWorkURL di job
	if uploadedFilePath != nil {
		existingJob.ProofOfWorkURL = uploadedFilePath
	}

	if err := h.AnnualJobRepo.UpdateAnnualJob(existingJob); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update annual job: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, existingJob)
}

// CreateAnnualTaxReport handles adding a new SPT Tahunan report to an existing annual job
func (h *AnnualJobHandler) CreateAnnualTaxReport(c *gin.Context) {
	jobID := c.Param("id") // Annual Job ID from route
	var req models.NewAnnualTaxReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	staffID, exists := c.Get("staffID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Staff ID not found in context"})
		return
	}
	role, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Role not found in context"})
		return
	}
	isAdmin := (role.(string) == "admin")
	staffIDStr := staffID.(string)

	// Validate annual job existence AND user access
	job, err := h.AnnualJobRepo.GetAnnualJobByID(jobID, staffIDStr, isAdmin)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Annual job not found for tax report or access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check annual job existence: " + err.Error()})
		return
	}
	if job == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Annual job not found for tax report or access denied"})
		return
	}
	// Check if a tax report already exists for this job (based on UNIQUE constraint)
	if len(job.TaxReports) > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "An annual tax report already exists for this job"})
		return
	}


	annualTaxReport := &models.AnnualTaxReport{
		JobID:         jobID,
		BillingCode:   req.BillingCode,
		PaymentDate:   req.PaymentDate,
		PaymentAmount: req.PaymentAmount,
		ReportDate:    req.ReportDate,
		ReportStatus:  req.ReportStatus,
	}

	if err := h.AnnualJobRepo.CreateAnnualTaxReport(annualTaxReport); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create annual tax report: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, annualTaxReport)
}

// UpdateAnnualTaxReport handles updating a specific SPT Tahunan report
func (h *AnnualJobHandler) UpdateAnnualTaxReport(c *gin.Context) {
	reportID := c.Param("report_id")
	jobID := c.Param("id") // Ensure jobID is retrieved if needed for more complex validation

	var req models.UpdateAnnualTaxReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	staffID, exists := c.Get("staffID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Staff ID not found in context"})
		return
	}
	role, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Role not found in context"})
		return
	}
	isAdmin := (role.(string) == "admin")
	staffIDStr := staffID.(string)

	existingReport, err := h.AnnualJobRepo.GetAnnualTaxReportByID(reportID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Annual tax report not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve annual tax report for update: " + err.Error()})
		return
	}
	if existingReport == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Annual tax report not found"})
		return
	}

	 if existingReport.JobID != jobID {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Tax report ID does not match the annual job ID in the URL"})
        return
    }

	// Validate access to the parent annual job
	jobForReport, err := h.AnnualJobRepo.GetAnnualJobByID(existingReport.JobID, staffIDStr, isAdmin)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Annual job for this tax report not found or access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check access to annual job for tax report: " + err.Error()})
		return
	}
	if jobForReport == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Annual job for this tax report not found or access denied"})
		return
	}


	// Apply updates
	if req.BillingCode != nil {
		existingReport.BillingCode = *req.BillingCode
	}
	if req.PaymentDate != nil {
		existingReport.PaymentDate = req.PaymentDate
	}
	if req.PaymentAmount != nil {
		existingReport.PaymentAmount = req.PaymentAmount
	}
	if req.ReportDate != nil {
		existingReport.ReportDate = req.ReportDate
	}
	if req.ReportStatus != nil {
		existingReport.ReportStatus = *req.ReportStatus
	}

	if err := h.AnnualJobRepo.UpdateAnnualTaxReport(existingReport); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update annual tax report: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, existingReport)
}

// DeleteAnnualTaxReport handles deleting a specific SPT Tahunan report
func (h *AnnualJobHandler) DeleteAnnualTaxReport(c *gin.Context) {
	reportID := c.Param("report_id")
	jobID := c.Param("id") // Annual Job ID from route, for potential access check

	staffID, exists := c.Get("staffID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Staff ID not found in context"})
		return
	}
	role, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Role not found in context"})
		return
	}
	isAdmin := (role.(string) == "admin")
	staffIDStr := staffID.(string)

	existingReport, err := h.AnnualJobRepo.GetAnnualTaxReportByID(reportID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Annual tax report not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve annual tax report for delete: " + err.Error()})
		return
	}
	if existingReport == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Annual tax report not found"})
		return
	}

	  if existingReport.JobID != jobID {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Tax report ID does not match the annual job ID in the URL"})
        return
    }

	// Validate access to the parent annual job
	jobForReport, err := h.AnnualJobRepo.GetAnnualJobByID(existingReport.JobID, staffIDStr, isAdmin)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Annual job for this tax report not found or access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check access to annual job for tax report: " + err.Error()})
		return
	}
	if jobForReport == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Annual job for this tax report not found or access denied"})
		return
	}


	if err := h.AnnualJobRepo.DeleteAnnualTaxReport(reportID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete annual tax report: " + err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// CreateAnnualDividendReport handles adding a new Investasi Dividen report to an existing annual job
func (h *AnnualJobHandler) CreateAnnualDividendReport(c *gin.Context) {
	jobID := c.Param("id") // Annual Job ID from route
	var req models.NewAnnualDividendReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validasi: Jika is_reported true, report_date tidak boleh null
	if req.IsReported && req.ReportDate == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ReportDate is required for reported dividend reports"})
		return
	}

	staffID, exists := c.Get("staffID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Staff ID not found in context"})
		return
	}
	role, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Role not found in context"})
		return
	}
	isAdmin := (role.(string) == "admin")
	staffIDStr := staffID.(string)

	// Validate annual job existence AND user access
	job, err := h.AnnualJobRepo.GetAnnualJobByID(jobID, staffIDStr, isAdmin)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Annual job not found for dividend report or access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check annual job existence: " + err.Error()})
		return
	}
	if job == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Annual job not found for dividend report or access denied"})
		return
	}
	// Check if a dividend report already exists for this job (based on UNIQUE constraint)
	if len(job.DividendReports) > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "An annual dividend report already exists for this job"})
		return
	}

	annualDividendReport := &models.AnnualDividendReport{
		JobID:        jobID,
		IsReported:   req.IsReported,
		ReportDate:   req.ReportDate,
		ReportStatus: req.ReportStatus,
	}

	if err := h.AnnualJobRepo.CreateAnnualDividendReport(annualDividendReport); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create annual dividend report: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, annualDividendReport)
}

// UpdateAnnualDividendReport handles updating a specific Investasi Dividen report
func (h *AnnualJobHandler) UpdateAnnualDividendReport(c *gin.Context) {
	reportID := c.Param("report_id")
	jobID := c.Param("id") // For potential access check

	var req models.UpdateAnnualDividendReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validasi: Jika is_reported di request true, dan report_date di request null
	if req.IsReported != nil && *req.IsReported && req.ReportDate == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ReportDate is required when setting is_reported to true"})
		return
	}
	// Tambahan validasi: Jika is_reported di request true, tapi report_date di request tidak di provide
	// (misalnya client mengirim {"is_reported": true} tapi tidak ada "report_date")
	// Ini butuh fetch existingReport dulu
	if req.IsReported != nil && *req.IsReported && req.ReportDate == nil {
		existingReport, err := h.AnnualJobRepo.GetAnnualDividendReportByID(reportID)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Annual dividend report not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve annual dividend report: " + err.Error()})
			return
		}
		if existingReport.ReportDate == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ReportDate is required when setting is_reported to true and no date is provided"})
			return
		}
	}


	staffID, exists := c.Get("staffID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Staff ID not found in context"})
		return
	}
	role, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Role not found in context"})
		return
	}
	isAdmin := (role.(string) == "admin")
	staffIDStr := staffID.(string)

	existingReport, err := h.AnnualJobRepo.GetAnnualDividendReportByID(reportID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Annual dividend report not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve annual dividend report for update: " + err.Error()})
		return
	}
	if existingReport == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Annual dividend report not found"})
		return
	}

	   if existingReport.JobID != jobID {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Dividend report ID does not match the annual job ID in the URL"})
        return
    }

	// Validate access to the parent annual job
	jobForReport, err := h.AnnualJobRepo.GetAnnualJobByID(existingReport.JobID, staffIDStr, isAdmin)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Annual job for this dividend report not found or access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check access to annual job for dividend report: " + err.Error()})
		return
	}
	if jobForReport == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Annual job for this dividend report not found or access denied"})
		return
	}

	// Apply updates
	if req.IsReported != nil {
		existingReport.IsReported = *req.IsReported
	}
	if req.ReportDate != nil {
		existingReport.ReportDate = req.ReportDate
	}
	if req.ReportStatus != nil {
		existingReport.ReportStatus = *req.ReportStatus
	}

	if err := h.AnnualJobRepo.UpdateAnnualDividendReport(existingReport); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update annual dividend report: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, existingReport)
}

// DeleteAnnualDividendReport handles deleting a specific Investasi Dividen report
func (h *AnnualJobHandler) DeleteAnnualDividendReport(c *gin.Context) {
	reportID := c.Param("report_id")
	jobID := c.Param("id") // For potential access check

	staffID, exists := c.Get("staffID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Staff ID not found in context"})
		return
	}
	role, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Role not found in context"})
		return
	}
	isAdmin := (role.(string) == "admin")
	staffIDStr := staffID.(string)

	existingReport, err := h.AnnualJobRepo.GetAnnualDividendReportByID(reportID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Annual dividend report not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve annual dividend report for delete: " + err.Error()})
		return
	}
	if existingReport == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Annual dividend report not found"})
		return
	}
	
	  if existingReport.JobID != jobID {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Dividend report ID does not match the annual job ID in the URL"})
        return
    }

	// Validate access to the parent annual job
	jobForReport, err := h.AnnualJobRepo.GetAnnualJobByID(existingReport.JobID, staffIDStr, isAdmin)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Annual job for this dividend report not found or access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check access to annual job for dividend report: " + err.Error()})
		return
	}
	if jobForReport == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Annual job for this dividend report not found or access denied"})
		return
	}

	if err := h.AnnualJobRepo.DeleteAnnualDividendReport(reportID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete annual dividend report: " + err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
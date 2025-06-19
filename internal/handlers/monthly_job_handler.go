package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	// For time.Parse and time.Time pointers
	"github.com/gin-gonic/gin"
	"github.com/iqsanfm/dashboard-pekerjaan-backend/internal/models"
	"github.com/iqsanfm/dashboard-pekerjaan-backend/internal/repositories"
	"github.com/iqsanfm/dashboard-pekerjaan-backend/internal/services"
	"github.com/iqsanfm/dashboard-pekerjaan-backend/pkg/auth"
)

// MonthlyJobHandler handles HTTP requests for monthly job operations
type MonthlyJobHandler struct {
	MonthlyJobRepo repositories.MonthlyJobRepository
	ClientRepo     repositories.ClientRepository // Need client repo to validate client_id
	StaffRepo      repositories.StaffRepository
	InvoiceService services.InvoiceService
}

// NewMonthlyJobHandler creates a new MonthlyJobHandler
func NewMonthlyJobHandler(mjRepo repositories.MonthlyJobRepository, cRepo repositories.ClientRepository, sRepo repositories.StaffRepository, invService services.InvoiceService,) *MonthlyJobHandler {
	return &MonthlyJobHandler{
		MonthlyJobRepo: mjRepo,
		ClientRepo:     cRepo,
		InvoiceService: invService,
	}
}

// CreateMonthlyJob handles the creation of a new monthly job and its initial tax reports
func (h *MonthlyJobHandler) CreateMonthlyJob(c *gin.Context) {
	var req models.NewMonthlyJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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

	// Basic validation: Check if client_id exists
	client, err := h.ClientRepo.GetClientByID(req.ClientID, userClaims.StaffID, userClaims.IsAdmin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate client ID"})
		return
	}
	if client == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Client ID not found"})
		return
	}

	monthlyJob := &models.MonthlyJob{
		ClientID:              req.ClientID,
		JobMonth:              req.JobMonth,
		JobYear:               req.JobYear,
		AssignedPicStaffSigmaID: req.AssignedPicStaffSigmaID,
		OverallStatus:         req.OverallStatus,
	}

	// Convert NewMonthlyTaxReportRequest to MonthlyTaxReport for the repository
	for _, trReq := range req.TaxReports {
		monthlyJob.TaxReports = append(monthlyJob.TaxReports, models.MonthlyTaxReport{
			TaxType:       trReq.TaxType,
			BillingCode:   trReq.BillingCode,
			PaymentDate:   trReq.PaymentDate,
			PaymentAmount: trReq.PaymentAmount,
			ReportStatus:  trReq.ReportStatus,
			ReportDate:    trReq.ReportDate,
		})
	}

	if err := h.MonthlyJobRepo.CreateMonthlyJob(monthlyJob); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create monthly job: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, monthlyJob)
}



// GetAllMonthlyJobs fetches all monthly jobs with their associated client and tax reports
func (h *MonthlyJobHandler) GetAllMonthlyJobs(c *gin.Context) {
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

	jobs, err := h.MonthlyJobRepo.GetAllMonthlyJobs(userClaims.StaffID, userClaims.IsAdmin) // <-- TERUSKAN PARAMETER FILTER
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve monthly jobs: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, jobs)
}

// GetMonthlyJobByID fetches a single monthly job by ID with associated client and tax reports
func (h *MonthlyJobHandler) GetMonthlyJobByID(c *gin.Context) {
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

	job, err := h.MonthlyJobRepo.GetMonthlyJobByID(id, userClaims.StaffID, userClaims.IsAdmin) // <-- TERUSKAN PARAMETER FILTER
	if err != nil {
		if err == sql.ErrNoRows { // Ini akan menangani not found (baik karena ID salah atau tidak punya akses)
			c.JSON(http.StatusNotFound, gin.H{"error": "Monthly job not found or access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve monthly job: " + err.Error()})
		return
	}
	if job == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Monthly job not found or access denied"})
		return
	}
	c.JSON(http.StatusOK, job)
}

// UpdateMonthlyJob handles partial updates to a monthly job's main fields
func (h *MonthlyJobHandler) UpdateMonthlyJob(c *gin.Context) {
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

	existingJob, err := h.MonthlyJobRepo.GetMonthlyJobByID(id, userClaims.StaffID, userClaims.IsAdmin)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Monthly job not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve job for update: " + err.Error()})
		return
	}
	if existingJob == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Monthly job not found"})
		return
		// Important: If job is not found, client.GetClientByID returns nil, sql.ErrNoRows, not a *client.
		// So checking `client == nil` after `if err != nil` is redundant for `sql.ErrNoRows` but good for other error types.
	}

	// --- PENANGANAN MULTIPART FORM-DATA (FILE DAN FIELD LAINNYA) ---
	// Mendapatkan nilai field dari form (bukan JSON)
	overallStatusForm := c.PostForm("overall_status")
	assignedPicStaffSigmaIDForm := c.PostForm("assigned_pic_staff_sigma_id")
	// ... (Tambahkan field lain yang bisa diupdate jika ada di UpdateMonthlyJobRequest) ...
	// Misalnya job_month, job_year, dll.
	jobMonthForm := c.PostForm("job_month")
	jobYearForm := c.PostForm("job_year")

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
	// Contoh parsing field lain (jika ada di UpdateMonthlyJobRequest)
	if jobMonthForm != "" {
		month, err := strconv.Atoi(jobMonthForm)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job_month format"})
			return
		}
		existingJob.JobMonth = month
	}
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

	if err := h.MonthlyJobRepo.UpdateMonthlyJob(existingJob); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update monthly job: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, existingJob)
}

// CreateMonthlyTaxReport handles adding a new tax report to an existing monthly job
func (h *MonthlyJobHandler) CreateMonthlyTaxReport(c *gin.Context) {
	jobID := c.Param("job_id")
	var req models.NewMonthlyTaxReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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

	// Optional: Validate if jobID exists before adding tax report
	_, err := h.MonthlyJobRepo.GetMonthlyJobByID(jobID, userClaims.StaffID, userClaims.IsAdmin)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Monthly job not found for tax report"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check monthly job existence: " + err.Error()})
		return
	}

	taxReport := &models.MonthlyTaxReport{
		JobID:         jobID,
		TaxType:       req.TaxType,
		BillingCode:   req.BillingCode,
		PaymentDate:   req.PaymentDate,
		PaymentAmount: req.PaymentAmount,
		ReportStatus:  req.ReportStatus,
		ReportDate:    req.ReportDate,
	}

	if err := h.MonthlyJobRepo.CreateMonthlyTaxReport(taxReport); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create monthly tax report: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, taxReport)
}

// UpdateMonthlyTaxReport handles updating a specific tax report for a monthly job
func (h *MonthlyJobHandler) UpdateMonthlyTaxReport(c *gin.Context) {
	reportID := c.Param("report_id") // ID of the specific tax report
	// jobID := c.Param("job_id") // job_id can be used for extra validation if needed

	var req models.UpdateMonthlyTaxReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Fetch existing report to get current values (especially JobID)
	// Note: We don't have a GetMonthlyTaxReportByID in repo yet, so this might need adjustment
	// For now, let's assume direct update or add GetMonthlyTaxReportByID to repo if needed.
	// For simplicity, let's just create a new MonthlyTaxReport object and set its ID
	// If you want to only update provided fields, you would fetch it first, then apply changes.
	
	// Temporarily: To update, we ideally need to fetch the existing report first.
	// Let's assume we can directly update based on reportID and provided fields
	// For a real application, you'd fetch the existing report, then apply partial updates.
	// For now, we'll construct a model from the request + ID and pass it.
	
	// This part needs `GetMonthlyTaxReportByID` in the repository for proper partial update.
	// For simplicity, we'll assume the request provides all fields for the update.
	// Better approach for PATCH:
    // existingReport, err := h.MonthlyJobRepo.GetMonthlyTaxReportByID(reportID) // This function is not yet in repo
    // if err != nil { ... handle not found ... }
    // if req.TaxType != nil { existingReport.TaxType = *req.TaxType }
    // ... apply other field updates ...
    // then call repo.UpdateMonthlyTaxReport(existingReport)

	// For now, a simpler (but less robust for PATCH) approach assuming full model is sent for update.
	// Let's modify this to reflect actual partial update logic more accurately.
	
    // Fetch the existing report
    // (This requires adding GetMonthlyTaxReportByID to MonthlyJobRepository interface and implementation)
    // For now, assuming direct update where all fields are sent, or client sends ID.
    // If we want a true partial update, we must get the existing record first.
    
    // As GetMonthlyTaxReportByID is not yet in repo, we'll modify the Update logic
    // to build a report with the ID and only update the fields that are not nil.
    // This requires a `GetMonthlyTaxReportByID` method in your repository.
    // Let's temporarily assume we have a way to fetch it, or simplify.
    
    // For a robust PATCH for tax reports, you should add:
    // GetMonthlyTaxReportByID(reportID string) (*models.MonthlyTaxReport, error)
    // in your MonthlyJobRepository.
    
    // For this example, let's just construct the update and assume the repo handles missing fields
    // or that the client sends all necessary fields for the update based on UpdateMonthlyTaxReportRequest.
    // A more precise PATCH handler for `MonthlyTaxReport` would first fetch it.

    // Let's add a placeholder for fetching the existing report
    // For now, we'll just try to update. If GetMonthlyTaxReportByID is added later,
    // this handler can be refined.

    // A simpler approach for this example: directly update based on given fields and ID
    // This assumes the client sends all fields to be updated, or we use a `map[string]interface{}` approach.
    // Given UpdateMonthlyTaxReportRequest uses pointers, we should fetch first.

    // **Revised UpdateMonthlyTaxReport handler to be robust with PATCH:**
    existingReport, err := h.MonthlyJobRepo.GetMonthlyTaxReportByID(reportID) // Assumes this function exists now
    if err != nil {
        if err == sql.ErrNoRows {
            c.JSON(http.StatusNotFound, gin.H{"error": "Tax report not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tax report for update: " + err.Error()})
        return
    }
    if existingReport == nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Tax report not found"})
        return
    }

    // Apply updates only for fields that are provided in the request
    if req.TaxType != nil {
        existingReport.TaxType = *req.TaxType
    }
    if req.BillingCode != nil {
        existingReport.BillingCode = *req.BillingCode
    }
    if req.PaymentDate != nil {
        existingReport.PaymentDate = req.PaymentDate
    }
    if req.PaymentAmount != nil {
        existingReport.PaymentAmount = req.PaymentAmount
    }
    if req.ReportStatus != nil {
        existingReport.ReportStatus = *req.ReportStatus
    }
    if req.ReportDate != nil {
        existingReport.ReportDate = req.ReportDate
    }

	if err := h.MonthlyJobRepo.UpdateMonthlyTaxReport(existingReport); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update monthly tax report: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, existingReport)
}

// DeleteMonthlyTaxReport handles deleting a specific tax report
func (h *MonthlyJobHandler) DeleteMonthlyTaxReport(c *gin.Context) {
	reportID := c.Param("report_id")

	// Optional: Check if report exists before deleting
	_, err := h.MonthlyJobRepo.GetMonthlyTaxReportByID(reportID) // Assumes this function exists now
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tax report not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check tax report existence: " + err.Error()})
		return
	}

	if err := h.MonthlyJobRepo.DeleteMonthlyTaxReport(reportID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete monthly tax report: " + err.Error()})
		return
	}
	c.Status(http.StatusNoContent) // 204 No Content for successful deletion
}
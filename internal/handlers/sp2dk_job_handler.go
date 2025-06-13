package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iqsanfm/dashboard-pekerjaan-backend/internal/models"       // Pastikan ini modul Anda
	"github.com/iqsanfm/dashboard-pekerjaan-backend/internal/repositories" // Pastikan ini modul Anda
)

// Sp2dkJobHandler handles HTTP requests for SP2DK job operations
type Sp2dkJobHandler struct {
	Sp2dkJobRepo repositories.Sp2dkJobRepository
	ClientRepo   repositories.ClientRepository
	StaffRepo    repositories.StaffRepository
}

// NewSp2dkJobHandler creates a new Sp2dkJobHandler
func NewSp2dkJobHandler(sjRepo repositories.Sp2dkJobRepository, cRepo repositories.ClientRepository, sRepo repositories.StaffRepository) *Sp2dkJobHandler {
	return &Sp2dkJobHandler{
		Sp2dkJobRepo: sjRepo,
		ClientRepo:   cRepo,
		StaffRepo:    sRepo,
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
		JobStatus:             req.JobStatus,
	}

	if err := h.Sp2dkJobRepo.CreateSp2dkJob(sp2dkJob); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create SP2DK job: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, sp2dkJob)
}

// GetAllSp2dkJobs fetches all SP2DK jobs, filtered by PIC if not admin
func (h *Sp2dkJobHandler) GetAllSp2dkJobs(c *gin.Context) {
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

	jobs, err := h.Sp2dkJobRepo.GetAllSp2dkJobs(staffIDStr, isAdmin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve SP2DK jobs: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, jobs)
}

// GetSp2dkJobByID fetches a single SP2DK job by ID, filtered by PIC if not admin
func (h *Sp2dkJobHandler) GetSp2dkJobByID(c *gin.Context) {
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

	job, err := h.Sp2dkJobRepo.GetSp2dkJobByID(id, staffIDStr, isAdmin)
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

// UpdateSp2dkJob handles partial updates to an SP2DK job's main fields
func (h *Sp2dkJobHandler) UpdateSp2dkJob(c *gin.Context) {
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

	existingJob, err := h.Sp2dkJobRepo.GetSp2dkJobByID(id, staffIDStr, isAdmin)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "SP2DK job not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve job for update: " + err.Error()})
		return
	}
	if existingJob == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "SP2DK job not found"})
		return
	}

	// --- PENANGANAN MULTIPART FORM-DATA (FILE DAN FIELD LAINNYA) ---
	// Mendapatkan nilai field dari form (bukan JSON)
	jobStatusForm := c.PostForm("job_status")
	assignedPicStaffSigmaIDForm := c.PostForm("assigned_pic_staff_sigma_id")
	contractNoForm := c.PostForm("contract_no")
	sp2dkNoForm := c.PostForm("sp2dk_no")
	bap2dkNoForm := c.PostForm("bap2dk_no")
	
	// Tanggal dalam format string dari form, perlu di-parse ke time.Time
	contractDateForm := c.PostForm("contract_date")
	sp2dkDateForm := c.PostForm("sp2dk_date")
	bap2dkDateForm := c.PostForm("bap2dk_date")
	paymentDateForm := c.PostForm("payment_date")
	reportDateForm := c.PostForm("report_date")


	// Mendapatkan file yang diupload (bukti PDF)
	file, err := c.FormFile("proof_of_work_pdf")
	fileReceived := (err == nil && file != nil) // True jika file diterima

	// Validasi: Jika status berubah menjadi "Selesai", file PDF harus ada
	if jobStatusForm != "" && jobStatusForm == "Selesai" {
		if !fileReceived {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Proof of work PDF is required when setting status to 'Selesai'"})
			return
		}
		// Validasi tipe file (opsional tapi disarankan)
		if file.Header.Get("Content-Type") != "application/pdf" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Uploaded file must be a PDF"})
			return
		}
	} else if jobStatusForm != "" && jobStatusForm != "Selesai" && fileReceived {
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
		url := fmt.Sprintf("/uploads/%s", filename)
		uploadedFilePath = &url
	}

	// --- TERAPKAN UPDATE KE existingJob BERDASARKAN FORM FIELD ---
	if jobStatusForm != "" {
		existingJob.JobStatus = jobStatusForm
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
	if contractNoForm != "" {
		existingJob.ContractNo = contractNoForm
	}
	if sp2dkNoForm != "" {
		existingJob.Sp2dkNo = sp2dkNoForm
	}
	if bap2dkNoForm != "" {
		existingJob.Bap2dkNo = bap2dkNoForm
	}

	// Parsing Tanggal-tanggal dari form (format YYYY-MM-DD atau RFC3339 jika dari JS/frontend)
    // Contoh format: "2006-01-02T15:04:05Z" atau "2006-01-02"
    // Penting: Pastikan format ini konsisten dengan cara frontend mengirim tanggal
	if contractDateForm != "" {
		parsedDate, err := time.Parse("2006-01-02", contractDateForm) // Contoh: YYYY-MM-DD
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contract_date format"})
			return
		}
		existingJob.ContractDate = &parsedDate
	}
	if sp2dkDateForm != "" {
		parsedDate, err := time.Parse("2006-01-02", sp2dkDateForm)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sp2dk_date format"})
			return
		}
		existingJob.Sp2dkDate = &parsedDate
	}
	if bap2dkDateForm != "" {
		parsedDate, err := time.Parse("2006-01-02", bap2dkDateForm)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bap2dk_date format"})
			return
		}
		existingJob.Bap2dkDate = &parsedDate
	}
	if paymentDateForm != "" {
		parsedDate, err := time.Parse("2006-01-02", paymentDateForm)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment_date format"})
			return
		}
		existingJob.PaymentDate = &parsedDate
	}
	if reportDateForm != "" {
		parsedDate, err := time.Parse("2006-01-02", reportDateForm)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report_date format"})
			return
		}
		existingJob.ReportDate = &parsedDate
	}

	// Jika file diupload, update ProofOfWorkURL di job
	if uploadedFilePath != nil {
		existingJob.ProofOfWorkURL = uploadedFilePath
	}

	if err := h.Sp2dkJobRepo.UpdateSp2dkJob(existingJob); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update SP2DK job: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, existingJob)
}

// DeleteSp2dkJob handles deleting an SP2DK job by ID
func (h *Sp2dkJobHandler) DeleteSp2dkJob(c *gin.Context) {
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

	existingJob, err := h.Sp2dkJobRepo.GetSp2dkJobByID(id, staffIDStr, isAdmin) // Filter by access
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
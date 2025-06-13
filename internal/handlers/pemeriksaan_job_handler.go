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

// PemeriksaanJobHandler handles HTTP requests for Pemeriksaan job operations
type PemeriksaanJobHandler struct {
	PemeriksaanJobRepo repositories.PemeriksaanJobRepository
	ClientRepo         repositories.ClientRepository
	StaffRepo          repositories.StaffRepository
}

// NewPemeriksaanJobHandler creates a new PemeriksaanJobHandler
func NewPemeriksaanJobHandler(pjRepo repositories.PemeriksaanJobRepository, cRepo repositories.ClientRepository, sRepo repositories.StaffRepository) *PemeriksaanJobHandler {
	return &PemeriksaanJobHandler{
		PemeriksaanJobRepo: pjRepo,
		ClientRepo:         cRepo,
		StaffRepo:          sRepo,
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
		JobStatus:             req.JobStatus,
	}

	if err := h.PemeriksaanJobRepo.CreatePemeriksaanJob(pemeriksaanJob); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Pemeriksaan job: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, pemeriksaanJob)
}

// GetAllPemeriksaanJobs fetches all Pemeriksaan jobs, filtered by PIC if not admin
func (h *PemeriksaanJobHandler) GetAllPemeriksaanJobs(c *gin.Context) {
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

	jobs, err := h.PemeriksaanJobRepo.GetAllPemeriksaanJobs(staffIDStr, isAdmin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve Pemeriksaan jobs: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, jobs)
}

// GetPemeriksaanJobByID fetches a single Pemeriksaan job by ID, filtered by PIC if not admin
func (h *PemeriksaanJobHandler) GetPemeriksaanJobByID(c *gin.Context) {
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

	job, err := h.PemeriksaanJobRepo.GetPemeriksaanJobByID(id, staffIDStr, isAdmin)
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
func (h *PemeriksaanJobHandler) UpdatePemeriksaanJob(c *gin.Context) {
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

	existingJob, err := h.PemeriksaanJobRepo.GetPemeriksaanJobByID(id, staffIDStr, isAdmin) // Filter by access
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Pemeriksaan job not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve job for update: " + err.Error()})
		return
	}
	if existingJob == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pemeriksaan job not found"})
		return
	}

	// --- PENANGANAN MULTIPART FORM-DATA (FILE DAN FIELD LAINNYA) ---
	// Mendapatkan nilai field dari form (bukan JSON)
	jobStatusForm := c.PostForm("job_status")
	assignedPicStaffSigmaIDForm := c.PostForm("assigned_pic_staff_sigma_id")
	contractNoForm := c.PostForm("contract_no")
	sp2NoForm := c.PostForm("sp2_no")
	skpNoForm := c.PostForm("skp_no")
	
	// Tanggal dalam format string dari form, perlu di-parse ke time.Time
	contractDateForm := c.PostForm("contract_date")
	sp2DateForm := c.PostForm("sp2_date")
	skpDateForm := c.PostForm("skp_date")


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
	if sp2NoForm != "" {
		existingJob.Sp2No = sp2NoForm
	}
	if skpNoForm != "" {
		existingJob.SkpNo = skpNoForm
	}

	// Parsing Tanggal-tanggal dari form (format YYYY-MM-DD atau RFC3339 jika dari JS/frontend)
    // Penting: Pastikan format ini konsisten dengan cara frontend mengirim tanggal
	if contractDateForm != "" {
		parsedDate, err := time.Parse("2006-01-02", contractDateForm) // Contoh: YYYY-MM-DD
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contract_date format"})
			return
		}
		existingJob.ContractDate = &parsedDate
	}
	if sp2DateForm != "" {
		parsedDate, err := time.Parse("2006-01-02", sp2DateForm)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sp2_date format"})
			return
		}
		existingJob.Sp2Date = &parsedDate
	}
	if skpDateForm != "" {
		parsedDate, err := time.Parse("2006-01-02", skpDateForm)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid skp_date format"})
			return
		}
		existingJob.SkpDate = &parsedDate
	}

	// Jika file diupload, update ProofOfWorkURL di job
	if uploadedFilePath != nil {
		existingJob.ProofOfWorkURL = uploadedFilePath
	}

	if err := h.PemeriksaanJobRepo.UpdatePemeriksaanJob(existingJob); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update Pemeriksaan job: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, existingJob)
}

// DeletePemeriksaanJob handles deleting a Pemeriksaan job by ID
func (h *PemeriksaanJobHandler) DeletePemeriksaanJob(c *gin.Context) {
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

	existingJob, err := h.PemeriksaanJobRepo.GetPemeriksaanJobByID(id, staffIDStr, isAdmin) // Filter by access
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
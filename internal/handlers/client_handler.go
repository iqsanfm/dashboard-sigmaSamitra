package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iqsanfm/dashboard-pekerjaan-backend/internal/models"
	"github.com/iqsanfm/dashboard-pekerjaan-backend/internal/repositories"
	"github.com/iqsanfm/dashboard-pekerjaan-backend/pkg/auth"
	"github.com/iqsanfm/dashboard-pekerjaan-backend/pkg/utils" // For password hashing
)

// ClientHandler handles HTTP requests for client operations
type ClientHandler struct {
	ClientRepo repositories.ClientRepository
	StaffRepo  repositories.StaffRepository
	MonthlyJobRepo repositories.MonthlyJobRepository // <-- TAMBAH INI
	AnnualJobRepo  repositories.AnnualJobRepository  // <-- TAMBAH INI
	Sp2dkJobRepo   repositories.Sp2dkJobRepository 
	PemeriksaanJobRepo repositories.PemeriksaanJobRepository
}

// NewClientHandler creates a new ClientHandler
func NewClientHandler(cRepo repositories.ClientRepository, sRepo repositories.StaffRepository,
	mjRepo repositories.MonthlyJobRepository, ajRepo repositories.AnnualJobRepository, sjRepo repositories.Sp2dkJobRepository, pjRepo repositories.PemeriksaanJobRepository) *ClientHandler { 
	return &ClientHandler{
		ClientRepo:     cRepo,
		StaffRepo:      sRepo,
		MonthlyJobRepo: mjRepo, 
		AnnualJobRepo:  ajRepo,  
		Sp2dkJobRepo:   sjRepo,
		PemeriksaanJobRepo: pjRepo,   
	}
}

// CreateClient handles the creation of a new client
// CreateClient handles the creation of a new client
func (h *ClientHandler) CreateClient(c *gin.Context) {
	var req models.NewClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

    // Optional: Validate if PicStaffSigmaID exists
    if req.PicStaffSigmaID != "" {
        staff, err := h.StaffRepo.GetStaffByID(req.PicStaffSigmaID) // Need StaffRepo in ClientHandler
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate PIC Staff ID"})
            return
        }
        if staff == nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "PIC Staff ID not found"})
            return
        }
    }

	// Hash the CoretaxPassword from the request
	hashedPassword, err := utils.HashPassword(req.CoretaxPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	client := &models.Client{
		ClientName:           req.ClientName,
		NpwpClient:           req.NpwpClient,
		AddressClient:        req.AddressClient,
		MembershipStatus:     req.MembershipStatus,
		PhoneClient:          req.PhoneClient,
		EmailClient:          req.EmailClient,
		PicClient:            req.PicClient,
		DjpOnlineUsername:    req.DjpOnlineUsername,
		CoretaxUsername:      req.CoretaxUsername,
		CoretaxPasswordHashed: hashedPassword,
		PicStaffSigmaID:      req.PicStaffSigmaID, // UBAH INI
		ClientCategory:       req.ClientCategory,
		PphFinalUmkm:         req.PphFinalUmkm,
		Pph25:                req.Pph25,
		Pph21:                req.Pph21,
		PphUnifikasi:         req.PphUnifikasi,
		Ppn:                  req.Ppn,
		SptTahunan:           req.SptTahunan,
		PelaporanDeviden:     req.PelaporanDeviden,
		LaporanKeuangan:      req.LaporanKeuangan,
		InvestasiDeviden:     req.InvestasiDeviden,
	}

	if err := h.ClientRepo.CreateClient(client); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create client: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, client)
}

// GetAllClients fetches all clients
func (h *ClientHandler) GetAllClients(c *gin.Context) {
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
	clients, err := h.ClientRepo.GetAllClients(userClaims.StaffID, userClaims.IsAdmin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve clients"})
		return
	}
	c.JSON(http.StatusOK, clients)
}

// internal/handlers/client_handler.go

func (h *ClientHandler) GetClientDashboardJobs(c *gin.Context) {
	clientID := c.Param("id")
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

	// 1. Validasi keberadaan klien
	client, err := h.ClientRepo.GetClientByID(clientID, userClaims.StaffID, userClaims.IsAdmin)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Client not found or access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve client for dashboard: " + err.Error()})
		return
	}

	// 2. Ambil semua jenis pekerjaan untuk klien ini
	// ================== AWAL PERBAIKAN LOGIKA ==================

	monthlyJobs, err := h.MonthlyJobRepo.GetMonthlyJobsByClientID(clientID, userClaims.StaffID, userClaims.IsAdmin)
	if err != nil && err != sql.ErrNoRows { // Hanya return jika error BUKAN karena tidak ditemukan
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve monthly jobs for client: " + err.Error()})
		return
	}

	annualJobs, err := h.AnnualJobRepo.GetAnnualJobsByClientID(clientID, userClaims.StaffID, userClaims.IsAdmin)
	if err != nil && err != sql.ErrNoRows { // Terapkan pola yang sama
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve annual jobs for client: " + err.Error()})
		return
	}

	sp2dkJobs, err := h.Sp2dkJobRepo.GetSp2dkJobsByClientID(clientID, userClaims.StaffID, userClaims.IsAdmin)
	if err != nil && err != sql.ErrNoRows { // Terapkan pola yang sama
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve SP2DK jobs for client: " + err.Error()})
		return
	}

	pemeriksaanJobs, err := h.PemeriksaanJobRepo.GetPemeriksaanJobsByClientID(clientID, userClaims.StaffID, userClaims.IsAdmin)
	if err != nil && err != sql.ErrNoRows { // Terapkan pola yang sama
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve Pemeriksaan jobs for client: " + err.Error()})
		return
	}

	// ================== AKHIR PERBAIKAN LOGIKA ==================

	// 3. Gabungkan hasilnya ke dalam satu respons JSON
	// Jika sebuah pekerjaan tidak ada, variabelnya akan `nil` dan Gin akan mengubahnya menjadi array kosong `[]` di JSON.
	response := gin.H{
		"client_id":        client.ClientID,
		"client_name":      client.ClientName,
		"npwp_client":      client.NpwpClient,
		"monthly_jobs":     monthlyJobs,
		"annual_jobs":      annualJobs,
		"sp2dk_jobs":       sp2dkJobs,
		"pemeriksaan_jobs": pemeriksaanJobs,
	}

	c.JSON(http.StatusOK, response)
}

// GetClientByID fetches a single client by ID
func (h *ClientHandler) GetClientByID(c *gin.Context) {
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
	client, err := h.ClientRepo.GetClientByID(id, userClaims.StaffID, userClaims.IsAdmin)
	if err != nil {
		if err == sql.ErrNoRows { // Ini akan menangani not found (baik karena ID salah atau tidak punya akses)
			c.JSON(http.StatusNotFound, gin.H{"error": "Client not found or access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve client: " + err.Error()})
		return
	}
	if client == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found or access denied"})
		return
	}
	c.JSON(http.StatusOK, client)
}

// UpdateClient handles updating an existing client
// UpdateClient handles updating an existing client
func (h *ClientHandler) UpdateClient(c *gin.Context) {
	id := c.Param("id")
	var req models.UpdateClientRequest
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
	existingClient, err := h.ClientRepo.GetClientByID(id, userClaims.StaffID, userClaims.IsAdmin)
	if err != nil {
        if err == sql.ErrNoRows {
            c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
            return
        }
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve client for update: " + err.Error()})
		return
	}
	if existingClient == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	// Apply updates only for fields that are provided in the request
	if req.ClientName != nil {
		existingClient.ClientName = *req.ClientName
	}
	if req.NpwpClient != nil {
		existingClient.NpwpClient = *req.NpwpClient
	}
	if req.AddressClient != nil {
		existingClient.AddressClient = *req.AddressClient
	}
	if req.MembershipStatus != nil {
		existingClient.MembershipStatus = *req.MembershipStatus
	}
	if req.PhoneClient != nil {
		existingClient.PhoneClient = *req.PhoneClient
	}
	if req.EmailClient != nil {
        // ... (logic for email conflict check remains the same) ...
        existingClient.EmailClient = *req.EmailClient
	}
	if req.PicClient != nil {
		existingClient.PicClient = *req.PicClient
	}
	if req.DjpOnlineUsername != nil {
		existingClient.DjpOnlineUsername = *req.DjpOnlineUsername
	}
	if req.CoretaxUsername != nil {
		existingClient.CoretaxUsername = *req.CoretaxUsername
	}
	if req.PicStaffSigmaID != nil { // UBAH INI
        // Optional: Validate if PicStaffSigmaID exists
        staff, err := h.StaffRepo.GetStaffByID(*req.PicStaffSigmaID) // Need StaffRepo
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate PIC Staff ID"})
            return
        }
        if staff == nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "PIC Staff ID not found"})
            return
        }
		existingClient.PicStaffSigmaID = *req.PicStaffSigmaID
	}
	if req.ClientCategory != nil {
		existingClient.ClientCategory = *req.ClientCategory
	}

	// ... (bagian layanan pajak tetap sama) ...

	// Only hash and update password if a new password is provided in the request
	if req.CoretaxPassword != nil && *req.CoretaxPassword != "" {
		hashedPassword, err := utils.HashPassword(*req.CoretaxPassword)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password"})
			return
		}
		existingClient.CoretaxPasswordHashed = hashedPassword
	}

	if err := h.ClientRepo.UpdateClient(existingClient); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update client: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, existingClient)
}

// DeleteClient handles deleting a client by ID
func (h *ClientHandler) DeleteClient(c *gin.Context) {
	id := c.Param("id")

	 // Ambil staffID dan role dari context untuk filtering (untuk validasi keberadaan dan akses)
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

	// Check if client exists before deleting
	client, err := h.ClientRepo.GetClientByID(id, userClaims.StaffID, userClaims.IsAdmin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check client existence"})
		return
	}
	if client == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	if err := h.ClientRepo.DeleteClient(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete client"})
		return
	}
	c.JSON(http.StatusNoContent, nil) // 204 No Content for successful deletion
}
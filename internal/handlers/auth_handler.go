package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5" // Import library JWT

	// Pastikan ini modul Anda
	"github.com/iqsanfm/dashboard-pekerjaan-backend/internal/repositories" // Pastikan ini modul Anda
	"github.com/iqsanfm/dashboard-pekerjaan-backend/pkg/utils"             // Untuk CheckPasswordHash
)

// AuthHandler handles authentication related requests
type AuthHandler struct {
	StaffRepo repositories.StaffRepository
	JWTSecret string
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(staffRepo repositories.StaffRepository, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		StaffRepo: staffRepo,
		JWTSecret: jwtSecret,
	}
}

// LoginRequest struct for login input
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Login handles staff login and JWT token generation
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	staff, err := h.StaffRepo.GetStaffByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication failed"})
		return
	}
	if staff == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Verify password
	if !utils.CheckPasswordHash(req.Password, staff.PasswordHashed) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token
	// Custom claims: staff_id and role
	claims := jwt.MapClaims{
		"staff_id": staff.StaffID,
		"role":     staff.Role,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // Token berlaku 24 jam
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(h.JWTSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": signedToken})
}
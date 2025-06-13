package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iqsanfm/dashboard-pekerjaan-backend/internal/models"
	"github.com/iqsanfm/dashboard-pekerjaan-backend/internal/repositories"
	"github.com/iqsanfm/dashboard-pekerjaan-backend/pkg/utils"
)

// StaffHandler handles HTTP requests for staff operations
type StaffHandler struct {
	StaffRepo repositories.StaffRepository
}

// NewStaffHandler creates a new StaffHandler
func NewStaffHandler(repo repositories.StaffRepository) *StaffHandler {
	return &StaffHandler{StaffRepo: repo}
}

// CreateStaff handles the creation of a new staff member
func (h *StaffHandler) CreateStaff(c *gin.Context) {
	var req models.NewStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if email already exists
	existingStaff, err := h.StaffRepo.GetStaffByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check email existence"})
		return
	}
	if existingStaff != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	staff := &models.Staff{
		Nama:           req.Nama,
		Email:          req.Email,
		PasswordHashed: req.Password, // Temporarily hold plain password for hashing in repo
		Role:           req.Role,
	}

	if err := h.StaffRepo.CreateStaff(staff); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create staff: " + err.Error()})
		return
	}

	// Remove password hash before sending response
	staff.PasswordHashed = ""
	c.JSON(http.StatusCreated, staff)
}

// GetAllStaffs fetches all staff members
func (h *StaffHandler) GetAllStaffs(c *gin.Context) {
	staffs, err := h.StaffRepo.GetAllStaffs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve staffs: " + err.Error()})
		return
	}

	// Remove password hashes from response
	for i := range staffs {
		staffs[i].PasswordHashed = ""
	}
	c.JSON(http.StatusOK, staffs)
}

// GetStaffByID fetches a single staff member by ID
func (h *StaffHandler) GetStaffByID(c *gin.Context) {
	id := c.Param("id")

	staff, err := h.StaffRepo.GetStaffByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Staff not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve staff: " + err.Error()})
		return
	}
	if staff == nil { // Double-check in case repo returns nil without sql.ErrNoRows
		c.JSON(http.StatusNotFound, gin.H{"error": "Staff not found"})
		return
	}

	// Remove password hash from response
	staff.PasswordHashed = ""
	c.JSON(http.StatusOK, staff)
}

// UpdateStaff handles partial updates to an existing staff member
func (h *StaffHandler) UpdateStaff(c *gin.Context) {
	id := c.Param("id")
	var req models.UpdateStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existingStaff, err := h.StaffRepo.GetStaffByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Staff not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve staff for update: " + err.Error()})
		return
	}
	if existingStaff == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Staff not found"})
		return
	}

	// Apply updates only for fields that are provided in the request
	if req.Nama != nil {
		existingStaff.Nama = *req.Nama
	}
	if req.Email != nil {
		// Check for email conflict if email is updated
		if *req.Email != existingStaff.Email {
			conflictStaff, err := h.StaffRepo.GetStaffByEmail(*req.Email)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check new email existence"})
				return
			}
			if conflictStaff != nil {
				c.JSON(http.StatusConflict, gin.H{"error": "New email already registered by another staff"})
				return
			}
		}
		existingStaff.Email = *req.Email
	}
	if req.Role != nil {
		existingStaff.Role = *req.Role
	}
	// Note: Password update is usually handled by a separate "change password" endpoint
	// For this general update, we don't allow changing password here.

	// Re-generate NIP if name is changed (optional, but good practice if NIP is derived from name)
	if req.Nama != nil {
		existingStaff.NIP = utils.GenerateNIP(existingStaff.Nama)
	}

	if err := h.StaffRepo.UpdateStaff(existingStaff); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update staff: " + err.Error()})
		return
	}

	// Remove password hash from response
	existingStaff.PasswordHashed = ""
	c.JSON(http.StatusOK, existingStaff)
}

// DeleteStaff handles deleting a staff member by ID
func (h *StaffHandler) DeleteStaff(c *gin.Context) {
	id := c.Param("id")

	// Optional: Check if staff exists before deleting
	staff, err := h.StaffRepo.GetStaffByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check staff existence"})
		return
	}
	if staff == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Staff not found"})
		return
	}

	if err := h.StaffRepo.DeleteStaff(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete staff: " + err.Error()})
		return
	}
	c.Status(http.StatusNoContent) // 204 No Content for successful deletion
}

// ChangeStaffPassword handles changing a staff member's password
func (h *StaffHandler) ChangeStaffPassword(c *gin.Context) {
    id := c.Param("id")
    var req struct {
        NewPassword string `json:"new_password" binding:"required,min=6"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    staff, err := h.StaffRepo.GetStaffByID(id)
    if err != nil {
        if err == sql.ErrNoRows {
            c.JSON(http.StatusNotFound, gin.H{"error": "Staff not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve staff: " + err.Error()})
        return
    }
    if staff == nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Staff not found"})
        return
    }

    hashedPassword, err := utils.HashPassword(req.NewPassword)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password"})
        return
    }

    staff.PasswordHashed = hashedPassword // Update with the new hashed password

    if err := h.StaffRepo.UpdateStaff(staff); err != nil { // Reuse UpdateStaff
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update staff password: " + err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}
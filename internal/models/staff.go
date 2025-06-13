package models

import (
	"time"
)

// Staff represents a staff member
type Staff struct {
	StaffID        string    `json:"staff_id"`
	NIP            string    `json:"nip"`
	Nama           string    `json:"nama"`
	Email          string    `json:"email"`
	PasswordHashed string    `json:"-"` // Jangan kirim password ke client
	Role           string    `json:"role"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// NewStaffRequest represents the structure for creating a new staff member (input from API)
type NewStaffRequest struct {
	Nama     string `json:"nama" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required"`
}

// UpdateStaffRequest represents the structure for updating a staff member (partial update via PATCH)
type UpdateStaffRequest struct {
	Nama  *string `json:"nama"`
	Email *string `json:"email" binding:"omitempty,email"`
	Role  *string `json:"role"`
}
package models

import (
	"time"
)

// PemeriksaanJob represents a Pemeriksaan job for a client
type PemeriksaanJob struct {
	JobID                     string     `json:"job_id"`
	ClientID                  string     `json:"client_id"`
	ClientName                string     `json:"client_name"` // Populated from clients table
	NpwpClient                string     `json:"npwp_client"` // Populated from clients table
	AssignedPicStaffSigmaID   string     `json:"assigned_pic_staff_sigma_id"`
	AssignedPicStaffSigmaName string     `json:"assigned_pic_staff_sigma_name"` // Populated from staffs table
	ContractNo                string     `json:"contract_no"`
	ContractDate              *time.Time `json:"contract_date"` // Use pointer for nullable DATE
	Sp2No                     string     `json:"sp2_no"`
	Sp2Date                   *time.Time `json:"sp2_date"` // Use pointer for nullable DATE
	SkpNo                     string     `json:"skp_no"`
	SkpDate                   *time.Time `json:"skp_date"` // Use pointer for nullable DATE
	JobStatus                 string     `json:"job_status"` // "Dikerjakan", "Selesai", "Dibatalkan"
	ProofOfWorkURL            *string    `json:"proof_of_work_url"`
	CreatedAt                 time.Time  `json:"created_at"`
	UpdatedAt                 time.Time  `json:"updated_at"`
}

// NewPemeriksaanJobRequest represents the structure for creating a new Pemeriksaan job (input from API)
type NewPemeriksaanJobRequest struct {
	ClientID                string     `json:"client_id" binding:"required"`
	AssignedPicStaffSigmaID string     `json:"assigned_pic_staff_sigma_id"`
	ContractNo              string     `json:"contract_no"`
	ContractDate            *time.Time `json:"contract_date"`
	Sp2No                   string     `json:"sp2_no"`
	Sp2Date                 *time.Time `json:"sp2_date"`
	SkpNo                   string     `json:"skp_no"`
	SkpDate                 *time.Time `json:"skp_date"`
	JobStatus               string     `json:"job_status"` // e.g., "Dikerjakan", "Selesai", "Dibatalkan"
}

// UpdatePemeriksaanJobRequest represents the structure for updating a Pemeriksaan job (partial update via PATCH)
type UpdatePemeriksaanJobRequest struct {
	AssignedPicStaffSigmaID *string    `json:"assigned_pic_staff_sigma_id"`
	ContractNo              *string    `json:"contract_no"`
	ContractDate            *time.Time `json:"contract_date"`
	Sp2No                   *string    `json:"sp2_no"`
	Sp2Date                 *time.Time `json:"sp2_date"`
	SkpNo                   *string    `json:"skp_no"`
	SkpDate                 *time.Time `json:"skp_date"`
	JobStatus               *string    `json:"job_status"`
	ProofOfWorkURL          *string    `json:"proof_of_work_url"`
}
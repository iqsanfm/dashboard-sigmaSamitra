package models

import (
	"time"
)

// Sp2dkJob represents an SP2DK job for a client
type Sp2dkJob struct {
	JobID                     string     `json:"job_id"`
	ClientID                  string     `json:"client_id"`
	ClientName                string     `json:"client_name"` // Populated from clients table
	NpwpClient                string     `json:"npwp_client"` // Populated from clients table
	AssignedPicStaffSigmaID   string     `json:"assigned_pic_staff_sigma_id"`
	AssignedPicStaffSigmaName string     `json:"assigned_pic_staff_sigma_name"` // Populated from staffs table
	ContractNo                string     `json:"contract_no"`
	ContractDate              *time.Time `json:"contract_date"` // Use pointer for nullable DATE
	Sp2dkNo                   string     `json:"sp2dk_no"`
	Sp2dkDate                 *time.Time `json:"sp2dk_date"` // Use pointer for nullable DATE
	Bap2dkNo                  string     `json:"bap2dk_no"`
	Bap2dkDate                *time.Time `json:"bap2dk_date"` // Use pointer for nullable DATE
	PaymentDate               *time.Time `json:"payment_date"` // Use pointer for nullable DATE
	ReportDate                *time.Time `json:"report_date"` // Use pointer for nullable DATE
	OverallStatus    					string     `json:"overall_status"`
	ProofOfWorkURL            *string    `json:"proof_of_work_url"`
	CreatedAt                 time.Time  `json:"created_at"`
	UpdatedAt                 time.Time  `json:"updated_at"`
}

// NewSp2dkJobRequest represents the structure for creating a new SP2DK job (input from API)
type NewSp2dkJobRequest struct {
	ClientID                string     `json:"client_id" binding:"required"`
	AssignedPicStaffSigmaID string     `json:"assigned_pic_staff_sigma_id"`
	ContractNo              string     `json:"contract_no"`
	ContractDate            *time.Time `json:"contract_date"`
	Sp2dkNo                 string     `json:"sp2dk_no"`
	Sp2dkDate               *time.Time `json:"sp2dk_date"`
	Bap2dkNo                string     `json:"bap2dk_no"`
	Bap2dkDate              *time.Time `json:"bap2dk_date"`
	PaymentDate             *time.Time `json:"payment_date"`
	ReportDate              *time.Time `json:"report_date"`
	OverallStatus    				 string     `json:"overall_status"` // e.g., "Dikerjakan", "Selesai", "Dibatalkan"
}

// UpdateSp2dkJobRequest represents the structure for updating an SP2DK job (partial update via PATCH)
type UpdateSp2dkJobRequest struct {
	AssignedPicStaffSigmaID *string    `json:"assigned_pic_staff_sigma_id"`
	ContractNo              *string    `json:"contract_no"`
	ContractDate            *time.Time `json:"contract_date"`
	Sp2dkNo                 *string    `json:"sp2dk_no"`
	Sp2dkDate               *time.Time `json:"sp2dk_date"`
	Bap2dkNo                *string    `json:"bap2dk_no"`
	Bap2dkDate              *time.Time `json:"bap2dk_date"`
	PaymentDate             *time.Time `json:"payment_date"`
	ReportDate              *time.Time `json:"report_date"`
	OverallStatus     			string     `json:"overall_status"`
	ProofOfWorkURL          *string    `json:"proof_of_work_url"`
}
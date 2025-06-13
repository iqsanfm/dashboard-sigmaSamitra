package models

import (
	"time"
)

// MonthlyTaxReport (Ini tidak berubah)
type MonthlyTaxReport struct {
	ReportID       string    `json:"report_id"`
	JobID          string    `json:"job_id"`
	TaxType        string    `json:"tax_type"`
	BillingCode    string    `json:"billing_code"`
	PaymentDate    *time.Time `json:"payment_date"`
	PaymentAmount  *float64  `json:"payment_amount"`
	ReportStatus   string    `json:"report_status"`
	ReportDate     *time.Time `json:"report_date"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// MonthlyJob represents a monthly job for a client, including client details and tax reports
type MonthlyJob struct {
	JobID                 string             `json:"job_id"`
	ClientID              string             `json:"client_id"`
	ClientName            string             `json:"client_name"`
	NpwpClient            string             `json:"npwp_client"`
	JobMonth              int                `json:"job_month"`
	JobYear               int                `json:"job_year"`
	
	AssignedPicStaffSigmaID string             `json:"assigned_pic_staff_sigma_id"`   // ID PIC Staff Sigma (baru)
	AssignedPicStaffSigmaName string           `json:"assigned_pic_staff_sigma_name"` // Nama PIC Staff Sigma (dari JOIN)
	OverallStatus         string             `json:"overall_status"`

	ProofOfWorkURL            *string            `json:"proof_of_work_url"`
	CreatedAt             time.Time          `json:"created_at"`
	UpdatedAt             time.Time          `json:"updated_at"`
	TaxReports            []MonthlyTaxReport `json:"tax_reports"`
}

// NewMonthlyJobRequest (sesuaikan dengan perubahan)
type NewMonthlyJobRequest struct {
	ClientID              string `json:"client_id" binding:"required"`
	JobMonth              int    `json:"job_month" binding:"required"`
	JobYear               int    `json:"job_year" binding:"required"`
	AssignedPicStaffSigmaID string `json:"assigned_pic_staff_sigma_id"` // ID PIC Staff Sigma
	OverallStatus         string `json:"overall_status"`
	ProofOfWorkURL        *string `json:"proof_of_work_url"`

	TaxReports []NewMonthlyTaxReportRequest `json:"tax_reports"`
}

// NewMonthlyTaxReportRequest (tidak berubah)
type NewMonthlyTaxReportRequest struct {
	TaxType       string     `json:"tax_type" binding:"required"`
	BillingCode   string     `json:"billing_code"`
	PaymentDate   *time.Time `json:"payment_date"`
	PaymentAmount *float64   `json:"payment_amount"`
	ReportStatus  string     `json:"report_status"`
	ReportDate    *time.Time `json:"report_date"`
}

// UpdateMonthlyJobRequest (sesuaikan dengan perubahan)
type UpdateMonthlyJobRequest struct {
	JobMonth              *int    `json:"job_month"`
	JobYear               *int    `json:"job_year"`
	AssignedPicStaffSigmaID *string `json:"assigned_pic_staff_sigma_id"` // ID PIC Staff Sigma
	OverallStatus         *string `json:"overall_status"`
}

// UpdateMonthlyTaxReportRequest (tidak berubah)
type UpdateMonthlyTaxReportRequest struct {
	TaxType       *string    `json:"tax_type"`
	BillingCode   *string    `json:"billing_code"`
	PaymentDate   *time.Time `json:"payment_date"`
	PaymentAmount *float64   `json:"payment_amount"`
	ReportStatus  *string    `json:"report_status"`
	ReportDate    *time.Time `json:"report_date"`
}
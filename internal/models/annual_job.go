package models

import (
	"time"
)

// AnnualTaxReport represents the annual tax (SPT Tahunan) report details
type AnnualTaxReport struct {
	ReportID       string     `json:"report_id"`
	JobID          string     `json:"job_id"` // Foreign key
	BillingCode    string     `json:"billing_code"`
	PaymentDate    *time.Time `json:"payment_date"` // Use pointer for nullable DATE
	PaymentAmount  *float64   `json:"payment_amount"` // Use pointer for nullable NUMERIC
	ReportDate     *time.Time `json:"report_date"` // Use pointer for nullable DATE
	ReportStatus   string     `json:"report_status"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// AnnualDividendReport represents the dividend investment report details
type AnnualDividendReport struct {
	ReportID       string     `json:"report_id"`
	JobID          string     `json:"job_id"` // Foreign key
	IsReported     bool       `json:"is_reported"`
	ReportDate     *time.Time `json:"report_date"` // Use pointer for nullable DATE
	ReportStatus   string     `json:"report_status"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// AnnualJob represents an annual job for a client, including client details and nested reports
type AnnualJob struct {
	JobID                     string                `json:"job_id"`
	ClientID                  string                `json:"client_id"`
	ClientName                string                `json:"client_name"` // Populated from clients table
	NpwpClient                string                `json:"npwp_client"` // Populated from clients table
	JobYear                   int                   `json:"job_year"`
	AssignedPicStaffSigmaID   string                `json:"assigned_pic_staff_sigma_id"`
	AssignedPicStaffSigmaName string                `json:"assigned_pic_staff_sigma_name"` // Populated from staffs table
	OverallStatus             string                `json:"overall_status"`
	ProofOfWorkURL            *string               `json:"proof_of_work_url"`
	CreatedAt                 time.Time             `json:"created_at"`
	UpdatedAt                 time.Time             `json:"updated_at"`
	TaxReports                []AnnualTaxReport     `json:"tax_reports"` // Nested slice for SPT Tahunan
	DividendReports           []AnnualDividendReport `json:"dividend_reports"` // Nested slice for Investasi Dividen
}

// NewAnnualJobRequest represents the structure for creating a new annual job (input from API)
type NewAnnualJobRequest struct {
	ClientID                string `json:"client_id" binding:"required"`
	JobYear                 int    `json:"job_year" binding:"required"`
	AssignedPicStaffSigmaID string `json:"assigned_pic_staff_sigma_id"`
	OverallStatus           string `json:"overall_status"`

	// Initial reports can be part of creation request
	TaxReport      *NewAnnualTaxReportRequest      `json:"tax_report"`      // Pakai pointer karena bisa null
	DividendReport *NewAnnualDividendReportRequest `json:"dividend_report"` // Pakai pointer karena bisa null
}

// NewAnnualTaxReportRequest represents the input for creating an SPT Tahunan report
type NewAnnualTaxReportRequest struct {
	BillingCode   string     `json:"billing_code"`
	PaymentDate   *time.Time `json:"payment_date"`
	PaymentAmount *float64   `json:"payment_amount"`
	ReportDate    *time.Time `json:"report_date"`
	ReportStatus  string     `json:"report_status"`
}

// NewAnnualDividendReportRequest represents the input for creating an Investasi Dividen report
type NewAnnualDividendReportRequest struct {
	IsReported   bool       `json:"is_reported"`
	ReportDate   *time.Time `json:"report_date"`
	ReportStatus string     `json:"report_status"`
}

// UpdateAnnualJobRequest represents the structure for updating an annual job (partial update via PATCH)
type UpdateAnnualJobRequest struct {
	JobYear                 *int    `json:"job_year"`
	AssignedPicStaffSigmaID *string `json:"assigned_pic_staff_sigma_id"`
	OverallStatus           *string `json:"overall_status"`
	ProofOfWorkURL          *string `json:"proof_of_work_url"`
}

// UpdateAnnualTaxReportRequest represents the input for updating an SPT Tahunan report
type UpdateAnnualTaxReportRequest struct {
	BillingCode   *string    `json:"billing_code"`
	PaymentDate   *time.Time `json:"payment_date"`
	PaymentAmount *float64   `json:"payment_amount"`
	ReportDate    *time.Time `json:"report_date"`
	ReportStatus  *string    `json:"report_status"`
}

// UpdateAnnualDividendReportRequest represents the input for updating an Investasi Dividen report
type UpdateAnnualDividendReportRequest struct {
	IsReported   *bool      `json:"is_reported"`
	ReportDate   *time.Time `json:"report_date"`
	ReportStatus *string    `json:"report_status"`
}
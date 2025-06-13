package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/iqsanfm/dashboard-pekerjaan-backend/internal/models" // Pastikan ini adalah modul Anda
)

// AnnualJobRepository defines the interface for annual job data operations
type AnnualJobRepository interface {
	CreateAnnualJob(job *models.AnnualJob) error
	GetAllAnnualJobs(staffIDFilter string, isAdmin bool) ([]models.AnnualJob, error)
	GetAnnualJobByID(id string, staffIDFilter string, isAdmin bool) (*models.AnnualJob, error)
	UpdateAnnualJob(job *models.AnnualJob) error

	// Methods for Annual Tax Reports (SPT Tahunan)
	CreateAnnualTaxReport(report *models.AnnualTaxReport) error
	GetAnnualTaxReportByID(reportID string) (*models.AnnualTaxReport, error)
	UpdateAnnualTaxReport(report *models.AnnualTaxReport) error
	DeleteAnnualTaxReport(reportID string) error

	// Methods for Annual Dividend Reports (Investasi Dividen)
	CreateAnnualDividendReport(report *models.AnnualDividendReport) error
	GetAnnualDividendReportByID(reportID string) (*models.AnnualDividendReport, error)
	UpdateAnnualDividendReport(report *models.AnnualDividendReport) error
	DeleteAnnualDividendReport(reportID string) error
	GetAnnualJobsByClientID(clientID string, staffIDFilter string, isAdmin bool) ([]models.AnnualJob, error)
}

// annualJobRepository implements AnnualJobRepository interface
type annualJobRepository struct {
	db *sql.DB
}

// NewAnnualJobRepository creates a new AnnualJobRepository
func NewAnnualJobRepository(db *sql.DB) AnnualJobRepository {
	return &annualJobRepository{db: db}
}

// CreateAnnualJob inserts a new annual job and its associated reports into the database.
// This operation is wrapped in a transaction.
func (r *annualJobRepository) CreateAnnualJob(job *models.AnnualJob) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback on error, commit manually on success

	var assignedPicStaffSigmaID sql.NullString
	if job.AssignedPicStaffSigmaID != "" {
		assignedPicStaffSigmaID = sql.NullString{String: job.AssignedPicStaffSigmaID, Valid: true}
	} else {
		assignedPicStaffSigmaID = sql.NullString{Valid: false} // Akan menjadi NULL di database
	}

	// 1. Insert into annual_jobs table
	jobQuery := `INSERT INTO annual_jobs (
		client_id, job_year, assigned_pic_staff_sigma_id, overall_status, created_at, updated_at
	) VALUES (
		$1, $2, $3, $4, $5, $6
	) RETURNING job_id, created_at, updated_at`

	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now()
	}
	job.UpdatedAt = time.Now()

	err = tx.QueryRow(jobQuery,
		job.ClientID, job.JobYear, assignedPicStaffSigmaID, job.OverallStatus,
		job.CreatedAt, job.UpdatedAt,
	).Scan(&job.JobID, &job.CreatedAt, &job.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create annual job: %w", err)
	}

	// 2. Insert associated Annual Tax Report (if provided)
	if len(job.TaxReports) > 0 { // Kita tahu ini hanya 0 atau 1 laporan SPT Tahunan
		report := &job.TaxReports[0] // Ambil laporan pertama (dan satu-satunya)
		report.JobID = job.JobID     // Link report to the newly created job
		if report.CreatedAt.IsZero() {
			report.CreatedAt = time.Now()
		}
		report.UpdatedAt = time.Now()

		taxReportQuery := `INSERT INTO annual_tax_reports (
			job_id, billing_code, payment_date, payment_amount, report_date, report_status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		) RETURNING report_id, created_at, updated_at`

		err := tx.QueryRow(taxReportQuery,
			report.JobID, report.BillingCode, report.PaymentDate, report.PaymentAmount,
			report.ReportDate, report.ReportStatus,
			report.CreatedAt, report.UpdatedAt,
		).Scan(&report.ReportID, &report.CreatedAt, &report.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to create annual tax report for job %s: %w", job.JobID, err)
		}
	}

	// 3. Insert associated Annual Dividend Report (if provided)
	if len(job.DividendReports) > 0 { // Kita tahu ini hanya 0 atau 1 laporan Dividen
		report := &job.DividendReports[0] // Ambil laporan pertama (dan satu-satunya)
		report.JobID = job.JobID          // Link report to the newly created job
		if report.CreatedAt.IsZero() {
			report.CreatedAt = time.Now()
		}
		report.UpdatedAt = time.Now()

		dividendReportQuery := `INSERT INTO annual_dividend_reports (
			job_id, is_reported, report_date, report_status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6
		) RETURNING report_id, created_at, updated_at`

		err := tx.QueryRow(dividendReportQuery,
			report.JobID, report.IsReported, report.ReportDate, report.ReportStatus,
			report.CreatedAt, report.UpdatedAt,
		).Scan(&report.ReportID, &report.CreatedAt, &report.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to create annual dividend report for job %s: %w", job.JobID, err)
		}
	}

	return tx.Commit() // Commit the transaction
}

// GetAllAnnualJobs fetches all annual jobs with their associated client and reports.
func (r *annualJobRepository) GetAllAnnualJobs(staffIDFilter string, isAdmin bool) ([]models.AnnualJob, error) {
	query := `
	SELECT
		aj.job_id, aj.client_id, c.client_name, c.npwp_client, aj.job_year,
		aj.assigned_pic_staff_sigma_id, s.nama AS assigned_pic_staff_sigma_name,
		aj.overall_status, aj.proof_of_work_url, aj.created_at, aj.updated_at,
		
		atr.report_id AS atr_report_id, atr.billing_code, atr.payment_date AS atr_payment_date,
		atr.payment_amount AS atr_payment_amount, atr.report_date AS atr_report_date, atr.report_status AS atr_report_status,
		atr.created_at AS atr_created_at, atr.updated_at AS atr_updated_at,

		adr.report_id AS adr_report_id, adr.is_reported, adr.report_date AS adr_report_date, adr.report_status AS adr_report_status,
		adr.created_at AS adr_created_at, adr.updated_at AS adr_updated_at
	FROM annual_jobs AS aj
	JOIN clients AS c ON aj.client_id = c.client_id
	LEFT JOIN staffs AS s ON aj.assigned_pic_staff_sigma_id = s.staff_id
	LEFT JOIN annual_tax_reports AS atr ON aj.job_id = atr.job_id
	LEFT JOIN annual_dividend_reports AS adr ON aj.job_id = adr.job_id`

	args := []interface{}{}
	paramCounter := 1

	if !isAdmin && staffIDFilter != "" {
		query += fmt.Sprintf(" WHERE aj.assigned_pic_staff_sigma_id = $%d", paramCounter)
		args = append(args, staffIDFilter)
		paramCounter++
	}

	query += " ORDER BY aj.job_year DESC, aj.created_at DESC;"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get all annual jobs: %w", err)
	}
	defer rows.Close()

	annualJobsMap := make(map[string]*models.AnnualJob)
	var annualJobsList []models.AnnualJob

	for rows.Next() {
		var (
			jobID                     string
			clientID                  string
			clientName                string
			npwpClient                string
			jobYear                   int
			assignedPicStaffSigmaID   sql.NullString
			assignedPicStaffSigmaName sql.NullString
			overallStatus             string
			proofOfWorkURL            sql.NullString
			jobCreatedAt              time.Time
			jobUpdatedAt              time.Time

			// Annual Tax Report fields (nullable from LEFT JOIN)
			atrReportID       sql.NullString
			atrBillingCode    sql.NullString
			atrPaymentDate    sql.NullTime
			atrPaymentAmount  sql.NullFloat64
			atrReportDate     sql.NullTime
			atrReportStatus   sql.NullString
			atrCreatedAt      sql.NullTime
			atrUpdatedAt      sql.NullTime

			// Annual Dividend Report fields (nullable from LEFT JOIN)
			adrReportID       sql.NullString
			adrIsReported     sql.NullBool
			adrReportDate     sql.NullTime
			adrReportStatus   sql.NullString
			adrCreatedAt      sql.NullTime
			adrUpdatedAt      sql.NullTime
		)

		err := rows.Scan(
			&jobID, &clientID, &clientName, &npwpClient, &jobYear,
			&assignedPicStaffSigmaID, &assignedPicStaffSigmaName,
			&overallStatus, &proofOfWorkURL, &jobCreatedAt, &jobUpdatedAt,

			&atrReportID, &atrBillingCode, &atrPaymentDate, &atrPaymentAmount, &atrReportDate, &atrReportStatus,
			&atrCreatedAt, &atrUpdatedAt,

			&adrReportID, &adrIsReported, &adrReportDate, &adrReportStatus,
			&adrCreatedAt, &adrUpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan annual job row: %w", err)
		}

		job, ok := annualJobsMap[jobID]
		if !ok {
			job = &models.AnnualJob{
				JobID:           jobID,
				ClientID:        clientID,
				ClientName:      clientName,
				NpwpClient:      npwpClient,
				JobYear:         jobYear,
				OverallStatus:   overallStatus,
				CreatedAt:       jobCreatedAt,
				UpdatedAt:       jobUpdatedAt,
				TaxReports:      []models.AnnualTaxReport{},
				DividendReports: []models.AnnualDividendReport{},
			}
			if assignedPicStaffSigmaID.Valid {
				job.AssignedPicStaffSigmaID = assignedPicStaffSigmaID.String
			}
			if assignedPicStaffSigmaName.Valid {
				job.AssignedPicStaffSigmaName = assignedPicStaffSigmaName.String
			}
			if proofOfWorkURL.Valid {
				job.ProofOfWorkURL = &proofOfWorkURL.String
			} else {
				job.ProofOfWorkURL = nil // Penting: set nil jika NULL di DB
			}
			annualJobsMap[jobID] = job
			annualJobsList = append(annualJobsList, *job)
		}

		// Add Annual Tax Report if exists
		if atrReportID.Valid {
			taxReport := models.AnnualTaxReport{
				ReportID:      atrReportID.String,
				JobID:         jobID,
				BillingCode:   atrBillingCode.String,
				ReportStatus:  atrReportStatus.String,
				CreatedAt:     atrCreatedAt.Time,
				UpdatedAt:     atrUpdatedAt.Time,
			}
			if atrPaymentDate.Valid {
				taxReport.PaymentDate = &atrPaymentDate.Time
			} else {
				taxReport.PaymentDate = nil
			}
			if atrPaymentAmount.Valid {
				taxReport.PaymentAmount = &atrPaymentAmount.Float64
			} else {
				taxReport.PaymentAmount = nil
			}
			if atrReportDate.Valid {
				taxReport.ReportDate = &atrReportDate.Time
			} else {
				taxReport.ReportDate = nil
			}
			job.TaxReports = append(job.TaxReports, taxReport)
			
		}

		// Add Annual Dividend Report if exists
		if adrReportID.Valid {
			dividendReport := models.AnnualDividendReport{
				ReportID:      adrReportID.String,
				JobID:         jobID,
				IsReported:    adrIsReported.Bool,
				ReportStatus:  adrReportStatus.String,
				CreatedAt:     adrCreatedAt.Time,
				UpdatedAt:     adrUpdatedAt.Time,
			}
			if adrReportDate.Valid {
				dividendReport.ReportDate = &adrReportDate.Time
			} else {
				dividendReport.ReportDate = nil
			}
			job.DividendReports = append(job.DividendReports, dividendReport)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return annualJobsList, nil
}

// GetAnnualJobByID fetches a single annual job by its ID with associated client and reports.
func (r *annualJobRepository) GetAnnualJobByID(id string, staffIDFilter string, isAdmin bool) (*models.AnnualJob, error) {
	query := `
	SELECT
		aj.job_id, aj.client_id, c.client_name, c.npwp_client, aj.job_year,
		aj.assigned_pic_staff_sigma_id, s.nama AS assigned_pic_staff_sigma_name,
		aj.overall_status, aj.proof_of_work_url, aj.created_at, aj.updated_at,
		
		atr.report_id AS atr_report_id, atr.billing_code, atr.payment_date AS atr_payment_date,
		atr.payment_amount AS atr_payment_amount, atr.report_date AS atr_report_date, atr.report_status AS atr_report_status,
		atr.created_at AS atr_created_at, atr.updated_at AS atr_updated_at,

		adr.report_id AS adr_report_id, adr.is_reported, adr.report_date AS adr_report_date, adr.report_status AS adr_report_status,
		adr.created_at AS adr_created_at, adr.updated_at AS adr_updated_at
	FROM annual_jobs AS aj
	JOIN clients AS c ON aj.client_id = c.client_id
	LEFT JOIN staffs AS s ON aj.assigned_pic_staff_sigma_id = s.staff_id
	LEFT JOIN annual_tax_reports AS atr ON aj.job_id = atr.job_id
	LEFT JOIN annual_dividend_reports AS adr ON aj.job_id = adr.job_id
	WHERE aj.job_id = $1`

	args := []interface{}{id}
	paramCounter := 2

	if !isAdmin && staffIDFilter != "" {
		query += fmt.Sprintf(" AND aj.assigned_pic_staff_sigma_id = $%d", paramCounter)
		args = append(args, staffIDFilter)
	}

	query += " ORDER BY atr.report_id ASC, adr.report_id ASC;" // Order reports for consistent scanning

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get annual job by ID: %w", err)
	}
	defer rows.Close()

	var annualJob *models.AnnualJob
	for rows.Next() {
		var (
			jobID                     string
			clientID                  string
			clientName                string
			npwpClient                string
			jobYear                   int
			assignedPicStaffSigmaID   sql.NullString
			assignedPicStaffSigmaName sql.NullString
			overallStatus             string
			proofOfWorkURL            sql.NullString
			jobCreatedAt              time.Time
			jobUpdatedAt              time.Time

			atrReportID       sql.NullString
			atrBillingCode    sql.NullString
			atrPaymentDate    sql.NullTime
			atrPaymentAmount  sql.NullFloat64
			atrReportDate     sql.NullTime
			atrReportStatus   sql.NullString
			atrCreatedAt      sql.NullTime
			atrUpdatedAt      sql.NullTime

			adrReportID       sql.NullString
			adrIsReported     sql.NullBool
			adrReportDate     sql.NullTime
			adrReportStatus   sql.NullString
			adrCreatedAt      sql.NullTime
			adrUpdatedAt      sql.NullTime
		)

		err := rows.Scan(
			&jobID, &clientID, &clientName, &npwpClient, &jobYear,
			&assignedPicStaffSigmaID, &assignedPicStaffSigmaName,
			&overallStatus, &proofOfWorkURL, &jobCreatedAt, &jobUpdatedAt,

			&atrReportID, &atrBillingCode, &atrPaymentDate, &atrPaymentAmount, &atrReportDate, &atrReportStatus,
			&atrCreatedAt, &atrUpdatedAt,

			&adrReportID, &adrIsReported, &adrReportDate, &adrReportStatus,
			&adrCreatedAt, &adrUpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan annual job by ID row: %w", err)
		}

		if annualJob == nil {
			annualJob = &models.AnnualJob{
				JobID:           jobID,
				ClientID:        clientID,
				ClientName:      clientName,
				NpwpClient:      npwpClient,
				JobYear:         jobYear,
				OverallStatus:   overallStatus,
				CreatedAt:       jobCreatedAt,
				UpdatedAt:       jobUpdatedAt,
				TaxReports:      []models.AnnualTaxReport{},
				DividendReports: []models.AnnualDividendReport{},
			}
			if assignedPicStaffSigmaID.Valid {
				annualJob.AssignedPicStaffSigmaID = assignedPicStaffSigmaID.String
			}
			if assignedPicStaffSigmaName.Valid {
				annualJob.AssignedPicStaffSigmaName = assignedPicStaffSigmaName.String
			}
			 if proofOfWorkURL.Valid {
                annualJob.ProofOfWorkURL = &proofOfWorkURL.String
            } else {
                annualJob.ProofOfWorkURL = nil // Penting: set nil jika NULL di DB
            }

		}

		// Add Annual Tax Report if exists
		if atrReportID.Valid {
			taxReport := models.AnnualTaxReport{
				ReportID:      atrReportID.String,
				JobID:         jobID,
				BillingCode:   atrBillingCode.String,
				ReportStatus:  atrReportStatus.String,
				CreatedAt:     atrCreatedAt.Time,
				UpdatedAt:     atrUpdatedAt.Time,
			}
			if atrPaymentDate.Valid {
				taxReport.PaymentDate = &atrPaymentDate.Time
			} else {
				taxReport.PaymentDate = nil
			}
			if atrPaymentAmount.Valid {
				taxReport.PaymentAmount = &atrPaymentAmount.Float64
			} else {
				taxReport.PaymentAmount = nil
			}
			if atrReportDate.Valid {
				taxReport.ReportDate = &atrReportDate.Time
			} else {
				taxReport.ReportDate = nil
			}
			annualJob.TaxReports = append(annualJob.TaxReports, taxReport)
		}

		// Add Annual Dividend Report if exists
		if adrReportID.Valid {
			dividendReport := models.AnnualDividendReport{
				ReportID:      adrReportID.String,
				JobID:         jobID,
				IsReported:    adrIsReported.Bool,
				ReportStatus:  adrReportStatus.String,
				CreatedAt:     adrCreatedAt.Time,
				UpdatedAt:     adrUpdatedAt.Time,
			}
			if adrReportDate.Valid {
				dividendReport.ReportDate = &adrReportDate.Time
			} else {
				dividendReport.ReportDate = nil
			}
			annualJob.DividendReports = append(annualJob.DividendReports, dividendReport)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	if annualJob == nil {
		return nil, sql.ErrNoRows
	}

	return annualJob, nil
}

func (r *annualJobRepository) GetAnnualJobsByClientID(clientID string, staffIDFilter string, isAdmin bool) ([]models.AnnualJob, error) { // <-- IMPLEMENTASI BARU
	query := `
	SELECT
		aj.job_id, aj.client_id, c.client_name, c.npwp_client, aj.job_year,
		aj.assigned_pic_staff_sigma_id, s.nama AS assigned_pic_staff_sigma_name,
		aj.overall_status,  aj.created_at, aj.updated_at,
		
		atr.report_id AS atr_report_id, atr.billing_code, atr.payment_date AS atr_payment_date,
		atr.payment_amount AS atr_payment_amount, atr.report_date AS atr_report_date, atr.report_status AS atr_report_status,
		atr.created_at AS atr_created_at, atr.updated_at AS atr_updated_at,

		adr.report_id AS adr_report_id, adr.is_reported, adr.report_date AS adr_report_date, adr.report_status AS adr_report_status,
		adr.created_at AS adr_created_at, adr.updated_at AS adr_updated_at
	FROM annual_jobs AS aj
	JOIN clients AS c ON aj.client_id = c.client_id
	LEFT JOIN staffs AS s ON aj.assigned_pic_staff_sigma_id = s.staff_id
	LEFT JOIN annual_tax_reports AS atr ON aj.job_id = atr.job_id
	LEFT JOIN annual_dividend_reports AS adr ON aj.job_id = adr.job_id
	WHERE aj.client_id = $1` // Filter berdasarkan client_id

	args := []interface{}{clientID} // Argumen pertama adalah clientID
	paramCounter := 2 // Argumen berikutnya dimulai dari $2

	if !isAdmin && staffIDFilter != "" { // Jika bukan admin DAN ada staffID yang difilter
		query += fmt.Sprintf(" AND aj.assigned_pic_staff_sigma_id = $%d", paramCounter)
		args = append(args, staffIDFilter)
		paramCounter++
	}

	query += " ORDER BY aj.job_year DESC, aj.created_at DESC;"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get annual jobs by client ID: %w", err)
	}
	defer rows.Close()

	// Bagian rekonstruksi map dan slice sama persis dengan GetAllAnnualJobs
	annualJobsMap := make(map[string]*models.AnnualJob)
	var annualJobsList []models.AnnualJob

	for rows.Next() {
		var (
			jobID                     string
			cID                       string // Ganti clientID dengan cID karena clientID sudah jadi parameter
			clientName                string
			npwpClient                string
			jobYear                   int
			assignedPicStaffSigmaID   sql.NullString
			assignedPicStaffSigmaName sql.NullString
			overallStatus             string
			jobCreatedAt              time.Time
			jobUpdatedAt              time.Time

			atrReportID       sql.NullString
			atrBillingCode    sql.NullString
			atrPaymentDate    sql.NullTime
			atrPaymentAmount  sql.NullFloat64
			atrReportDate     sql.NullTime
			atrReportStatus   sql.NullString
			atrCreatedAt      sql.NullTime
			atrUpdatedAt      sql.NullTime

			adrReportID       sql.NullString
			adrIsReported     sql.NullBool
			adrReportDate     sql.NullTime
			adrReportStatus   sql.NullString
			adrCreatedAt      sql.NullTime
			adrUpdatedAt      sql.NullTime
		)

		err := rows.Scan(
			&jobID, &cID, &clientName, &npwpClient, &jobYear,
			&assignedPicStaffSigmaID, &assignedPicStaffSigmaName,
			&overallStatus, &jobCreatedAt, &jobUpdatedAt,

			&atrReportID, &atrBillingCode, &atrPaymentDate, &atrPaymentAmount, &atrReportDate, &atrReportStatus,
			&atrCreatedAt, &atrUpdatedAt,

			&adrReportID, &adrIsReported, &adrReportDate, &adrReportStatus,
			&adrCreatedAt, &adrUpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan annual job row by client ID: %w", err)
		}

		job, ok := annualJobsMap[jobID]
		if !ok {
			job = &models.AnnualJob{
				JobID:           jobID,
				ClientID:        cID, // Gunakan cID di sini
				ClientName:      clientName,
				NpwpClient:      npwpClient,
				JobYear:         jobYear,
				OverallStatus:   overallStatus,
				CreatedAt:       jobCreatedAt,
				UpdatedAt:       jobUpdatedAt,
				TaxReports:      []models.AnnualTaxReport{},
				DividendReports: []models.AnnualDividendReport{},
			}
			if assignedPicStaffSigmaID.Valid {
				job.AssignedPicStaffSigmaID = assignedPicStaffSigmaID.String
			}
			if assignedPicStaffSigmaName.Valid {
				job.AssignedPicStaffSigmaName = assignedPicStaffSigmaName.String
			}
			annualJobsMap[jobID] = job
			annualJobsList = append(annualJobsList, *job)
		}

		if atrReportID.Valid {
			taxReport := models.AnnualTaxReport{
				ReportID:      atrReportID.String,
				JobID:         jobID,
				BillingCode:   atrBillingCode.String,
				ReportStatus:  atrReportStatus.String,
				CreatedAt:     atrCreatedAt.Time,
				UpdatedAt:     atrUpdatedAt.Time,
			}
			if atrPaymentDate.Valid {
				taxReport.PaymentDate = &atrPaymentDate.Time
			} else {
				taxReport.PaymentDate = nil
			}
			if atrPaymentAmount.Valid {
				taxReport.PaymentAmount = &atrPaymentAmount.Float64
			} else {
				taxReport.PaymentAmount = nil
			}
			if atrReportDate.Valid {
				taxReport.ReportDate = &atrReportDate.Time
			} else {
				taxReport.ReportDate = nil
			}
			job.TaxReports = append(job.TaxReports, taxReport)
		}

		if adrReportID.Valid {
			dividendReport := models.AnnualDividendReport{
				ReportID:      adrReportID.String,
				JobID:         jobID,
				IsReported:    adrIsReported.Bool,
				ReportStatus:  adrReportStatus.String,
				CreatedAt:     adrCreatedAt.Time,
				UpdatedAt:     adrUpdatedAt.Time,
			}
			if adrReportDate.Valid {
				dividendReport.ReportDate = &adrReportDate.Time
			} else {
				dividendReport.ReportDate = nil
			}
			job.DividendReports = append(job.DividendReports, dividendReport)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration for annual jobs by client ID: %w", err)
	}

	return annualJobsList, nil
}

// UpdateAnnualJob updates only the main fields of an annual job.
func (r *annualJobRepository) UpdateAnnualJob(job *models.AnnualJob) error {

	var assignedPicStaffSigmaID sql.NullString
	if job.AssignedPicStaffSigmaID != "" {
		assignedPicStaffSigmaID = sql.NullString{String: job.AssignedPicStaffSigmaID, Valid: true}
	} else {
		assignedPicStaffSigmaID = sql.NullString{Valid: false} // Akan menjadi NULL di database
	}

	 var proofOfWorkURL sql.NullString
    if job.ProofOfWorkURL != nil && *job.ProofOfWorkURL != "" { // Jika ada URL dan tidak kosong
        proofOfWorkURL = sql.NullString{String: *job.ProofOfWorkURL, Valid: true}
    } else {
        proofOfWorkURL = sql.NullString{Valid: false} // Akan menjadi NULL di DB jika nil atau string kosong
    }

	query := `UPDATE annual_jobs SET
		client_id = $1, job_year = $2, assigned_pic_staff_sigma_id = $3,
		overall_status = $4, proof_of_work_url = $5, updated_at = $6
	WHERE job_id = $7`

	job.UpdatedAt = time.Now()

	_, err := r.db.Exec(query,
		job.ClientID, job.JobYear, assignedPicStaffSigmaID,
		job.OverallStatus, proofOfWorkURL, job.UpdatedAt, job.JobID,
	)

	if err != nil {
		return fmt.Errorf("failed to update annual job: %w", err)
	}
	return nil
}

// CreateAnnualTaxReport inserts a new annual tax report for an existing annual job
func (r *annualJobRepository) CreateAnnualTaxReport(report *models.AnnualTaxReport) error {
	query := `INSERT INTO annual_tax_reports (
		job_id, billing_code, payment_date, payment_amount, report_date, report_status, created_at, updated_at
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8
	) RETURNING report_id, created_at, updated_at`

	if report.CreatedAt.IsZero() {
		report.CreatedAt = time.Now()
	}
	report.UpdatedAt = time.Now()

	err := r.db.QueryRow(query,
		report.JobID, report.BillingCode, report.PaymentDate, report.PaymentAmount,
		report.ReportDate, report.ReportStatus,
		report.CreatedAt, report.UpdatedAt,
	).Scan(&report.ReportID, &report.CreatedAt, &report.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create annual tax report: %w", err)
	}
	return nil
}

// GetAnnualTaxReportByID fetches a single annual tax report by its ID.
func (r *annualJobRepository) GetAnnualTaxReportByID(reportID string) (*models.AnnualTaxReport, error) {
	query := `SELECT
		report_id, job_id, billing_code, payment_date, payment_amount, report_date, report_status, created_at, updated_at
	FROM annual_tax_reports WHERE report_id = $1`

	var report models.AnnualTaxReport
	var paymentDate, reportDate sql.NullTime
	var paymentAmount sql.NullFloat64

	err := r.db.QueryRow(query, reportID).Scan(
		&report.ReportID, &report.JobID, &report.BillingCode, &paymentDate, &paymentAmount,
		&report.ReportDate, &report.ReportStatus, &report.CreatedAt, &report.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get annual tax report by ID: %w", err)
	}

	if paymentDate.Valid {
		report.PaymentDate = &paymentDate.Time
	} else {
		report.PaymentDate = nil
	}
	if paymentAmount.Valid {
		report.PaymentAmount = &paymentAmount.Float64
	} else {
		report.PaymentAmount = nil
	}
	if reportDate.Valid {
		report.ReportDate = &reportDate.Time
	} else {
		report.ReportDate = nil
	}

	return &report, nil
}

// UpdateAnnualTaxReport updates an existing annual tax report
func (r *annualJobRepository) UpdateAnnualTaxReport(report *models.AnnualTaxReport) error {
	query := `UPDATE annual_tax_reports SET
		billing_code = $1, payment_date = $2, payment_amount = $3, report_date = $4,
		report_status = $5, updated_at = $6
	WHERE report_id = $7`

	report.UpdatedAt = time.Now()

	_, err := r.db.Exec(query,
		report.BillingCode, report.PaymentDate, report.PaymentAmount, report.ReportDate,
		report.ReportStatus, report.UpdatedAt, report.ReportID,
	)

	if err != nil {
		return fmt.Errorf("failed to update annual tax report: %w", err)
	}
	return nil
}

// DeleteAnnualTaxReport deletes an annual tax report by its ID
func (r *annualJobRepository) DeleteAnnualTaxReport(reportID string) error {
	query := `DELETE FROM annual_tax_reports WHERE report_id = $1`
	_, err := r.db.Exec(query, reportID)
	if err != nil {
		return fmt.Errorf("failed to delete annual tax report: %w", err)
	}
	return nil
}

// CreateAnnualDividendReport inserts a new annual dividend report for an existing annual job
func (r *annualJobRepository) CreateAnnualDividendReport(report *models.AnnualDividendReport) error {
	query := `INSERT INTO annual_dividend_reports (
		job_id, is_reported, report_date, report_status, created_at, updated_at
	) VALUES (
		$1, $2, $3, $4, $5, $6
	) RETURNING report_id, created_at, updated_at`

	if report.CreatedAt.IsZero() {
		report.CreatedAt = time.Now()
	}
	report.UpdatedAt = time.Now()

	err := r.db.QueryRow(query,
		report.JobID, report.IsReported, report.ReportDate, report.ReportStatus,
		report.CreatedAt, report.UpdatedAt,
	).Scan(&report.ReportID, &report.CreatedAt, &report.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create annual dividend report: %w", err)
	}
	return nil
}

// GetAnnualDividendReportByID fetches a single annual dividend report by its ID.
func (r *annualJobRepository) GetAnnualDividendReportByID(reportID string) (*models.AnnualDividendReport, error) {
	query := `SELECT
		report_id, job_id, is_reported, report_date, report_status, created_at, updated_at
	FROM annual_dividend_reports WHERE report_id = $1`

	var report models.AnnualDividendReport
	var reportDate sql.NullTime
	var isReported sql.NullBool // Tambahkan ini

	err := r.db.QueryRow(query, reportID).Scan(
		&report.ReportID, &report.JobID, &isReported, &reportDate, &report.ReportStatus,
		&report.CreatedAt, &report.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get annual dividend report by ID: %w", err)
	}
	
	// Handle Nullable fields
	report.IsReported = isReported.Bool // Tetap isi boolean
	if reportDate.Valid {
		report.ReportDate = &reportDate.Time
	} else {
		report.ReportDate = nil
	}

	return &report, nil
}

// UpdateAnnualDividendReport updates an existing annual dividend report
func (r *annualJobRepository) UpdateAnnualDividendReport(report *models.AnnualDividendReport) error {
	query := `UPDATE annual_dividend_reports SET
		is_reported = $1, report_date = $2, report_status = $3, updated_at = $4
	WHERE report_id = $5`

	report.UpdatedAt = time.Now()

	_, err := r.db.Exec(query,
		report.IsReported, report.ReportDate, report.ReportStatus, report.UpdatedAt, report.ReportID,
	)

	if err != nil {
		return fmt.Errorf("failed to update annual dividend report: %w", err)
	}
	return nil
}

// DeleteAnnualDividendReport deletes an annual dividend report by its ID
func (r *annualJobRepository) DeleteAnnualDividendReport(reportID string) error {
	query := `DELETE FROM annual_dividend_reports WHERE report_id = $1`
	_, err := r.db.Exec(query, reportID)
	if err != nil {
		return fmt.Errorf("failed to delete annual dividend report: %w", err)
	}
	return nil
}
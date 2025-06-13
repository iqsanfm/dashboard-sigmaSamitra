package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/iqsanfm/dashboard-pekerjaan-backend/internal/models" // Pastikan ini adalah modul Anda
)

// MonthlyJobRepository defines the interface for monthly job data operations
type MonthlyJobRepository interface {
	GetMonthlyTaxReportByID(reportID string) (*models.MonthlyTaxReport, error)
	CreateMonthlyJob(job *models.MonthlyJob) error
	GetAllMonthlyJobs(staffIDFilter string, isAdmin bool) ([]models.MonthlyJob, error)
	GetMonthlyJobByID(id string, staffIDFilter string, isAdmin bool) (*models.MonthlyJob, error)
	UpdateMonthlyJob(job *models.MonthlyJob) error // For updating main job fields
	CreateMonthlyTaxReport(report *models.MonthlyTaxReport) error
	UpdateMonthlyTaxReport(report *models.MonthlyTaxReport) error
	DeleteMonthlyTaxReport(reportID string) error
	GetMonthlyJobsByClientID(clientID string, staffIDFilter string, isAdmin bool) ([]models.MonthlyJob, error)
}

// monthlyJobRepository implements MonthlyJobRepository interface
type monthlyJobRepository struct {
	db *sql.DB
}

// NewMonthlyJobRepository creates a new MonthlyJobRepository
func NewMonthlyJobRepository(db *sql.DB) MonthlyJobRepository {
	return &monthlyJobRepository{db: db}
}

// CreateMonthlyJob inserts a new monthly job and its associated tax reports into the database.
// This operation is wrapped in a transaction to ensure atomicity.
func (r *monthlyJobRepository) CreateMonthlyJob(job *models.MonthlyJob) error {
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

	// 1. Insert into monthly_jobs table
	// --- PERUBAHAN DI SINI: Gunakan assigned_pic_staff_sigma_id ---
	jobQuery := `INSERT INTO monthly_jobs (
		client_id, job_month, job_year, assigned_pic_staff_sigma_id, overall_status, created_at, updated_at
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7
	) RETURNING job_id, created_at, updated_at`

	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now()
	}
	job.UpdatedAt = time.Now()

	err = tx.QueryRow(jobQuery,
		job.ClientID, job.JobMonth, job.JobYear, assignedPicStaffSigmaID, job.OverallStatus, // Gunakan AssignedPicStaffSigmaID
		job.CreatedAt, job.UpdatedAt,
	).Scan(&job.JobID, &job.CreatedAt, &job.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create monthly job: %w", err)
	}

	// 2. Insert associated tax reports into monthly_tax_reports table
	if len(job.TaxReports) > 0 {
		reportQuery := `INSERT INTO monthly_tax_reports (
			job_id, tax_type, billing_code, payment_date, payment_amount, report_status, report_date, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		) RETURNING report_id, created_at, updated_at`

		for i := range job.TaxReports {
			report := &job.TaxReports[i] // Use pointer to update original slice element

			report.JobID = job.JobID // Link report to the newly created job
			if report.CreatedAt.IsZero() {
				report.CreatedAt = time.Now()
			}
			report.UpdatedAt = time.Now()

			err := tx.QueryRow(reportQuery,
				report.JobID, report.TaxType, report.BillingCode, report.PaymentDate,
				report.PaymentAmount, report.ReportStatus, report.ReportDate,
				report.CreatedAt, report.UpdatedAt,
			).Scan(&report.ReportID, &report.CreatedAt, &report.UpdatedAt)
			if err != nil {
				return fmt.Errorf("failed to create tax report for job %s: %w", job.JobID, err)
			}
		}
	}

	return tx.Commit() // Commit the transaction if all inserts are successful
}

// GetMonthlyTaxReportByID fetches a single monthly tax report by its ID.
func (r *monthlyJobRepository) GetMonthlyTaxReportByID(reportID string) (*models.MonthlyTaxReport, error) {
	query := `SELECT
		report_id, job_id, tax_type, billing_code, payment_date, payment_amount,
		report_status, report_date, created_at, updated_at
	FROM monthly_tax_reports WHERE report_id = $1`

	var report models.MonthlyTaxReport
	var paymentDate, reportDate sql.NullTime
	var paymentAmount sql.NullFloat64

	err := r.db.QueryRow(query, reportID).Scan(
		&report.ReportID, &report.JobID, &report.TaxType, &report.BillingCode, &paymentDate, &paymentAmount,
		&report.ReportStatus, &report.ReportDate, &report.CreatedAt, &report.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows // Explicitly return ErrNoRows if not found
		}
		return nil, fmt.Errorf("failed to get monthly tax report by ID: %w", err)
	}

	// Assign nullable fields if valid
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

// GetAllMonthlyJobs fetches all monthly jobs with their associated client and tax reports.
func (r *monthlyJobRepository) GetAllMonthlyJobs(staffIDFilter string, isAdmin bool) ([]models.MonthlyJob, error) {
	query := `
	SELECT
		mj.job_id, mj.client_id, c.client_name, c.npwp_client, mj.job_month, mj.job_year,
		mj.assigned_pic_staff_sigma_id, s.nama AS assigned_pic_staff_sigma_name,
		mj.overall_status, mj.proof_of_work_url,
		mj.created_at, mj.updated_at,
		
		mtr.report_id, mtr.tax_type, mtr.billing_code, mtr.payment_date, mtr.payment_amount, -- <-- URUTAN KOLOM MTR DI SINI
		mtr.report_status, mtr.report_date, mtr.created_at AS report_created_at, mtr.updated_at AS report_updated_at
	FROM monthly_jobs AS mj
	JOIN clients AS c ON mj.client_id = c.client_id
	LEFT JOIN staffs AS s ON mj.assigned_pic_staff_sigma_id = s.staff_id
	LEFT JOIN monthly_tax_reports AS mtr ON mj.job_id = mtr.job_id`

	args := []interface{}{}
	paramCounter := 1

	if !isAdmin && staffIDFilter != "" {
		query += fmt.Sprintf(" WHERE mj.assigned_pic_staff_sigma_id = $%d", paramCounter)
		args = append(args, staffIDFilter)
		paramCounter++
	}

	query += " ORDER BY mj.job_year DESC, mj.job_month DESC, mj.created_at DESC, mtr.tax_type ASC;"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get all monthly jobs: %w", err)
	}
	defer rows.Close()

	var monthlyJobsMap map[string]*models.MonthlyJob
	var monthlyJobsList []models.MonthlyJob

	for rows.Next() {
		var (
			jobID                     string
			clientID                  string
			clientName                string
			npwpClient                string
			jobMonth                  int
			jobYear                   int
			assignedPicStaffSigmaID   sql.NullString
			assignedPicStaffSigmaName sql.NullString
			overallStatus             string
			proofOfWorkURL            sql.NullString
			jobCreatedAt              time.Time
			jobUpdatedAt              time.Time

			// Variabel untuk kolom mtr, pastikan urutannya SAMA dengan SELECT
			reportID        sql.NullString
			taxType         sql.NullString
			billingCode     sql.NullString    // <-- Pastikan ini sql.NullString
			paymentDate     sql.NullTime      // <-- Pastikan ini sql.NullTime
			paymentAmount   sql.NullFloat64   // <-- Pastikan ini sql.NullFloat64
			reportStatus    sql.NullString
			reportDate      sql.NullTime
			reportCreatedAt sql.NullTime
			reportUpdatedAt sql.NullTime
		)

		err := rows.Scan(
			&jobID, &clientID, &clientName, &npwpClient, &jobMonth, &jobYear,
			&assignedPicStaffSigmaID, &assignedPicStaffSigmaName,
			&overallStatus, &proofOfWorkURL,
			&jobCreatedAt, &jobUpdatedAt,
			
			// --- PERBAIKAN PENTING DI SINI: PASTIKAN URUTANNYA SAMA PERSIS DENGAN SELECT MTR ---
			&reportID, &taxType, &billingCode, &paymentDate, &paymentAmount, // <-- URUTANNYA
			&reportStatus, &reportDate, &reportCreatedAt, &reportUpdatedAt,
			// --- AKHIR PERBAIKAN ---
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan monthly job row: %w", err)
		}

		job, ok := monthlyJobsMap[jobID]
		if !ok {
			job = &models.MonthlyJob{
				JobID:                 jobID,
				ClientID:              clientID,
				ClientName:            clientName,
				NpwpClient:            npwpClient,
				JobMonth:              jobMonth,
				JobYear:               jobYear,
				OverallStatus:         overallStatus,
				CreatedAt:             jobCreatedAt,
				UpdatedAt:             jobUpdatedAt,
				TaxReports:            []models.MonthlyTaxReport{},
			}
			if assignedPicStaffSigmaID.Valid {
				job.AssignedPicStaffSigmaID = assignedPicStaffSigmaID.String
			} else {
				job.AssignedPicStaffSigmaID = ""
			}
			if assignedPicStaffSigmaName.Valid {
				job.AssignedPicStaffSigmaName = assignedPicStaffSigmaName.String
			} else {
				job.AssignedPicStaffSigmaName = ""
			}
			if proofOfWorkURL.Valid {
				job.ProofOfWorkURL = &proofOfWorkURL.String
			} else {
				job.ProofOfWorkURL = nil
			}

			monthlyJobsList = append(monthlyJobsList, *job)
		}

		// Penambahan tax reports jika ada
		if reportID.Valid {
			report := models.MonthlyTaxReport{
				ReportID:      reportID.String,
				JobID:         jobID,
				TaxType:       taxType.String,
				BillingCode:   billingCode.String, // <-- Ini akan berisi string dari DB
				ReportStatus:  reportStatus.String,
				CreatedAt:     reportCreatedAt.Time,
				UpdatedAt:     reportUpdatedAt.Time,
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

			job.TaxReports = append(job.TaxReports, report)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return monthlyJobsList, nil
}

// GetMonthlyJobByID fetches a single monthly job by its ID with associated client and tax reports.
func (r *monthlyJobRepository) GetMonthlyJobByID(id string, staffIDFilter string, isAdmin bool) (*models.MonthlyJob, error) {
	query := `
	SELECT
		mj.job_id, mj.client_id, c.client_name, c.npwp_client, mj.job_month, mj.job_year,
		mj.assigned_pic_staff_sigma_id, s.nama AS assigned_pic_staff_sigma_name,
		mj.overall_status, mj.proof_of_work_url,
		mj.created_at, mj.updated_at,
		
		mtr.report_id, mtr.tax_type, mtr.billing_code, mtr.payment_date, mtr.payment_amount,
		mtr.report_status, mtr.report_date, mtr.created_at AS report_created_at, mtr.updated_at AS report_updated_at
	FROM monthly_jobs AS mj
	JOIN clients AS c ON mj.client_id = c.client_id
	LEFT JOIN staffs AS s ON mj.assigned_pic_staff_sigma_id = s.staff_id
	LEFT JOIN monthly_tax_reports AS mtr ON mj.job_id = mtr.job_id
	WHERE mj.job_id = $1`

	args := []interface{}{id}
	paramCounter := 2

	if !isAdmin && staffIDFilter != "" {
		query += fmt.Sprintf(" AND mj.assigned_pic_staff_sigma_id = $%d", paramCounter)
		args = append(args, staffIDFilter)
	}

	query += " ORDER BY mtr.tax_type ASC;"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly job by ID: %w", err)
	}
	defer rows.Close()

	var monthlyJob *models.MonthlyJob
	for rows.Next() {
		var (
			jobID                     string
			clientID                  string
			clientName                string
			npwpClient                string
			jobMonth                  int
			jobYear                   int
			assignedPicStaffSigmaID   sql.NullString
			assignedPicStaffSigmaName sql.NullString
			overallStatus             string
			proofOfWorkURL            sql.NullString
			jobCreatedAt              time.Time
			jobUpdatedAt              time.Time

			// Variabel untuk kolom mtr, pastikan urutannya SAMA dengan SELECT
			reportID        sql.NullString
			taxType         sql.NullString
			billingCode     sql.NullString
			paymentDate     sql.NullTime
			paymentAmount   sql.NullFloat64
			reportStatus    sql.NullString
			reportDate      sql.NullTime
			reportCreatedAt sql.NullTime
			reportUpdatedAt sql.NullTime
		)

		err := rows.Scan(
			&jobID, &clientID, &clientName, &npwpClient, &jobMonth, &jobYear,
			&assignedPicStaffSigmaID, &assignedPicStaffSigmaName,
			&overallStatus, &proofOfWorkURL,
			&jobCreatedAt, &jobUpdatedAt,
			
			// --- PERBAIKAN PENTING DI SINI: PASTIKAN URUTANNYA SAMA PERSIS DENGAN SELECT MTR ---
			&reportID, &taxType, &billingCode, &paymentDate, &paymentAmount, // <-- URUTANNYA
			&reportStatus, &reportDate, &reportCreatedAt, &reportUpdatedAt,
			// --- AKHIR PERBAIKAN ---
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan monthly job by ID row: %w", err)
		}

		if monthlyJob == nil {
			monthlyJob = &models.MonthlyJob{
				JobID:                 jobID,
				ClientID:              clientID,
				ClientName:            clientName,
				NpwpClient:            npwpClient,
				JobMonth:              jobMonth,
				JobYear:               jobYear,
				OverallStatus:         overallStatus,
				CreatedAt:             jobCreatedAt,
				UpdatedAt:             jobUpdatedAt,
				TaxReports:            []models.MonthlyTaxReport{},
			}
			if assignedPicStaffSigmaID.Valid {
				monthlyJob.AssignedPicStaffSigmaID = assignedPicStaffSigmaID.String
			} else {
				monthlyJob.AssignedPicStaffSigmaID = ""
			}
			if assignedPicStaffSigmaName.Valid {
				monthlyJob.AssignedPicStaffSigmaName = assignedPicStaffSigmaName.String
			} else {
				monthlyJob.AssignedPicStaffSigmaName = ""
			}
			if proofOfWorkURL.Valid {
				monthlyJob.ProofOfWorkURL = &proofOfWorkURL.String
			} else {
				monthlyJob.ProofOfWorkURL = nil
			}
		}

		// Penambahan tax reports jika ada
		if reportID.Valid {
			report := models.MonthlyTaxReport{
				ReportID:      reportID.String,
				JobID:         jobID,
				TaxType:       taxType.String,
				BillingCode:   billingCode.String, // <-- Ini akan berisi string dari DB
				ReportStatus:  reportStatus.String,
				CreatedAt:     reportCreatedAt.Time,
				UpdatedAt:     reportUpdatedAt.Time,
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

			monthlyJob.TaxReports = append(monthlyJob.TaxReports, report)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	if monthlyJob == nil {
		return nil, sql.ErrNoRows
	}

	return monthlyJob, nil
}

func (r *monthlyJobRepository) GetMonthlyJobsByClientID(clientID string, staffIDFilter string, isAdmin bool) ([]models.MonthlyJob, error) { // <-- IMPLEMENTASI BARU
	query := `
	SELECT
		mj.job_id, mj.client_id, c.client_name, c.npwp_client, mj.job_month, mj.job_year,
		mj.assigned_pic_staff_sigma_id, s.nama AS assigned_pic_staff_sigma_name,
		mj.overall_status, mj.created_at, mj.updated_at,
		
		mtr.report_id, mtr.tax_type, mtr.billing_code, mtr.payment_date, mtr.payment_amount,
		mtr.report_status, mtr.report_date, mtr.created_at AS report_created_at, mtr.updated_at AS report_updated_at
	FROM monthly_jobs AS mj
	JOIN clients AS c ON mj.client_id = c.client_id
	LEFT JOIN staffs AS s ON mj.assigned_pic_staff_sigma_id = s.staff_id
	LEFT JOIN monthly_tax_reports AS mtr ON mj.job_id = mtr.job_id
	WHERE mj.client_id = $1` // Filter berdasarkan client_id

	args := []interface{}{clientID} // Argumen pertama adalah clientID
	paramCounter := 2 // Argumen berikutnya dimulai dari $2

	if !isAdmin && staffIDFilter != "" { // Jika bukan admin DAN ada staffID yang difilter
		query += fmt.Sprintf(" AND mj.assigned_pic_staff_sigma_id = $%d", paramCounter)
		args = append(args, staffIDFilter)
		paramCounter++ // Jika ada filter staf, counter argumen bertambah
	}

	query += " ORDER BY mj.job_year DESC, mj.job_month DESC, mj.created_at DESC, mtr.tax_type ASC;"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly jobs by client ID: %w", err)
	}
	defer rows.Close()

	// Bagian rekonstruksi map dan slice sama persis dengan GetAllMonthlyJobs
	monthlyJobsMap := make(map[string]*models.MonthlyJob)
	var monthlyJobsList []models.MonthlyJob

	for rows.Next() {
		var (
			jobID                     string
			cID                       string // Ganti clientID dengan cID karena clientID sudah jadi parameter
			clientName                string
			npwpClient                string
			jobMonth                  int
			jobYear                   int
			assignedPicStaffSigmaID   sql.NullString
			assignedPicStaffSigmaName sql.NullString
			overallStatus             string
			jobCreatedAt              time.Time
			jobUpdatedAt              time.Time

			reportID        sql.NullString
			taxType         sql.NullString
			billingCode     sql.NullString
			paymentDate     sql.NullTime
			paymentAmount   sql.NullFloat64
			reportStatus    sql.NullString
			reportDate      sql.NullTime
			reportCreatedAt sql.NullTime
			reportUpdatedAt sql.NullTime
		)

		err := rows.Scan(
			&jobID, &cID, &clientName, &npwpClient, &jobMonth, &jobYear,
			&assignedPicStaffSigmaID, &assignedPicStaffSigmaName,
			&overallStatus, &jobCreatedAt, &jobUpdatedAt,
			&reportID, &taxType, &billingCode, &paymentDate, &paymentAmount,
			&reportStatus, &reportDate, &reportCreatedAt, &reportUpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan monthly job row by client ID: %w", err)
		}

		job, ok := monthlyJobsMap[jobID]
		if !ok {
			job = &models.MonthlyJob{
				JobID:                 jobID,
				ClientID:              cID, // Gunakan cID di sini
				ClientName:            clientName,
				NpwpClient:            npwpClient,
				JobMonth:              jobMonth,
				JobYear:               jobYear,
				OverallStatus:         overallStatus,
				CreatedAt:             jobCreatedAt,
				UpdatedAt:             jobUpdatedAt,
				TaxReports:            []models.MonthlyTaxReport{},
			}
			if assignedPicStaffSigmaID.Valid {
				job.AssignedPicStaffSigmaID = assignedPicStaffSigmaID.String
			}
			if assignedPicStaffSigmaName.Valid {
				job.AssignedPicStaffSigmaName = assignedPicStaffSigmaName.String
			}
			monthlyJobsMap[jobID] = job
			monthlyJobsList = append(monthlyJobsList, *job)
		}

		if reportID.Valid {
			report := models.MonthlyTaxReport{
				ReportID:      reportID.String,
				JobID:         jobID,
				TaxType:       taxType.String,
				BillingCode:   billingCode.String,
				PaymentDate:   &paymentDate.Time,
				PaymentAmount: &paymentAmount.Float64,
				ReportStatus:  reportStatus.String,
				ReportDate:    &reportDate.Time,
				CreatedAt:     reportCreatedAt.Time,
				UpdatedAt:     reportUpdatedAt.Time,
			}
			if !paymentDate.Valid { report.PaymentDate = nil }
			if !paymentAmount.Valid { report.PaymentAmount = nil }
			if !reportDate.Valid { report.ReportDate = nil }

			job.TaxReports = append(job.TaxReports, report)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration for monthly jobs by client ID: %w", err)
	}

	return monthlyJobsList, nil
}

// UpdateMonthlyJob updates only the main fields of a monthly job.
// Tax reports are handled via separate functions if partial updates are needed.
func (r *monthlyJobRepository) UpdateMonthlyJob(job *models.MonthlyJob) error {
	// --- PERUBAHAN DI SINI: Gunakan assigned_pic_staff_sigma_id ---
	var assignedPicStaffSigmaID sql.NullString
	if job.AssignedPicStaffSigmaID != "" {
		assignedPicStaffSigmaID = sql.NullString{String: job.AssignedPicStaffSigmaID, Valid: true}
	} else {
		assignedPicStaffSigmaID = sql.NullString{Valid: false} // Akan menjadi NULL di database
	}

	var proofOfWorkURL sql.NullString
	if job.ProofOfWorkURL != nil && *job.ProofOfWorkURL != "" {
		proofOfWorkURL = sql.NullString{String: *job.ProofOfWorkURL, Valid: true}
	} else {
		proofOfWorkURL = sql.NullString{Valid: false}
	}

 var proofOfWorkURLValue interface{}
    if proofOfWorkURL.Valid {
        proofOfWorkURLValue = proofOfWorkURL.String // Gunakan string jika valid
    } else {
        proofOfWorkURLValue = nil // Gunakan nil jika tidak valid (untuk NULL di DB)
    }

	query := `UPDATE monthly_jobs SET
		client_id = $1, job_month = $2, job_year = $3, assigned_pic_staff_sigma_id = $4,
		overall_status = $5, proof_of_work_url = $6, updated_at = $7
	WHERE job_id = $8`

	job.UpdatedAt = time.Now()

	_, err := r.db.Exec(query,
		job.ClientID, job.JobMonth, job.JobYear, assignedPicStaffSigmaID, // Gunakan AssignedPicStaffSigmaID
		job.OverallStatus, proofOfWorkURLValue, job.UpdatedAt, job.JobID,
	)

	if err != nil {
		return fmt.Errorf("failed to update monthly job: %w", err)
	}
	return nil
}

// CreateMonthlyTaxReport inserts a new tax report for an existing monthly job
func (r *monthlyJobRepository) CreateMonthlyTaxReport(report *models.MonthlyTaxReport) error {
	query := `INSERT INTO monthly_tax_reports (
		job_id, tax_type, billing_code, payment_date, payment_amount, report_status, report_date, created_at, updated_at
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9
	) RETURNING report_id, created_at, updated_at`

	if report.CreatedAt.IsZero() {
		report.CreatedAt = time.Now()
	}
	report.UpdatedAt = time.Now()

	err := r.db.QueryRow(query,
		report.JobID, report.TaxType, report.BillingCode, report.PaymentDate,
		report.PaymentAmount, report.ReportStatus, report.ReportDate,
		report.CreatedAt, report.UpdatedAt,
	).Scan(&report.ReportID, &report.CreatedAt, &report.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create monthly tax report: %w", err)
	}
	return nil
}

// UpdateMonthlyTaxReport updates an existing monthly tax report
func (r *monthlyJobRepository) UpdateMonthlyTaxReport(report *models.MonthlyTaxReport) error {
	query := `UPDATE monthly_tax_reports SET
		tax_type = $1, billing_code = $2, payment_date = $3, payment_amount = $4,
		report_status = $5, report_date = $6, updated_at = $7
	WHERE report_id = $8`

	report.UpdatedAt = time.Now()

	_, err := r.db.Exec(query,
		report.TaxType, report.BillingCode, report.PaymentDate, report.PaymentAmount,
		report.ReportStatus, report.ReportDate, report.UpdatedAt, report.ReportID,
	)

	if err != nil {
		return fmt.Errorf("failed to update monthly tax report: %w", err)
	}
	return nil
}

// DeleteMonthlyTaxReport deletes a monthly tax report by its ID
func (r *monthlyJobRepository) DeleteMonthlyTaxReport(reportID string) error {
	query := `DELETE FROM monthly_tax_reports WHERE report_id = $1`
	_, err := r.db.Exec(query, reportID)
	if err != nil {
		return fmt.Errorf("failed to delete monthly tax report: %w", err)
	}
	return nil
}
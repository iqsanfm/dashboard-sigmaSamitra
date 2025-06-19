package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/iqsanfm/dashboard-pekerjaan-backend/internal/models"
)

// Sp2dkJobRepository defines the interface for SP2DK job data operations
type Sp2dkJobRepository interface {
	CreateSp2dkJob(job *models.Sp2dkJob) error
	GetAllSp2dkJobs(staffIDFilter string, isAdmin bool) ([]models.Sp2dkJob, error)
	GetSp2dkJobByID(id string, staffIDFilter string, isAdmin bool) (*models.Sp2dkJob, error)
	UpdateSp2dkJob(job *models.Sp2dkJob) error
	DeleteSp2dkJob(id string) error
	GetSp2dkJobsByClientID(clientID string, staffIDFilter string, isAdmin bool) ([]models.Sp2dkJob, error)
}

// sp2dkJobRepository implements Sp2dkJobRepository interface
type sp2dkJobRepository struct {
	db *sql.DB
}

// NewSp2dkJobRepository creates a new Sp2dkJobRepository
func NewSp2dkJobRepository(db *sql.DB) Sp2dkJobRepository {
	return &sp2dkJobRepository{db: db}
}

// CreateSp2dkJob inserts a new SP2DK job into the database.
func (r *sp2dkJobRepository) CreateSp2dkJob(job *models.Sp2dkJob) error {
	// PERBAIKAN: Menggunakan `overall_status` di query dan `job.OverallStatus` sebagai parameter
	query := `INSERT INTO sp2dk_jobs (
		client_id, assigned_pic_staff_sigma_id, contract_no, contract_date, sp2dk_no, sp2dk_date,
		bap2dk_no, bap2dk_date, payment_date, report_date, overall_status, created_at, updated_at
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
	) RETURNING job_id, created_at, updated_at`

	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now()
	}
	job.UpdatedAt = time.Now()

	err := r.db.QueryRow(query,
		job.ClientID, job.AssignedPicStaffSigmaID, job.ContractNo, job.ContractDate, job.Sp2dkNo, job.Sp2dkDate,
		job.Bap2dkNo, job.Bap2dkDate, job.PaymentDate, job.ReportDate, job.OverallStatus,
		job.CreatedAt, job.UpdatedAt,
	).Scan(&job.JobID, &job.CreatedAt, &job.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create SP2DK job: %w", err)
	}
	return nil
}

// GetAllSp2dkJobs fetches all SP2DK jobs.
func (r *sp2dkJobRepository) GetAllSp2dkJobs(staffIDFilter string, isAdmin bool) ([]models.Sp2dkJob, error) {
	// PERBAIKAN: Menggunakan `overall_status` di SELECT statement
	query := `
	SELECT
		sj.job_id, sj.client_id, c.client_name, c.npwp_client,
		sj.assigned_pic_staff_sigma_id, s.nama AS assigned_pic_staff_sigma_name,
		sj.contract_no, sj.contract_date, sj.sp2dk_no, sj.sp2dk_date, sj.bap2dk_no, sj.bap2dk_date,
		sj.payment_date, sj.report_date, sj.overall_status, sj.proof_of_work_url, sj.created_at, sj.updated_at
	FROM sp2dk_jobs AS sj
	JOIN clients AS c ON sj.client_id = c.client_id
	LEFT JOIN staffs AS s ON sj.assigned_pic_staff_sigma_id = s.staff_id`

	args := []interface{}{}
	if !isAdmin && staffIDFilter != "" {
		query += " WHERE sj.assigned_pic_staff_sigma_id = $1"
		args = append(args, staffIDFilter)
	}
	query += " ORDER BY sj.created_at DESC;"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get all SP2DK jobs: %w", err)
	}
	defer rows.Close()

	var sp2dkJobs []models.Sp2dkJob
	for rows.Next() {
		var job models.Sp2dkJob
		// PERBAIKAN: Menyederhanakan Scan langsung ke field model
		err := rows.Scan(
			&job.JobID, &job.ClientID, &job.ClientName, &job.NpwpClient,
			&job.AssignedPicStaffSigmaID, &job.AssignedPicStaffSigmaName,
			&job.ContractNo, &job.ContractDate, &job.Sp2dkNo, &job.Sp2dkDate, &job.Bap2dkNo, &job.Bap2dkDate,
			&job.PaymentDate, &job.ReportDate, &job.OverallStatus, &job.ProofOfWorkURL, &job.CreatedAt, &job.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan SP2DK job row: %w", err)
		}
		sp2dkJobs = append(sp2dkJobs, job)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}
	return sp2dkJobs, nil
}

// GetSp2dkJobByID fetches a single SP2DK job by its ID.
func (r *sp2dkJobRepository) GetSp2dkJobByID(id string, staffIDFilter string, isAdmin bool) (*models.Sp2dkJob, error) {
	// PERBAIKAN: Menggunakan `overall_status` di SELECT statement
	query := `
	SELECT
		sj.job_id, sj.client_id, c.client_name, c.npwp_client,
		sj.assigned_pic_staff_sigma_id, s.nama AS assigned_pic_staff_sigma_name,
		sj.contract_no, sj.contract_date, sj.sp2dk_no, sj.sp2dk_date, sj.bap2dk_no, sj.bap2dk_date,
		sj.payment_date, sj.report_date, sj.overall_status, sj.proof_of_work_url, sj.created_at, sj.updated_at
	FROM sp2dk_jobs AS sj
	JOIN clients AS c ON sj.client_id = c.client_id
	LEFT JOIN staffs AS s ON sj.assigned_pic_staff_sigma_id = s.staff_id
	WHERE sj.job_id = $1`

	args := []interface{}{id}
	if !isAdmin && staffIDFilter != "" {
		query += " AND sj.assigned_pic_staff_sigma_id = $2"
		args = append(args, staffIDFilter)
	}

	var job models.Sp2dkJob
	// PERBAIKAN: Menyederhanakan Scan langsung ke field model
	err := r.db.QueryRow(query, args...).Scan(
		&job.JobID, &job.ClientID, &job.ClientName, &job.NpwpClient,
		&job.AssignedPicStaffSigmaID, &job.AssignedPicStaffSigmaName,
		&job.ContractNo, &job.ContractDate, &job.Sp2dkNo, &job.Sp2dkDate, &job.Bap2dkNo, &job.Bap2dkDate,
		&job.PaymentDate, &job.ReportDate, &job.OverallStatus, &job.ProofOfWorkURL, &job.CreatedAt, &job.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get SP2DK job by ID: %w", err)
	}
	return &job, nil
}

// UpdateSp2dkJob updates an existing SP2DK job.
func (r *sp2dkJobRepository) UpdateSp2dkJob(job *models.Sp2dkJob) error {
	// PERBAIKAN: Menggunakan `overall_status` dan memperbaiki nomor parameter
	query := `UPDATE sp2dk_jobs SET
		client_id = $1, assigned_pic_staff_sigma_id = $2, contract_no = $3, contract_date = $4,
		sp2dk_no = $5, sp2dk_date = $6, bap2dk_no = $7, bap2dk_date = $8,
		payment_date = $9, report_date = $10, overall_status = $11, proof_of_work_url = $12, updated_at = $13
	WHERE job_id = $14`

	job.UpdatedAt = time.Now()

	_, err := r.db.Exec(query,
		job.ClientID, job.AssignedPicStaffSigmaID, job.ContractNo, job.ContractDate,
		job.Sp2dkNo, job.Sp2dkDate, job.Bap2dkNo, job.Bap2dkDate,
		job.PaymentDate, job.ReportDate, job.OverallStatus, job.ProofOfWorkURL, job.UpdatedAt, job.JobID,
	)

	if err != nil {
		return fmt.Errorf("failed to update SP2DK job: %w", err)
	}
	return nil
}

// DeleteSp2dkJob deletes an SP2DK job by its ID.
func (r *sp2dkJobRepository) DeleteSp2dkJob(id string) error {
	query := `DELETE FROM sp2dk_jobs WHERE job_id = $1`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete SP2DK job: %w", err)
	}
	return nil
}


// GetSp2dkJobsByClientID fetches all SP2DK jobs for a specific client.
func (r *sp2dkJobRepository) GetSp2dkJobsByClientID(clientID string, staffIDFilter string, isAdmin bool) ([]models.Sp2dkJob, error) {
	// PERBAIKAN: Menggunakan `overall_status` di SELECT statement
	query := `
	SELECT
		sj.job_id, sj.client_id, c.client_name, c.npwp_client,
		sj.assigned_pic_staff_sigma_id, s.nama AS assigned_pic_staff_sigma_name,
		sj.contract_no, sj.contract_date, sj.sp2dk_no, sj.sp2dk_date, sj.bap2dk_no, sj.bap2dk_date,
		sj.payment_date, sj.report_date, sj.overall_status, sj.proof_of_work_url, sj.created_at, sj.updated_at
	FROM sp2dk_jobs AS sj
	JOIN clients AS c ON sj.client_id = c.client_id
	LEFT JOIN staffs AS s ON sj.assigned_pic_staff_sigma_id = s.staff_id
	WHERE sj.client_id = $1`

	args := []interface{}{clientID}
	if !isAdmin && staffIDFilter != "" {
		query += " AND sj.assigned_pic_staff_sigma_id = $2"
		args = append(args, staffIDFilter)
	}
	query += " ORDER BY sj.created_at DESC;"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get SP2DK jobs by client ID: %w", err)
	}
	defer rows.Close()

	var sp2dkJobs []models.Sp2dkJob
	for rows.Next() {
		var job models.Sp2dkJob
        // PERBAIKAN: Menyederhanakan Scan langsung ke field model
		err := rows.Scan(
			&job.JobID, &job.ClientID, &job.ClientName, &job.NpwpClient,
			&job.AssignedPicStaffSigmaID, &job.AssignedPicStaffSigmaName,
			&job.ContractNo, &job.ContractDate, &job.Sp2dkNo, &job.Sp2dkDate, &job.Bap2dkNo, &job.Bap2dkDate,
			&job.PaymentDate, &job.ReportDate, &job.OverallStatus, &job.ProofOfWorkURL, &job.CreatedAt, &job.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan SP2DK job row by client ID: %w", err)
		}
		sp2dkJobs = append(sp2dkJobs, job)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration for SP2DK jobs by client ID: %w", err)
	}
	return sp2dkJobs, nil
}

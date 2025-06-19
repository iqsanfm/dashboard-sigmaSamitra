package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/iqsanfm/dashboard-pekerjaan-backend/internal/models"
)

// PemeriksaanJobRepository defines the interface for Pemeriksaan job data operations
type PemeriksaanJobRepository interface {
	CreatePemeriksaanJob(job *models.PemeriksaanJob) error
	GetAllPemeriksaanJobs(staffIDFilter string, isAdmin bool) ([]models.PemeriksaanJob, error)
	GetPemeriksaanJobByID(id string, staffIDFilter string, isAdmin bool) (*models.PemeriksaanJob, error)
	UpdatePemeriksaanJob(job *models.PemeriksaanJob) error
	DeletePemeriksaanJob(id string) error
	GetPemeriksaanJobsByClientID(clientID string, staffIDFilter string, isAdmin bool) ([]models.PemeriksaanJob, error)
}

// pemeriksaanJobRepository implements PemeriksaanJobRepository interface
type pemeriksaanJobRepository struct {
	db *sql.DB
}

// NewPemeriksaanJobRepository creates a new PemeriksaanJobRepository
func NewPemeriksaanJobRepository(db *sql.DB) PemeriksaanJobRepository {
	return &pemeriksaanJobRepository{db: db}
}

// CreatePemeriksaanJob inserts a new Pemeriksaan job into the database.
func (r *pemeriksaanJobRepository) CreatePemeriksaanJob(job *models.PemeriksaanJob) error {
	// PERBAIKAN: Menggunakan `overall_status` di query dan `job.OverallStatus` sebagai parameter
	query := `INSERT INTO pemeriksaan_jobs (
		client_id, assigned_pic_staff_sigma_id, contract_no, contract_date, sp2_no, sp2_date,
		skp_no, skp_date, overall_status, created_at, updated_at
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
	) RETURNING job_id, created_at, updated_at`

	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now()
	}
	job.UpdatedAt = time.Now()

	err := r.db.QueryRow(query,
		job.ClientID, job.AssignedPicStaffSigmaID, job.ContractNo, job.ContractDate, job.Sp2No, job.Sp2Date,
		job.SkpNo, job.SkpDate, job.OverallStatus,
		job.CreatedAt, job.UpdatedAt,
	).Scan(&job.JobID, &job.CreatedAt, &job.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create Pemeriksaan job: %w", err)
	}
	return nil
}

// GetAllPemeriksaanJobs fetches all Pemeriksaan jobs.
func (r *pemeriksaanJobRepository) GetAllPemeriksaanJobs(staffIDFilter string, isAdmin bool) ([]models.PemeriksaanJob, error) {
	// PERBAIKAN: Menggunakan `overall_status` di SELECT statement
	query := `
	SELECT
		pj.job_id, pj.client_id, c.client_name, c.npwp_client,
		pj.assigned_pic_staff_sigma_id, s.nama AS assigned_pic_staff_sigma_name,
		pj.contract_no, pj.contract_date, pj.sp2_no, pj.sp2_date, pj.skp_no, pj.skp_date,
		pj.overall_status, pj.proof_of_work_url, pj.created_at, pj.updated_at
	FROM pemeriksaan_jobs AS pj
	JOIN clients AS c ON pj.client_id = c.client_id
	LEFT JOIN staffs AS s ON pj.assigned_pic_staff_sigma_id = s.staff_id`

	args := []interface{}{}
	if !isAdmin && staffIDFilter != "" {
		query += " WHERE pj.assigned_pic_staff_sigma_id = $1"
		args = append(args, staffIDFilter)
	}
	query += " ORDER BY pj.created_at DESC;"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get all Pemeriksaan jobs: %w", err)
	}
	defer rows.Close()

	var pemeriksaanJobs []models.PemeriksaanJob
	for rows.Next() {
		var job models.PemeriksaanJob
		// PERBAIKAN: Menyederhanakan Scan langsung ke field model
		err := rows.Scan(
			&job.JobID, &job.ClientID, &job.ClientName, &job.NpwpClient,
			&job.AssignedPicStaffSigmaID, &job.AssignedPicStaffSigmaName,
			&job.ContractNo, &job.ContractDate, &job.Sp2No, &job.Sp2Date, &job.SkpNo, &job.SkpDate,
			&job.OverallStatus, &job.ProofOfWorkURL, &job.CreatedAt, &job.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan Pemeriksaan job row: %w", err)
		}
		pemeriksaanJobs = append(pemeriksaanJobs, job)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration for Pemeriksaan jobs: %w", err)
	}
	return pemeriksaanJobs, nil
}

// GetPemeriksaanJobByID fetches a single Pemeriksaan job by its ID.
func (r *pemeriksaanJobRepository) GetPemeriksaanJobByID(id string, staffIDFilter string, isAdmin bool) (*models.PemeriksaanJob, error) {
	// PERBAIKAN: Menggunakan `overall_status` di SELECT statement
	query := `
	SELECT
		pj.job_id, pj.client_id, c.client_name, c.npwp_client,
		pj.assigned_pic_staff_sigma_id, s.nama AS assigned_pic_staff_sigma_name,
		pj.contract_no, pj.contract_date, pj.sp2_no, pj.sp2_date, pj.skp_no, pj.skp_date,
		pj.overall_status, pj.proof_of_work_url, pj.created_at, pj.updated_at
	FROM pemeriksaan_jobs AS pj
	JOIN clients AS c ON pj.client_id = c.client_id
	LEFT JOIN staffs AS s ON pj.assigned_pic_staff_sigma_id = s.staff_id
	WHERE pj.job_id = $1`

	args := []interface{}{id}
	if !isAdmin && staffIDFilter != "" {
		query += " AND pj.assigned_pic_staff_sigma_id = $2"
		args = append(args, staffIDFilter)
	}

	var job models.PemeriksaanJob
	// PERBAIKAN: Menyederhanakan Scan langsung ke field model
	err := r.db.QueryRow(query, args...).Scan(
		&job.JobID, &job.ClientID, &job.ClientName, &job.NpwpClient,
		&job.AssignedPicStaffSigmaID, &job.AssignedPicStaffSigmaName,
		&job.ContractNo, &job.ContractDate, &job.Sp2No, &job.Sp2Date, &job.SkpNo, &job.SkpDate,
		&job.OverallStatus, &job.ProofOfWorkURL, &job.CreatedAt, &job.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Return nil, nil jika tidak ditemukan
		}
		return nil, fmt.Errorf("failed to get Pemeriksaan job by ID: %w", err)
	}
	return &job, nil
}

// UpdatePemeriksaanJob updates an existing Pemeriksaan job.
func (r *pemeriksaanJobRepository) UpdatePemeriksaanJob(job *models.PemeriksaanJob) error {
	// PERBAIKAN: Menggunakan `overall_status`
	query := `UPDATE pemeriksaan_jobs SET
		client_id = $1, assigned_pic_staff_sigma_id = $2, contract_no = $3, contract_date = $4,
		sp2_no = $5, sp2_date = $6, skp_no = $7, skp_date = $8,
		overall_status = $9, proof_of_work_url = $10, updated_at = $11
	WHERE job_id = $12`

	job.UpdatedAt = time.Now()

	_, err := r.db.Exec(query,
		job.ClientID, job.AssignedPicStaffSigmaID, job.ContractNo, job.ContractDate,
		job.Sp2No, job.Sp2Date, job.SkpNo, job.SkpDate,
		job.OverallStatus, job.ProofOfWorkURL, job.UpdatedAt, job.JobID,
	)

	if err != nil {
		return fmt.Errorf("failed to update Pemeriksaan job: %w", err)
	}
	return nil
}

// DeletePemeriksaanJob deletes a Pemeriksaan job by its ID.
func (r *pemeriksaanJobRepository) DeletePemeriksaanJob(id string) error {
	query := `DELETE FROM pemeriksaan_jobs WHERE job_id = $1`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete Pemeriksaan job: %w", err)
	}
	return nil
}

// GetPemeriksaanJobsByClientID fetches all Pemeriksaan jobs for a specific client.
func (r *pemeriksaanJobRepository) GetPemeriksaanJobsByClientID(clientID string, staffIDFilter string, isAdmin bool) ([]models.PemeriksaanJob, error) {
	// PERBAIKAN: Menggunakan `overall_status` di SELECT statement
	query := `
	SELECT
		pj.job_id, pj.client_id, c.client_name, c.npwp_client,
		pj.assigned_pic_staff_sigma_id, s.nama AS assigned_pic_staff_sigma_name,
		pj.contract_no, pj.contract_date, pj.sp2_no, pj.sp2_date, pj.skp_no, pj.skp_date,
		pj.overall_status, pj.proof_of_work_url, pj.created_at, pj.updated_at
	FROM pemeriksaan_jobs AS pj
	JOIN clients AS c ON pj.client_id = c.client_id
	LEFT JOIN staffs AS s ON pj.assigned_pic_staff_sigma_id = s.staff_id
	WHERE pj.client_id = $1`

	args := []interface{}{clientID}
	if !isAdmin && staffIDFilter != "" {
		query += " AND pj.assigned_pic_staff_sigma_id = $2"
		args = append(args, staffIDFilter)
	}
	query += " ORDER BY pj.created_at DESC;"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get Pemeriksaan jobs by client ID: %w", err)
	}
	defer rows.Close()

	var pemeriksaanJobs []models.PemeriksaanJob
	for rows.Next() {
		var job models.PemeriksaanJob
		// PERBAIKAN: Menyederhanakan Scan langsung ke field model
		err := rows.Scan(
			&job.JobID, &job.ClientID, &job.ClientName, &job.NpwpClient,
			&job.AssignedPicStaffSigmaID, &job.AssignedPicStaffSigmaName,
			&job.ContractNo, &job.ContractDate, &job.Sp2No, &job.Sp2Date, &job.SkpNo, &job.SkpDate,
			&job.OverallStatus, &job.ProofOfWorkURL, &job.CreatedAt, &job.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan Pemeriksaan job row by client ID: %w", err)
		}
		pemeriksaanJobs = append(pemeriksaanJobs, job)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration for Pemeriksaan jobs by client ID: %w", err)
	}
	return pemeriksaanJobs, nil
}

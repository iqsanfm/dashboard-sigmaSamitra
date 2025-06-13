package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/iqsanfm/dashboard-pekerjaan-backend/internal/models" // Pastikan ini adalah modul Anda
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
	// Handle nullable fields for UUIDs
	var assignedPicStaffSigmaID sql.NullString
	if job.AssignedPicStaffSigmaID != "" {
		assignedPicStaffSigmaID = sql.NullString{String: job.AssignedPicStaffSigmaID, Valid: true}
	} else {
		assignedPicStaffSigmaID = sql.NullString{Valid: false}
	}

	// Handle nullable date/string fields
	var contractDate sql.NullTime
	if job.ContractDate != nil {
		contractDate = sql.NullTime{Time: *job.ContractDate, Valid: true}
	} else {
		contractDate = sql.NullTime{Valid: false}
	}

	var sp2dkDate sql.NullTime
	if job.Sp2dkDate != nil {
		sp2dkDate = sql.NullTime{Time: *job.Sp2dkDate, Valid: true}
	} else {
		sp2dkDate = sql.NullTime{Valid: false}
	}

	var bap2dkDate sql.NullTime
	if job.Bap2dkDate != nil {
		bap2dkDate = sql.NullTime{Time: *job.Bap2dkDate, Valid: true}
	} else {
		bap2dkDate = sql.NullTime{Valid: false}
	}

	var paymentDate sql.NullTime
	if job.PaymentDate != nil {
		paymentDate = sql.NullTime{Time: *job.PaymentDate, Valid: true}
	} else {
		paymentDate = sql.NullTime{Valid: false}
	}

	var reportDate sql.NullTime
	if job.ReportDate != nil {
		reportDate = sql.NullTime{Time: *job.ReportDate, Valid: true}
	} else {
		reportDate = sql.NullTime{Valid: false}
	}

	var contractNo sql.NullString
	if job.ContractNo != "" {
		contractNo = sql.NullString{String: job.ContractNo, Valid: true}
	} else {
		contractNo = sql.NullString{Valid: false}
	}

	var sp2dkNo sql.NullString
	if job.Sp2dkNo != "" {
		sp2dkNo = sql.NullString{String: job.Sp2dkNo, Valid: true}
	} else {
		sp2dkNo = sql.NullString{Valid: false}
	}

	var bap2dkNo sql.NullString
	if job.Bap2dkNo != "" {
		bap2dkNo = sql.NullString{String: job.Bap2dkNo, Valid: true}
	} else {
		bap2dkNo = sql.NullString{Valid: false}
	}


	query := `INSERT INTO sp2dk_jobs (
		client_id, assigned_pic_staff_sigma_id, contract_no, contract_date, sp2dk_no, sp2dk_date,
		bap2dk_no, bap2dk_date, payment_date, report_date, job_status, created_at, updated_at
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
	) RETURNING job_id, created_at, updated_at`

	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now()
	}
	job.UpdatedAt = time.Now()

	err := r.db.QueryRow(query,
		job.ClientID, assignedPicStaffSigmaID, contractNo, contractDate, sp2dkNo, sp2dkDate,
		bap2dkNo, bap2dkDate, paymentDate, reportDate, job.JobStatus,
		job.CreatedAt, job.UpdatedAt,
	).Scan(&job.JobID, &job.CreatedAt, &job.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create SP2DK job: %w", err)
	}
	return nil
}

func (r *sp2dkJobRepository) GetSp2dkJobsByClientID(clientID string, staffIDFilter string, isAdmin bool) ([]models.Sp2dkJob, error) { // <-- IMPLEMENTASI BARU
	query := `
	SELECT
		sj.job_id, sj.client_id, c.client_name, c.npwp_client,
		sj.assigned_pic_staff_sigma_id, s.nama AS assigned_pic_staff_sigma_name,
		sj.contract_no, sj.contract_date, sj.sp2dk_no, sj.sp2dk_date, sj.bap2dk_no, sj.bap2dk_date,
		sj.payment_date, sj.report_date, sj.job_status, sj.created_at, sj.updated_at
	FROM sp2dk_jobs AS sj
	JOIN clients AS c ON sj.client_id = c.client_id
	LEFT JOIN staffs AS s ON sj.assigned_pic_staff_sigma_id = s.staff_id
	WHERE sj.client_id = $1` // Filter berdasarkan client_id

	args := []interface{}{clientID} // Argumen pertama adalah clientID
	paramCounter := 2 // Argumen berikutnya dimulai dari $2

	if !isAdmin && staffIDFilter != "" {
		query += fmt.Sprintf(" AND sj.assigned_pic_staff_sigma_id = $%d", paramCounter)
		args = append(args, staffIDFilter)
		paramCounter++
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
		var cID sql.NullString // Ganti clientID dengan cID karena clientID sudah jadi parameter
		var clientName sql.NullString
		var npwpClient sql.NullString
		var assignedPicStaffSigmaID sql.NullString
		var assignedPicStaffSigmaName sql.NullString
		var contractNo sql.NullString
		var contractDate sql.NullTime
		var sp2dkNo sql.NullString
		var sp2dkDate sql.NullTime
		var bap2dkNo sql.NullString
		var bap2dkDate sql.NullTime
		var paymentDate sql.NullTime
		var reportDate sql.NullTime

		err := rows.Scan(
			&job.JobID, &cID, &clientName, &npwpClient,
			&assignedPicStaffSigmaID, &assignedPicStaffSigmaName,
			&contractNo, &contractDate, &sp2dkNo, &sp2dkDate, &bap2dkNo, &bap2dkDate,
			&paymentDate, &reportDate, &job.JobStatus, &job.CreatedAt, &job.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan SP2DK job row by client ID: %w", err)
		}

		// Assign nullable fields
		job.ClientID = cID.String // Gunakan cID di sini
		if clientName.Valid { job.ClientName = clientName.String } else { job.ClientName = "" }
		if npwpClient.Valid { job.NpwpClient = npwpClient.String } else { job.NpwpClient = "" }
		if assignedPicStaffSigmaID.Valid { job.AssignedPicStaffSigmaID = assignedPicStaffSigmaID.String } else { job.AssignedPicStaffSigmaID = "" }
		if assignedPicStaffSigmaName.Valid { job.AssignedPicStaffSigmaName = assignedPicStaffSigmaName.String } else { job.AssignedPicStaffSigmaName = "" }
		if contractNo.Valid { job.ContractNo = contractNo.String } else { job.ContractNo = "" }
		if contractDate.Valid { job.ContractDate = &contractDate.Time } else { job.ContractDate = nil }
		if sp2dkNo.Valid { job.Sp2dkNo = sp2dkNo.String } else { job.Sp2dkNo = "" }
		if sp2dkDate.Valid { job.Sp2dkDate = &sp2dkDate.Time } else { job.Sp2dkDate = nil }
		if bap2dkNo.Valid { job.Bap2dkNo = bap2dkNo.String } else { job.Bap2dkNo = "" }
		if bap2dkDate.Valid { job.Bap2dkDate = &bap2dkDate.Time } else { job.Bap2dkDate = nil }
		if paymentDate.Valid { job.PaymentDate = &paymentDate.Time } else { job.PaymentDate = nil }
		if reportDate.Valid { job.ReportDate = &reportDate.Time } else { job.ReportDate = nil }

		sp2dkJobs = append(sp2dkJobs, job)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration for SP2DK jobs by client ID: %w", err)
	}

	return sp2dkJobs, nil
}

// GetAllSp2dkJobs fetches all SP2DK jobs with their associated client and staff info.
func (r *sp2dkJobRepository) GetAllSp2dkJobs(staffIDFilter string, isAdmin bool) ([]models.Sp2dkJob, error) {
	query := `
	SELECT
		sj.job_id, sj.client_id, c.client_name, c.npwp_client,
		sj.assigned_pic_staff_sigma_id, s.nama AS assigned_pic_staff_sigma_name,
		sj.contract_no, sj.contract_date, sj.sp2dk_no, sj.sp2dk_date, sj.bap2dk_no, sj.bap2dk_date,
		sj.payment_date, sj.report_date, sj.job_status, sj.proof_of_work_url, sj.created_at, sj.updated_at
	FROM sp2dk_jobs AS sj
	JOIN clients AS c ON sj.client_id = c.client_id
	LEFT JOIN staffs AS s ON sj.assigned_pic_staff_sigma_id = s.staff_id`

	args := []interface{}{}
	paramCounter := 1

	if !isAdmin && staffIDFilter != "" {
		query += fmt.Sprintf(" WHERE sj.assigned_pic_staff_sigma_id = $%d", paramCounter)
		args = append(args, staffIDFilter)
		paramCounter++
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
		var assignedPicStaffSigmaID sql.NullString
		var assignedPicStaffSigmaName sql.NullString
		var contractNo sql.NullString
		var contractDate sql.NullTime
		var sp2dkNo sql.NullString
		var sp2dkDate sql.NullTime
		var bap2dkNo sql.NullString
		var bap2dkDate sql.NullTime
		var paymentDate sql.NullTime
		var reportDate sql.NullTime
		var proofOfWorkURL sql.NullString

		err := rows.Scan(
			&job.JobID, &job.ClientID, &job.ClientName, &job.NpwpClient,
			&assignedPicStaffSigmaID, &assignedPicStaffSigmaName,
			&contractNo, &contractDate, &sp2dkNo, &sp2dkDate, &bap2dkNo, &bap2dkDate,
			&paymentDate, &reportDate, &job.JobStatus, &proofOfWorkURL, &job.CreatedAt, &job.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan SP2DK job row: %w", err)
		}

		// Assign nullable fields
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
		if contractNo.Valid {
			job.ContractNo = contractNo.String
		} else {
			job.ContractNo = ""
		}
		if contractDate.Valid {
			job.ContractDate = &contractDate.Time
		} else {
			job.ContractDate = nil
		}
		if sp2dkNo.Valid {
			job.Sp2dkNo = sp2dkNo.String
		} else {
			job.Sp2dkNo = ""
		}
		if sp2dkDate.Valid {
			job.Sp2dkDate = &sp2dkDate.Time
		} else {
			job.Sp2dkDate = nil
		}
		if bap2dkNo.Valid {
			job.Bap2dkNo = bap2dkNo.String
		} else {
			job.Bap2dkNo = ""
		}
		if bap2dkDate.Valid {
			job.Bap2dkDate = &bap2dkDate.Time
		} else {
			job.Bap2dkDate = nil
		}
		if paymentDate.Valid {
			job.PaymentDate = &paymentDate.Time
		} else {
			job.PaymentDate = nil
		}
		if reportDate.Valid {
			job.ReportDate = &reportDate.Time
		} else {
			job.ReportDate = nil
		}
		if proofOfWorkURL.Valid {
			job.ProofOfWorkURL = &proofOfWorkURL.String
		} else {
			job.ProofOfWorkURL = nil // Penting: set nil jika NULL di DB
		}

		sp2dkJobs = append(sp2dkJobs, job)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return sp2dkJobs, nil
}

// GetSp2dkJobByID fetches a single SP2DK job by its ID with associated client and staff info.
func (r *sp2dkJobRepository) GetSp2dkJobByID(id string, staffIDFilter string, isAdmin bool) (*models.Sp2dkJob, error) {
	query := `
	SELECT
		sj.job_id, sj.client_id, c.client_name, c.npwp_client,
		sj.assigned_pic_staff_sigma_id, s.nama AS assigned_pic_staff_sigma_name,
		sj.contract_no, sj.contract_date, sj.sp2dk_no, sj.sp2dk_date, sj.bap2dk_no, sj.bap2dk_date,
		sj.payment_date, sj.report_date, sj.job_status, sj.proof_of_work_url, sj.created_at, sj.updated_at
	FROM sp2dk_jobs AS sj
	JOIN clients AS c ON sj.client_id = c.client_id
	LEFT JOIN staffs AS s ON sj.assigned_pic_staff_sigma_id = s.staff_id
	WHERE sj.job_id = $1`

	args := []interface{}{id}
	paramCounter := 2

	if !isAdmin && staffIDFilter != "" {
		query += fmt.Sprintf(" AND sj.assigned_pic_staff_sigma_id = $%d", paramCounter)
		args = append(args, staffIDFilter)
	}

	var job models.Sp2dkJob
	var assignedPicStaffSigmaID sql.NullString
	var assignedPicStaffSigmaName sql.NullString
	var contractNo sql.NullString
	var contractDate sql.NullTime
	var sp2dkNo sql.NullString
	var sp2dkDate sql.NullTime
	var bap2dkNo sql.NullString
	var bap2dkDate sql.NullTime
	var paymentDate sql.NullTime
	var reportDate sql.NullTime
	var proofOfWorkURL sql.NullString

	err := r.db.QueryRow(query, args...).Scan(
		&job.JobID, &job.ClientID, &job.ClientName, &job.NpwpClient,
		&assignedPicStaffSigmaID, &assignedPicStaffSigmaName,
		&contractNo, &contractDate, &sp2dkNo, &sp2dkDate, &bap2dkNo, &bap2dkDate,
		&paymentDate, &reportDate, &proofOfWorkURL, &job.JobStatus, &job.CreatedAt, &job.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get SP2DK job by ID: %w", err)
	}

	// Assign nullable fields
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
	if contractNo.Valid {
		job.ContractNo = contractNo.String
	} else {
		job.ContractNo = ""
	}
	if contractDate.Valid {
		job.ContractDate = &contractDate.Time
	} else {
		job.ContractDate = nil
	}
	if sp2dkNo.Valid {
		job.Sp2dkNo = sp2dkNo.String
	} else {
		job.Sp2dkNo = ""
	}
	if sp2dkDate.Valid {
		job.Sp2dkDate = &sp2dkDate.Time
	} else {
		job.Sp2dkDate = nil
	}
	if bap2dkNo.Valid {
		job.Bap2dkNo = bap2dkNo.String
	} else {
		job.Bap2dkNo = ""
	}
	if bap2dkDate.Valid {
		job.Bap2dkDate = &bap2dkDate.Time
	} else {
		job.Bap2dkDate = nil
	}
	if paymentDate.Valid {
		job.PaymentDate = &paymentDate.Time
	} else {
		job.PaymentDate = nil
	}
	if reportDate.Valid {
		job.ReportDate = &reportDate.Time
	} else {
		job.ReportDate = nil
	}
	if proofOfWorkURL.Valid {
		job.ProofOfWorkURL = &proofOfWorkURL.String
	} else {
		job.ProofOfWorkURL = nil // Penting: set nil jika NULL di DB
	}

	return &job, nil
}

// UpdateSp2dkJob updates an existing SP2DK job.
func (r *sp2dkJobRepository) UpdateSp2dkJob(job *models.Sp2dkJob) error {
	// Handle nullable fields for UUIDs
	var assignedPicStaffSigmaID sql.NullString
	if job.AssignedPicStaffSigmaID != "" {
		assignedPicStaffSigmaID = sql.NullString{String: job.AssignedPicStaffSigmaID, Valid: true}
	} else {
		assignedPicStaffSigmaID = sql.NullString{Valid: false}
	}

	var proofOfWorkURL sql.NullString
	if job.ProofOfWorkURL != nil && *job.ProofOfWorkURL != "" { // Jika ada URL dan tidak kosong
		proofOfWorkURL = sql.NullString{String: *job.ProofOfWorkURL, Valid: true}
	} else {
		proofOfWorkURL = sql.NullString{Valid: false} // Akan menjadi NULL di DB jika nil atau string kosong
	}

	// Handle nullable date/string fields
	var contractDate sql.NullTime
	if job.ContractDate != nil {
		contractDate = sql.NullTime{Time: *job.ContractDate, Valid: true}
	} else {
		contractDate = sql.NullTime{Valid: false}
	}

	var sp2dkDate sql.NullTime
	if job.Sp2dkDate != nil {
		sp2dkDate = sql.NullTime{Time: *job.Sp2dkDate, Valid: true}
	} else {
		sp2dkDate = sql.NullTime{Valid: false}
	}

	var bap2dkDate sql.NullTime
	if job.Bap2dkDate != nil {
		bap2dkDate = sql.NullTime{Time: *job.Bap2dkDate, Valid: true}
	} else {
		bap2dkDate = sql.NullTime{Valid: false}
	}

	var paymentDate sql.NullTime
	if job.PaymentDate != nil {
		paymentDate = sql.NullTime{Time: *job.PaymentDate, Valid: true}
	} else {
		paymentDate = sql.NullTime{Valid: false}
	}

	var reportDate sql.NullTime
	if job.ReportDate != nil {
		reportDate = sql.NullTime{Time: *job.ReportDate, Valid: true}
	} else {
		reportDate = sql.NullTime{Valid: false}
	}

	var contractNo sql.NullString
	if job.ContractNo != "" {
		contractNo = sql.NullString{String: job.ContractNo, Valid: true}
	} else {
		contractNo = sql.NullString{Valid: false}
	}

	var sp2dkNo sql.NullString
	if job.Sp2dkNo != "" {
		sp2dkNo = sql.NullString{String: job.Sp2dkNo, Valid: true}
	} else {
		sp2dkNo = sql.NullString{Valid: false}
	}

	var bap2dkNo sql.NullString
	if job.Bap2dkNo != "" {
		bap2dkNo = sql.NullString{String: job.Bap2dkNo, Valid: true}
	} else {
		bap2dkNo = sql.NullString{Valid: false}
	}


	query := `UPDATE sp2dk_jobs SET
		client_id = $1, assigned_pic_staff_sigma_id = $2, contract_no = $3, contract_date = $4,
		sp2dk_no = $5, sp2dk_date = $6, bap2dk_no = $7, bap2dk_date = $8,
		payment_date = $9, report_date = $10, job_status = $11, proof_of_work_url = $12, updated_at = $12
	WHERE job_id = $13`

	job.UpdatedAt = time.Now()

	_, err := r.db.Exec(query,
		job.ClientID, assignedPicStaffSigmaID, contractNo, contractDate,
		sp2dkNo, sp2dkDate, bap2dkNo, bap2dkDate,
		paymentDate, reportDate, job.JobStatus, proofOfWorkURL, job.UpdatedAt, job.JobID,
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
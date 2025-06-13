package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/iqsanfm/dashboard-pekerjaan-backend/internal/models" // Pastikan ini adalah modul Anda
)

// PemeriksaanJobRepository defines the interface for Pemeriksaan job data operations
type PemeriksaanJobRepository interface {
	CreatePemeriksaanJob(job *models.PemeriksaanJob) error
	GetAllPemeriksaanJobs(staffIDFilter string, isAdmin bool) ([]models.PemeriksaanJob, error)
	GetPemeriksaanJobByID(id string, staffIDFilter string, isAdmin bool) (*models.PemeriksaanJob, error)
	UpdatePemeriksaanJob(job *models.PemeriksaanJob) error
	DeletePemeriksaanJob(id string) error
	GetPemeriksaanJobsByClientID(clientID string, staffIDFilter string, isAdmin bool) ([]models.PemeriksaanJob, error) // <-- Pastikan ini ada di interface
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

	var sp2Date sql.NullTime
	if job.Sp2Date != nil {
		sp2Date = sql.NullTime{Time: *job.Sp2Date, Valid: true}
	} else {
		sp2Date = sql.NullTime{Valid: false}
	}

	var skpDate sql.NullTime
	if job.SkpDate != nil {
		skpDate = sql.NullTime{Time: *job.SkpDate, Valid: true}
	} else {
		skpDate = sql.NullTime{Valid: false}
	}

	var contractNo sql.NullString
	if job.ContractNo != "" {
		contractNo = sql.NullString{String: job.ContractNo, Valid: true}
	} else {
		contractNo = sql.NullString{Valid: false}
	}

	var sp2No sql.NullString
	if job.Sp2No != "" {
		sp2No = sql.NullString{String: job.Sp2No, Valid: true}
	} else {
		sp2No = sql.NullString{Valid: false}
	}

	var skpNo sql.NullString
	if job.SkpNo != "" {
		skpNo = sql.NullString{String: job.SkpNo, Valid: true}
	} else {
		skpNo = sql.NullString{Valid: false}
	}


	query := `INSERT INTO pemeriksaan_jobs (
		client_id, assigned_pic_staff_sigma_id, contract_no, contract_date, sp2_no, sp2_date,
		skp_no, skp_date, job_status, created_at, updated_at
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
	) RETURNING job_id, created_at, updated_at`

	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now()
	}
	job.UpdatedAt = time.Now()

	err := r.db.QueryRow(query,
		job.ClientID, assignedPicStaffSigmaID, contractNo, contractDate, sp2No, sp2Date,
		skpNo, skpDate, job.JobStatus,
		job.CreatedAt, job.UpdatedAt,
	).Scan(&job.JobID, &job.CreatedAt, &job.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create Pemeriksaan job: %w", err)
	}
	return nil
}

// GetAllPemeriksaanJobs fetches all Pemeriksaan jobs with their associated client and staff info.
func (r *pemeriksaanJobRepository) GetAllPemeriksaanJobs(staffIDFilter string, isAdmin bool) ([]models.PemeriksaanJob, error) {
	query := `
	SELECT
		pj.job_id, pj.client_id, c.client_name, c.npwp_client,
		pj.assigned_pic_staff_sigma_id, s.nama AS assigned_pic_staff_sigma_name,
		pj.contract_no, pj.contract_date, pj.sp2_no, pj.sp2_date, pj.skp_no, pj.skp_date,
		pj.job_status, pj.proof_of_work_url, pj.created_at, pj.updated_at
	FROM pemeriksaan_jobs AS pj
	JOIN clients AS c ON pj.client_id = c.client_id
	LEFT JOIN staffs AS s ON pj.assigned_pic_staff_sigma_id = s.staff_id`

	args := []interface{}{}
	paramCounter := 1

	if !isAdmin && staffIDFilter != "" {
		query += fmt.Sprintf(" WHERE pj.assigned_pic_staff_sigma_id = $%d", paramCounter)
		args = append(args, staffIDFilter)
		paramCounter++
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
		var assignedPicStaffSigmaID sql.NullString
		var assignedPicStaffSigmaName sql.NullString
		var contractNo sql.NullString
		var contractDate sql.NullTime
		var sp2No sql.NullString
		var sp2Date sql.NullTime
		var skpNo sql.NullString
		var skpDate sql.NullTime
		var proofOfWorkURL sql.NullString

		err := rows.Scan(
			&job.JobID, &job.ClientID, &job.ClientName, &job.NpwpClient,
			&assignedPicStaffSigmaID, &assignedPicStaffSigmaName,
			&contractNo, &proofOfWorkURL, &contractDate, &sp2No, &sp2Date, &skpNo, &skpDate,
			&job.JobStatus, &job.CreatedAt, &job.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan Pemeriksaan job row: %w", err)
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
		if sp2No.Valid {
			job.Sp2No = sp2No.String
		} else {
			job.Sp2No = ""
		}
		if sp2Date.Valid {
			job.Sp2Date = &sp2Date.Time
		} else {
			job.Sp2Date = nil
		}
		if skpNo.Valid {
			job.SkpNo = skpNo.String
		} else {
			job.SkpNo = ""
		}
		if skpDate.Valid {
			job.SkpDate = &skpDate.Time
		} else {
			job.SkpDate = nil
		}
		if proofOfWorkURL.Valid {
			job.ProofOfWorkURL = &proofOfWorkURL.String
		} else {
			job.ProofOfWorkURL = nil // Penting: set nil jika NULL di DB
		}

		pemeriksaanJobs = append(pemeriksaanJobs, job)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration for Pemeriksaan jobs: %w", err)
	}

	return pemeriksaanJobs, nil
}

// GetPemeriksaanJobByID fetches a single Pemeriksaan job by its ID with associated client and staff info.
func (r *pemeriksaanJobRepository) GetPemeriksaanJobByID(id string, staffIDFilter string, isAdmin bool) (*models.PemeriksaanJob, error) {
	query := `
	SELECT
		pj.job_id, pj.client_id, c.client_name, c.npwp_client,
		pj.assigned_pic_staff_sigma_id, s.nama AS assigned_pic_staff_sigma_name,
		pj.contract_no, pj.contract_date, pj.sp2_no, pj.sp2_date, pj.skp_no, pj.skp_date,
		pj.job_status, pj.proof_of_work_url, pj.created_at, pj.updated_at
	FROM pemeriksaan_jobs AS pj
	JOIN clients AS c ON pj.client_id = c.client_id
	LEFT JOIN staffs AS s ON pj.assigned_pic_staff_sigma_id = s.staff_id
	WHERE pj.job_id = $1`

	args := []interface{}{id}
	paramCounter := 2

	if !isAdmin && staffIDFilter != "" {
		query += fmt.Sprintf(" AND pj.assigned_pic_staff_sigma_id = $%d", paramCounter)
		args = append(args, staffIDFilter)
	}

	var job models.PemeriksaanJob
	var assignedPicStaffSigmaID sql.NullString
	var assignedPicStaffSigmaName sql.NullString
	var contractNo sql.NullString
	var contractDate sql.NullTime
	var sp2No sql.NullString
	var sp2Date sql.NullTime
	var skpNo sql.NullString
	var skpDate sql.NullTime
	var proofOfWorkURL sql.NullString

	err := r.db.QueryRow(query, args...).Scan(
		&job.JobID, &job.ClientID, &job.ClientName, &job.NpwpClient,
		&assignedPicStaffSigmaID, &assignedPicStaffSigmaName,
		&contractNo, &contractDate, &sp2No, &sp2Date, &skpNo, &skpDate,
		&job.JobStatus, &proofOfWorkURL, &job.CreatedAt, &job.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get Pemeriksaan job by ID: %w", err)
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
	if sp2No.Valid {
		job.Sp2No = sp2No.String
	} else {
		job.Sp2No = ""
	}
	if sp2Date.Valid {
		job.Sp2Date = &sp2Date.Time
	} else {
		job.Sp2Date = nil
	}
	if skpNo.Valid {
		job.SkpNo = skpNo.String
	} else {
		job.SkpNo = ""
	}
	if skpDate.Valid {
		job.SkpDate = &skpDate.Time
	} else {
		job.SkpDate = nil
	}
	if proofOfWorkURL.Valid {
		job.ProofOfWorkURL = &proofOfWorkURL.String
	} else {
		job.ProofOfWorkURL = nil // Penting: set nil jika NULL di DB
	}

	return &job, nil
}

// UpdatePemeriksaanJob updates an existing Pemeriksaan job.
func (r *pemeriksaanJobRepository) UpdatePemeriksaanJob(job *models.PemeriksaanJob) error {
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

	var sp2Date sql.NullTime
	if job.Sp2Date != nil {
		sp2Date = sql.NullTime{Time: *job.Sp2Date, Valid: true}
	} else {
		sp2Date = sql.NullTime{Valid: false}
	}

	var skpDate sql.NullTime
	if job.SkpDate != nil {
		skpDate = sql.NullTime{Time: *job.SkpDate, Valid: true}
	} else {
		skpDate = sql.NullTime{Valid: false}
	}

	var contractNo sql.NullString
	if job.ContractNo != "" {
		contractNo = sql.NullString{String: job.ContractNo, Valid: true}
	} else {
		contractNo = sql.NullString{Valid: false}
	}

	var sp2No sql.NullString
	if job.Sp2No != "" {
		sp2No = sql.NullString{String: job.Sp2No, Valid: true}
	} else {
		sp2No = sql.NullString{Valid: false}
	}

	var skpNo sql.NullString
	if job.SkpNo != "" {
		skpNo = sql.NullString{String: job.SkpNo, Valid: true}
	} else {
		skpNo = sql.NullString{Valid: false}
	}

	var proofOfWorkURL sql.NullString
	if job.ProofOfWorkURL != nil && *job.ProofOfWorkURL != "" { // Jika ada URL dan tidak kosong
		proofOfWorkURL = sql.NullString{String: *job.ProofOfWorkURL, Valid: true}
	} else {
		proofOfWorkURL = sql.NullString{Valid: false} // Akan menjadi NULL di DB jika nil atau string kosong
	}


	query := `UPDATE pemeriksaan_jobs SET
		client_id = $1, assigned_pic_staff_sigma_id = $2, contract_no = $3, contract_date = $4,
		sp2_no = $5, sp2_date = $6, skp_no = $7, skp_date = $8,
		job_status = $9, proof_of_work_url = $10, updated_at = $11
	WHERE job_id = $12`

	job.UpdatedAt = time.Now()

	_, err := r.db.Exec(query,
		job.ClientID, assignedPicStaffSigmaID, contractNo, contractDate,
		sp2No, sp2Date, skpNo, skpDate,
		job.JobStatus, proofOfWorkURL, job.UpdatedAt, job.JobID,
	)

	if err != nil {
		return fmt.Errorf("failed to update Pemeriksaan job: %w", err)
	}
	return nil
}

// DeletePemeriksaanJob deletes an Pemeriksaan job by its ID.
func (r *pemeriksaanJobRepository) DeletePemeriksaanJob(id string) error {
	query := `DELETE FROM pemeriksaan_jobs WHERE job_id = $1`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete Pemeriksaan job: %w", err)
	}
	return nil
}

// GetPemeriksaanJobsByClientID fetches Pemeriksaan jobs for a specific client, optionally filtered by staffID
func (r *pemeriksaanJobRepository) GetPemeriksaanJobsByClientID(clientID string, staffIDFilter string, isAdmin bool) ([]models.PemeriksaanJob, error) { // <-- Ini fungsi yang dimaksud
	query := `
	SELECT
		pj.job_id, pj.client_id, c.client_name, c.npwp_client,
		pj.assigned_pic_staff_sigma_id, s.nama AS assigned_pic_staff_sigma_name,
		pj.contract_no, pj.contract_date, pj.sp2_no, pj.sp2_date, pj.skp_no, pj.skp_date,
		pj.job_status, pj.created_at, pj.updated_at
	FROM pemeriksaan_jobs AS pj
	JOIN clients AS c ON pj.client_id = c.client_id
	LEFT JOIN staffs AS s ON pj.assigned_pic_staff_sigma_id = s.staff_id
	WHERE pj.client_id = $1` // Filter berdasarkan client_id

	args := []interface{}{clientID} // Argumen pertama adalah clientID
	paramCounter := 2             // Argumen berikutnya dimulai dari $2

	if !isAdmin && staffIDFilter != "" { // Jika bukan admin DAN ada staffID yang difilter
		query += fmt.Sprintf(" AND pj.assigned_pic_staff_sigma_id = $%d", paramCounter)
		args = append(args, staffIDFilter)
		// paramCounter tidak perlu diincrement karena ini query single-value (jika di-filter)
	}

	query += " ORDER BY pj.created_at DESC;"

	rows, err := r.db.Query(query, args...) // <-- Menggunakan db.Query() untuk banyak baris
	if err != nil {
		// Jika ada error saat menjalankan query (misal sintaks SQL salah, koneksi DB putus)
		return nil, fmt.Errorf("failed to get Pemeriksaan jobs by client ID: %w", err)
	}
	defer rows.Close() // Pastikan rows ditutup

	var pemeriksaanJobs []models.PemeriksaanJob // Slice untuk menampung hasil
	for rows.Next() {                          // Loop selama ada baris berikutnya
		var job models.PemeriksaanJob // Deklarasi struct untuk setiap baris
		
		// Variabel untuk menampung nilai nullable dari DB
		var assignedPicStaffSigmaID sql.NullString
		var assignedPicStaffSigmaName sql.NullString
		var contractNo sql.NullString
		var contractDate sql.NullTime
		var sp2No sql.NullString
		var sp2Date sql.NullTime
		var skpNo sql.NullString
		var skpDate sql.NullTime

		err := rows.Scan( // Scan nilai dari baris saat ini ke variabel
			&job.JobID, &job.ClientID, &job.ClientName, &job.NpwpClient,
			&assignedPicStaffSigmaID, &assignedPicStaffSigmaName,
			&contractNo, &contractDate, &sp2No, &sp2Date, &skpNo, &skpDate,
			&job.JobStatus, &job.CreatedAt, &job.UpdatedAt,
		)
		if err != nil {
			// Jika ada error saat scan satu baris (misal tipe data tidak cocok)
			return nil, fmt.Errorf("failed to scan Pemeriksaan job row by client ID: %w", err)
		}

		// Assign nilai dari sql.Null* ke struct Go, menangani NULL
		if assignedPicStaffSigmaID.Valid {
			job.AssignedPicStaffSigmaID = assignedPicStaffSigmaID.String
		} else {
			job.AssignedPicStaffSigmaID = "" // <-- Pastikan ini job.AssignedPicStaffSigmaID
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
		if sp2No.Valid {
			job.Sp2No = sp2No.String
		} else {
			job.Sp2No = ""
		}
		if sp2Date.Valid {
			job.Sp2Date = &sp2Date.Time
		} else {
			job.Sp2Date = nil
		}
		if skpNo.Valid {
			job.SkpNo = skpNo.String
		} else {
			job.SkpNo = ""
		}
		if skpDate.Valid {
			job.SkpDate = &skpDate.Time
		} else {
			job.SkpDate = nil
		}

		pemeriksaanJobs = append(pemeriksaanJobs, job) // Tambahkan job yang sudah diisi ke slice
	}

	// Setelah loop, cek apakah ada error saat iterasi (misal koneksi terputus di tengah jalan)
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration for Pemeriksaan jobs by client ID: %w", err)
	}

	// Jika tidak ada baris yang ditemukan, slice 'pemeriksaanJobs' akan kosong. Ini bukan error.
	return pemeriksaanJobs, nil // Kembalikan slice (bisa kosong) dan error nil
}
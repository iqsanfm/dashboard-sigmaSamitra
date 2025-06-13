package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/iqsanfm/dashboard-pekerjaan-backend/internal/models"
	"github.com/iqsanfm/dashboard-pekerjaan-backend/pkg/utils" // For password hashing and NIP generation
)

// StaffRepository defines the interface for staff data operations
type StaffRepository interface {
	CreateStaff(staff *models.Staff) error
	GetAllStaffs() ([]models.Staff, error)
	GetStaffByID(id string) (*models.Staff, error)
	GetStaffByEmail(email string) (*models.Staff, error) // Useful for authentication
	UpdateStaff(staff *models.Staff) error
	DeleteStaff(id string) error
}

// staffRepository implements StaffRepository interface
type staffRepository struct {
	db *sql.DB
}

// NewStaffRepository creates a new StaffRepository
func NewStaffRepository(db *sql.DB) StaffRepository {
	return &staffRepository{db: db}
}

// CreateStaff inserts a new staff member into the database.
// It generates NIP and hashes the password before saving.
func (r *staffRepository) CreateStaff(staff *models.Staff) error {
	// Generate NIP
	staff.NIP = utils.GenerateNIP(staff.Nama)

	// Hash the password
	hashedPassword, err := utils.HashPassword(staff.PasswordHashed) // Assuming PasswordHashed temporarily holds plain password
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	staff.PasswordHashed = hashedPassword

	query := `INSERT INTO staffs (
		nip, nama, email, password_hashed, role, created_at, updated_at
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7
	) RETURNING staff_id, created_at, updated_at`

	if staff.CreatedAt.IsZero() {
		staff.CreatedAt = time.Now()
	}
	staff.UpdatedAt = time.Now()

	err = r.db.QueryRow(query,
		staff.NIP, staff.Nama, staff.Email, staff.PasswordHashed, staff.Role,
		staff.CreatedAt, staff.UpdatedAt,
	).Scan(&staff.StaffID, &staff.CreatedAt, &staff.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create staff: %w", err)
	}
	return nil
}

// GetAllStaffs fetches all staff members from the database
func (r *staffRepository) GetAllStaffs() ([]models.Staff, error) {
	query := `SELECT
		staff_id, nip, nama, email, password_hashed, role, created_at, updated_at
	FROM staffs ORDER BY nama ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all staffs: %w", err)
	}
	defer rows.Close()

	var staffs []models.Staff
	for rows.Next() {
		var staff models.Staff
		err := rows.Scan(
			&staff.StaffID, &staff.NIP, &staff.Nama, &staff.Email,
			&staff.PasswordHashed, &staff.Role, &staff.CreatedAt, &staff.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan staff row: %w", err)
		}
		staffs = append(staffs, staff)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return staffs, nil
}

// GetStaffByID fetches a staff member by their ID from the database
func (r *staffRepository) GetStaffByID(id string) (*models.Staff, error) {
	query := `SELECT
		staff_id, nip, nama, email, password_hashed, role, created_at, updated_at
	FROM staffs WHERE staff_id = $1`

	var staff models.Staff
	err := r.db.QueryRow(query, id).Scan(
		&staff.StaffID, &staff.NIP, &staff.Nama, &staff.Email,
		&staff.PasswordHashed, &staff.Role, &staff.CreatedAt, &staff.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Staff not found
		}
		return nil, fmt.Errorf("failed to get staff by ID: %w", err)
	}
	return &staff, nil
}

// GetStaffByEmail fetches a staff member by their email from the database
func (r *staffRepository) GetStaffByEmail(email string) (*models.Staff, error) {
	query := `SELECT
		staff_id, nip, nama, email, password_hashed, role, created_at, updated_at
	FROM staffs WHERE email = $1`

	var staff models.Staff
	err := r.db.QueryRow(query, email).Scan(
		&staff.StaffID, &staff.NIP, &staff.Nama, &staff.Email,
		&staff.PasswordHashed, &staff.Role, &staff.CreatedAt, &staff.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Staff not found
		}
		return nil, fmt.Errorf("failed to get staff by email: %w", err)
	}
	return &staff, nil
}

// UpdateStaff updates an existing staff member in the database.
// It will hash the new password if provided.
func (r *staffRepository) UpdateStaff(staff *models.Staff) error {
	// If a new password (in hashed form) is provided in the staff struct, hash it again
	// This assumes the handler passed a plain text password to PasswordHashed for new hash
	// Or, if PasswordHashed already contains a new hash from client, it's just saved directly.
	// For PATCH, you'd fetch existing and only update if a new password string is given.

	// A more robust way for PATCH password:
	// If staff.PasswordHashed contains a NEW plain text password from request:
	if len(staff.PasswordHashed) > 0 && !utils.IsBcryptHash(staff.PasswordHashed) { // Add IsBcryptHash check
        hashedPassword, err := utils.HashPassword(staff.PasswordHashed)
        if err != nil {
            return fmt.Errorf("failed to hash new password: %w", err)
        }
        staff.PasswordHashed = hashedPassword
    }

	query := `UPDATE staffs SET
		nip = $1, nama = $2, email = $3, password_hashed = $4, role = $5, updated_at = $6
	WHERE staff_id = $7`

	staff.UpdatedAt = time.Now()

	_, err := r.db.Exec(query,
		staff.NIP, staff.Nama, staff.Email, staff.PasswordHashed, staff.Role,
		staff.UpdatedAt, staff.StaffID,
	)

	if err != nil {
		return fmt.Errorf("failed to update staff: %w", err)
	}
	return nil
}

// DeleteStaff deletes a staff member from the database
func (r *staffRepository) DeleteStaff(id string) error {
	query := `DELETE FROM staffs WHERE staff_id = $1`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete staff: %w", err)
	}
	return nil
}
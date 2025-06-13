package repositories

import (
	"database/sql"
	"fmt"
	"time" // Import time package

	"github.com/iqsanfm/dashboard-pekerjaan-backend/internal/models"
	"github.com/iqsanfm/dashboard-pekerjaan-backend/pkg/utils" // For password hashing
)

// ClientRepository defines the interface for client data operations
type ClientRepository interface {
	CreateClient(client *models.Client) error
	GetAllClients(staffIDFilter string, isAdmin bool) ([]models.Client, error)
	GetClientByID(id string, staffIDFilter string, isAdmin bool) (*models.Client, error) 
	UpdateClient(client *models.Client) error
	DeleteClient(id string) error
}

// clientRepository implements ClientRepository interface
type clientRepository struct {
	db *sql.DB
}

// NewClientRepository creates a new ClientRepository
func NewClientRepository(db *sql.DB) ClientRepository {
	return &clientRepository{db: db}
}

// CreateClient inserts a new client into the database
func (r *clientRepository) CreateClient(client *models.Client) error {
    // Hash the password before saving
    hashedPassword, err := utils.HashPassword(client.CoretaxPasswordHashed) // Use CoretaxPasswordHashed as the input for hashing
    if err != nil {
        return fmt.Errorf("failed to hash password: %w", err)
    }
    client.CoretaxPasswordHashed = hashedPassword // Update the struct field with the hashed password

	var picStaffSigmaID sql.NullString
	if client.PicStaffSigmaID != "" {
		picStaffSigmaID = sql.NullString{String: client.PicStaffSigmaID, Valid: true}
	} else {
		picStaffSigmaID = sql.NullString{Valid: false} // Akan menjadi NULL di database
	}

	query := `INSERT INTO clients (
		client_name, npwp_client, address_client, membership_status, phone_client, email_client,
		pic_client, djp_online_username, coretax_username, coretax_password_hashed, pic_staff_sigma_id,
		client_category, pph_final_umkm, pph_25, pph_21, pph_unifikasi, ppn, spt_tahunan,
		pelaporan_deviden, laporan_keuangan, investasi_deviden, created_at, updated_at
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23
	) RETURNING client_id, created_at, updated_at`

	// Set current time for CreatedAt and UpdatedAt if not already set
	if client.CreatedAt.IsZero() {
		client.CreatedAt = time.Now()
	}
	client.UpdatedAt = time.Now()

	err = r.db.QueryRow(query,
		client.ClientName, client.NpwpClient, client.AddressClient, client.MembershipStatus, client.PhoneClient,
		client.EmailClient, client.PicClient, client.DjpOnlineUsername, client.CoretaxUsername,
		client.CoretaxPasswordHashed, picStaffSigmaID, client.ClientCategory,
		client.PphFinalUmkm, client.Pph25, client.Pph21, client.PphUnifikasi, client.Ppn,
		client.SptTahunan, client.PelaporanDeviden, client.LaporanKeuangan, client.InvestasiDeviden,
		client.CreatedAt, client.UpdatedAt,
	).Scan(&client.ClientID, &client.CreatedAt, &client.UpdatedAt) // Scan the returned client_id and timestamps

	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	return nil
}

// GetAllClients fetches all clients from the database
func (r *clientRepository) GetAllClients(staffIDFilter string, isAdmin bool) ([]models.Client, error) {
	query := `SELECT
		c.client_id, c.client_name, c.npwp_client, c.address_client, c.membership_status, c.phone_client, c.email_client,
		c.pic_client, c.djp_online_username, c.coretax_username, c.coretax_password_hashed, c.pic_staff_sigma_id,
		s.nama AS pic_staff_sigma_name,
		c.client_category, c.pph_final_umkm, c.pph_25, c.pph_21, c.pph_unifikasi, c.ppn, c.spt_tahunan,
		c.pelaporan_deviden, c.laporan_keuangan, c.investasi_deviden, c.created_at, c.updated_at
	FROM clients AS c
	LEFT JOIN staffs AS s ON c.pic_staff_sigma_id = s.staff_id`

	args := []interface{}{}
	paramCounter := 1

	if !isAdmin && staffIDFilter != "" { // Jika bukan admin DAN ada staffID yang difilter
		query += fmt.Sprintf(" WHERE c.pic_staff_sigma_id = $%d", paramCounter)
		args = append(args, staffIDFilter)
		paramCounter++
	}

	query += " ORDER BY c.client_name ASC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get all clients: %w", err)
	}
	defer rows.Close()

	var clients []models.Client
	for rows.Next() {
		var client models.Client
		var picStaffSigmaID sql.NullString // Gunakan sql.NullString untuk ID staf yang bisa NULL
		var picStaffSigmaName sql.NullString

		// Deklarasi variabel sql.NullBool untuk kolom pajak
		var pphFinalUmkm sql.NullBool
		var pph25 sql.NullBool
		var pph21 sql.NullBool
		var pphUnifikasi sql.NullBool
		var ppn sql.NullBool
		var sptTahunan sql.NullBool
		var pelaporanDeviden sql.NullBool
		var laporanKeuangan sql.NullBool
		var investasiDeviden sql.NullBool

		// --- PERBAIKAN PENTING DI SINI:
		// HANYA gunakaan variabel sql.NullBool yang baru.
		// Jangan sertakan lagi &client.PphFinalUmkm, &client.Pph25, dst. di sini.
		// Total harus ada 25 variabel yang discan.
		// ---
		err := rows.Scan(
			&client.ClientID, &client.ClientName, &client.NpwpClient, &client.AddressClient, &client.MembershipStatus,
			&client.PhoneClient, &client.EmailClient, &client.PicClient, &client.DjpOnlineUsername,
			&client.CoretaxUsername, &client.CoretaxPasswordHashed,
			&picStaffSigmaID, &picStaffSigmaName,
			&client.ClientCategory,
			&pphFinalUmkm, &pph25, &pph21, &pphUnifikasi, &ppn, &sptTahunan,
			&pelaporanDeviden, &laporanKeuangan, &investasiDeviden,
			&client.CreatedAt, &client.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan client row: %w", err)
		}

		// Assignment dari sql.NullString ke string
		if picStaffSigmaID.Valid {
			client.PicStaffSigmaID = picStaffSigmaID.String
		} else {
			client.PicStaffSigmaID = ""
		}
		if picStaffSigmaName.Valid {
			client.PicStaffSigmaName = picStaffSigmaName.String
		} else {
			client.PicStaffSigmaName = ""
		}

		// Assignment dari sql.NullBool ke bool
		client.PphFinalUmkm = pphFinalUmkm.Bool
		client.Pph25 = pph25.Bool
		client.Pph21 = pph21.Bool
		client.PphUnifikasi = pphUnifikasi.Bool
		client.Ppn = ppn.Bool
		client.SptTahunan = sptTahunan.Bool
		client.PelaporanDeviden = pelaporanDeviden.Bool
		client.LaporanKeuangan = laporanKeuangan.Bool
		client.InvestasiDeviden = investasiDeviden.Bool

		clients = append(clients, client) // <-- Hanya ada SATU append di sini
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return clients, nil
}

// GetClientByID fetches a client by their ID from the database
func (r *clientRepository) GetClientByID(id string, staffIDFilter string, isAdmin bool) (*models.Client, error) {
	query := `SELECT
		c.client_id, c.client_name, c.npwp_client, c.address_client, c.membership_status, c.phone_client, c.email_client,
		c.pic_client, c.djp_online_username, c.coretax_username, c.coretax_password_hashed, c.pic_staff_sigma_id, -- UBAH INI
		s.nama AS pic_staff_sigma_name, -- TAMBAH INI
		c.client_category, c.pph_final_umkm, c.pph_25, c.pph_21, c.pph_unifikasi, c.ppn, c.spt_tahunan,
		c.pelaporan_deviden, c.laporan_keuangan, c.investasi_deviden, c.created_at, c.updated_at
	FROM clients AS c
	LEFT JOIN staffs AS s ON c.pic_staff_sigma_id = s.staff_id -- TAMBAH LEFT JOIN
	WHERE c.client_id = $1`

	args := []interface{}{id}
	paramCounter := 2 // Karena $1 sudah dipakai untuk client_id

	// --- LOGIKA FILTERING ---
	if !isAdmin && staffIDFilter != "" {
		query += fmt.Sprintf(" AND c.pic_staff_sigma_id = $%d", paramCounter)
		args = append(args, staffIDFilter)
		// paramCounter tidak perlu diincrement lagi karena ini query row tunggal
	}

	var client models.Client
	 var picStaffSigmaID sql.NullString
    var picStaffSigmaName sql.NullString
		var pphFinalUmkm sql.NullBool
	var pph25 sql.NullBool
	var pph21 sql.NullBool
	var pphUnifikasi sql.NullBool
	var ppn sql.NullBool
	var sptTahunan sql.NullBool
	var pelaporanDeviden sql.NullBool
	var laporanKeuangan sql.NullBool
	var investasiDeviden sql.NullBool

err := r.db.QueryRow(query, args...).Scan(
		&client.ClientID, &client.ClientName, &client.NpwpClient, &client.AddressClient, &client.MembershipStatus,
		&client.PhoneClient, &client.EmailClient, &client.PicClient, &client.DjpOnlineUsername,
		&client.CoretaxUsername, &client.CoretaxPasswordHashed,
		&picStaffSigmaID, &picStaffSigmaName,
		&client.ClientCategory,
		&pphFinalUmkm, &pph25, &pph21, &pphUnifikasi, &ppn, &sptTahunan,
		&pelaporanDeviden, &laporanKeuangan, &investasiDeviden,
		&client.CreatedAt, &client.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get client by ID: %w", err)
	}
	   if picStaffSigmaID.Valid {
        client.PicStaffSigmaID = picStaffSigmaID.String
    }
    if picStaffSigmaName.Valid {
        client.PicStaffSigmaName = picStaffSigmaName.String
    }
		  client.PphFinalUmkm = pphFinalUmkm.Bool
    client.Pph25 = pph25.Bool
    client.Pph21 = pph21.Bool
    client.PphUnifikasi = pphUnifikasi.Bool
    client.Ppn = ppn.Bool
    client.SptTahunan = sptTahunan.Bool
    client.PelaporanDeviden = pelaporanDeviden.Bool
    client.LaporanKeuangan = laporanKeuangan.Bool
    client.InvestasiDeviden = investasiDeviden.Bool
	return &client, nil
}

// UpdateClient updates an existing client in the database
func (r *clientRepository) UpdateClient(client *models.Client) error {
    // Note: We don't hash password here unless it's explicitly updated
    // If you allow updating password, you'd check if client.CoretaxPasswordHashed
    // contains a new value (e.g., from NewClientRequest.CoretaxPassword) and hash it.
    // For now, this assumes CoretaxPasswordHashed from the client struct is the final value.
	var picStaffSigmaID sql.NullString
	if client.PicStaffSigmaID != "" {
		picStaffSigmaID = sql.NullString{String: client.PicStaffSigmaID, Valid: true}
	} else {
		picStaffSigmaID = sql.NullString{Valid: false} // Akan menjadi NULL di database
	}

	query := `UPDATE clients SET
		client_name = $1, npwp_client = $2, address_client = $3, membership_status = $4,
		phone_client = $5, email_client = $6, pic_client = $7, djp_online_username = $8,
		coretax_username = $9, coretax_password_hashed = $10, pic_staff_sigma_id = $11, -- UBAH INI
		client_category = $12, pph_final_umkm = $13, pph_25 = $14, pph_21 = $15,
		pph_unifikasi = $16, ppn = $17, spt_tahunan = $18, pelaporan_deviden = $19,
		laporan_keuangan = $20, investasi_deviden = $21, updated_at = $22
	WHERE client_id = $23`

	client.UpdatedAt = time.Now() // Update the timestamp

	_, err := r.db.Exec(query,
		client.ClientName, client.NpwpClient, client.AddressClient, client.MembershipStatus,
		client.PhoneClient, client.EmailClient, client.PicClient, client.DjpOnlineUsername,
		client.CoretaxUsername, client.CoretaxPasswordHashed, picStaffSigmaID,
		client.ClientCategory, client.PphFinalUmkm, client.Pph25, client.Pph21,
		client.PphUnifikasi, client.Ppn, client.SptTahunan, client.PelaporanDeviden,
		client.LaporanKeuangan, client.InvestasiDeviden, client.UpdatedAt,
		client.ClientID,
	)

	if err != nil {
		return fmt.Errorf("failed to update client: %w", err)
	}
	return nil
}

// DeleteClient deletes a client from the database
func (r *clientRepository) DeleteClient(id string) error {
	query := `DELETE FROM clients WHERE client_id = $1`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete client: %w", err)
	}
	return nil
}
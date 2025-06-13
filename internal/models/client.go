package models

import (
	"time"
)

// Client represents a client entity in the database
type Client struct {
	ClientID             string    `json:"client_id"`
	ClientName           string    `json:"client_name"`
	NpwpClient           string    `json:"npwp_client"`
	AddressClient        string    `json:"address_client"`
	MembershipStatus     string    `json:"membership_status"`
	PhoneClient          string    `json:"phone_client"`
	EmailClient          string    `json:"email_client"`
	PicClient            string    `json:"pic_client"`
	DjpOnlineUsername    string    `json:"djp_online_username"`
	CoretaxUsername      string    `json:"coretax_username"`
	CoretaxPasswordHashed string    `json:"coretax_password_hashed"` // Hashed password
	
	PicStaffSigmaID      string    `json:"pic_staff_sigma_id"`    // ID PIC Staff Sigma (baru)
	PicStaffSigmaName    string    `json:"pic_staff_sigma_name"`  // Nama PIC Staff Sigma (dari JOIN)

	ClientCategory       string    `json:"client_category"` 
	
	// Layanan Pajak
	PphFinalUmkm         bool      `json:"pph_final_umkm"`
	Pph25                bool      `json:"pph_25"`
	Pph21                bool      `json:"pph_21"`
	PphUnifikasi         bool      `json:"pph_unifikasi"`
	Ppn                  bool      `json:"ppn"`
	SptTahunan           bool      `json:"spt_tahunan"`
	PelaporanDeviden     bool      `json:"pelaporan_deviden"`
	LaporanKeuangan      bool      `json:"laporan_keuangan"`
	InvestasiDeviden     bool      `json:"investasi_deviden"`

	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// NewClientRequest represents the structure for creating a new client (input from API)
type NewClientRequest struct {
	ClientName           string `json:"client_name" binding:"required"`
	NpwpClient           string `json:"npwp_client" binding:"required"`
	AddressClient        string `json:"address_client"`
	MembershipStatus     string `json:"membership_status"`
	PhoneClient          string `json:"phone_client"`
	EmailClient          string `json:"email_client"`
	PicClient            string `json:"pic_client"`
	DjpOnlineUsername    string `json:"djp_online_username"`
	CoretaxUsername      string `json:"coretax_username"`
	CoretaxPassword      string `json:"coretax_password"`
	PicStaffSigmaID      string `json:"pic_staff_sigma_id"` // ID PIC Staff Sigma

	ClientCategory       string `json:"client_category"`

	PphFinalUmkm         bool `json:"pph_final_umkm"`
	Pph25                bool `json:"pph_25"`
	Pph21                bool `json:"pph_21"`
	PphUnifikasi         bool `json:"pph_unifikasi"`
	Ppn                  bool `json:"ppn"`
	SptTahunan           bool `json:"spt_tahunan"`
	PelaporanDeviden     bool `json:"pelaporan_deviden"`
	LaporanKeuangan      bool `json:"laporan_keuangan"`
	InvestasiDeviden     bool `json:"investasi_deviden"`
}

// UpdateClientRequest (sesuaikan dengan perubahan)
type UpdateClientRequest struct {
	ClientName           *string `json:"client_name"`
	NpwpClient           *string `json:"npwp_client"`
	AddressClient        *string `json:"address_client"`
	MembershipStatus     *string `json:"membership_status"`
	PhoneClient          *string `json:"phone_client"`
	EmailClient          *string `json:"email_client" binding:"omitempty,email"`
	PicClient            *string `json:"pic_client"`
	DjpOnlineUsername    *string `json:"djp_online_username"`
	CoretaxUsername      *string `json:"coretax_username"`
	CoretaxPassword      *string `json:"coretax_password"`
	PicStaffSigmaID      *string `json:"pic_staff_sigma_id"` // ID PIC Staff Sigma

	ClientCategory       *string `json:"client_category"`

	PphFinalUmkm         *bool `json:"pph_final_umkm"`
	Pph25                *bool `json:"pph_25"`
	Pph21                *bool `json:"pph_21"`
	PphUnifikasi         *bool `json:"pph_unifikasi"`
	Ppn                  *bool `json:"ppn"`
	SptTahunan           *bool `json:"spt_tahunan"`
	PelaporanDeviden     *bool `json:"pelaporan_deviden"`
	LaporanKeuangan      *bool `json:"laporan_keuangan"`
	InvestasiDeviden     *bool `json:"investasi_deviden"`
}
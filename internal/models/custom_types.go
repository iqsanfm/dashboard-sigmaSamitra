package models

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

// CustomDate adalah tipe khusus untuk menangani format tanggal "YYYY-MM-DD"
// dalam komunikasi JSON dan interaksi Database.
type CustomDate struct {
	time.Time
}

// UnmarshalJSON mengajari Go cara membaca JSON string "YYYY-MM-DD" ke tipe CustomDate.
func (cd *CustomDate) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"") // Menghapus tanda kutip dari string JSON
	if s == "null" || s == "" {
		cd.Time = time.Time{} // Anggap sebagai waktu nol jika null atau kosong
		return
	}
	// "2006-01-02" adalah layout standar Go untuk format "YYYY-MM-DD"
	cd.Time, err = time.Parse("2006-01-02", s)
	if err != nil {
		return fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}
	return
}

// MarshalJSON mengajari Go cara menulis tipe CustomDate menjadi JSON string "YYYY-MM-DD".
func (cd CustomDate) MarshalJSON() ([]byte, error) {
	if cd.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", cd.Time.Format("2006-01-02"))), nil
}

// Value memberitahu driver database cara MENYIMPAN CustomDate ke database.
func (cd CustomDate) Value() (driver.Value, error) {
	// Jika waktunya nol, simpan sebagai NULL di database
	if cd.Time.IsZero() {
		return nil, nil
	}
	return cd.Time, nil // Simpan sebagai tipe time.Time standar
}

// Scan memberitahu CustomDate cara MEMBACA data dari database.
func (cd *CustomDate) Scan(value interface{}) error {
	if value == nil {
		cd.Time = time.Time{}
		return nil
	}
	t, ok := value.(time.Time)
	if !ok {
		return fmt.Errorf("failed to scan time.Time from database")
	}
	cd.Time = t
	return nil
}
// pkg/auth/claims.go
package auth

// Claims adalah struct kustom yang akan kita gunakan untuk data user dari JWT
type Claims struct {
	StaffID string `json:"staff_id"`
	IsAdmin bool   `json:"is_admin"`
	Role    string `json:"role"`
}
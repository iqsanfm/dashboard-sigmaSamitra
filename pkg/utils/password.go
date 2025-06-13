package utils

import (
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes the given password using bcrypt.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPasswordHash compares a plain password with its hashed version.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
func IsBcryptHash(s string) bool {
    // Bcrypt hashes typically start with $2a$, $2b$, or $2y$ followed by $ and two digits (cost)
    // and then 53 characters of salt and ciphertext. Total length ~60 characters.
	return len(s) == 60 && (strings.HasPrefix(s, "$2a$") || strings.HasPrefix(s, "$2b$") || strings.HasPrefix(s, "$2y$"))
}
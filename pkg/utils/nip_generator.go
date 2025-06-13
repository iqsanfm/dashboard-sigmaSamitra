package utils

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// GenerateNIP generates a NIP from the staff name.
// Example: "Iqsan Faisal" becomes "IF" + random number.
func GenerateNIP(nama string) string {
	// 1. Ambil inisial (dua huruf pertama dari nama)
	parts := strings.Split(nama, " ")
	inisial := ""
	for _, part := range parts {
		if len(part) > 0 {
			inisial += strings.ToUpper(part[:1])
		}
		if len(inisial) >= 2 { // Ambil maksimal 2 inisial
			break
		}
	}
	if len(inisial) < 2 {
		// Jika nama hanya satu kata, ambil dua huruf pertama
		if len(nama) >= 2 {
			inisial = strings.ToUpper(nama[:2])
		} else {
			inisial = strings.ToUpper(nama) // Jika sangat pendek, ambil semua
		}
	}

	// 2. Generate nomor random (6 digit)
	rand.Seed(time.Now().UnixNano()) // Penting untuk random yang benar
	randomNumber := rand.Intn(999999)
	return fmt.Sprintf("%s%06d", inisial, randomNumber)
}
// middlewares/auth_middleware.go
package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/iqsanfm/dashboard-pekerjaan-backend/config"
	"github.com/iqsanfm/dashboard-pekerjaan-backend/pkg/auth" // <-- BARIS BARU: Import package auth
)

type AuthMiddleware struct {
	JWTSecret []byte
}

func NewAuthMiddleware(cfg *config.Config) *AuthMiddleware {
	return &AuthMiddleware{
		JWTSecret: []byte(cfg.JWTSecretKey),
	}
}

func (m *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format (Bearer token expected)"})
			c.Abort()
			return
		}
		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return m.JWTSecret, nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
            // --- BLOK YANG DIPERBAIKI ---
			
			staffID, okID := claims["staff_id"].(string)
			role, okRole := claims["role"].(string)

			if !okID || !okRole {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
				c.Abort()
				return
			}
			
			// Buat struct Claims dari data token
			userClaims := &auth.Claims{
				StaffID: staffID,
				Role:    role,
				IsAdmin: role == "admin", // Konversi 'role' string menjadi boolean 'IsAdmin'
			}

			// Simpan SELURUH STRUCT ke dalam konteks dengan SATU KUNCI
			c.Set("user_claims", userClaims)

            // --- AKHIR BLOK YANG DIPERBAIKI ---

			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
		}
	}
}
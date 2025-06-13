package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"                          // Import library JWT
	"github.com/iqsanfm/dashboard-pekerjaan-backend/config" // Pastikan ini modul Anda
)

// AuthMiddleware struct to hold JWT secret
type AuthMiddleware struct {
	JWTSecret []byte
}

// NewAuthMiddleware creates a new AuthMiddleware instance
func NewAuthMiddleware(cfg *config.Config) *AuthMiddleware {
	return &AuthMiddleware{
		JWTSecret: []byte(cfg.JWTSecretKey),
	}
}

// AuthRequired is a Gin middleware to check JWT token
func (m *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			c.Abort()
			return
		}

		// Token format: "Bearer <token>"
		parts := strings.SplitN(tokenString, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format (Bearer token expected)"})
			c.Abort()
			return
		}
		tokenString = parts[1] // Get the actual token string

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the alg is what we expect:
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
			// Extract staff_id and role from claims
			staffID, ok := claims["staff_id"].(string)
			if !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims (staff_id missing or invalid)"})
				c.Abort()
				return
			}
			role, ok := claims["role"].(string)
			if !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims (role missing or invalid)"})
				c.Abort()
				return
			}

			// Store staff_id and role in Gin's context for downstream handlers
			c.Set("staffID", staffID)
			c.Set("role", role)

			c.Next() // Continue to the next handler
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
	}
}
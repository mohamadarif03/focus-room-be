package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mohamadarif03/focus-room-be/pkg/utils"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Error(c.Writer, nil, "Authorization header dibutuhkan", http.StatusUnauthorized)
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.Error(c.Writer, nil, "Format authorization header salah", http.StatusUnauthorized)
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			utils.Error(c.Writer, nil, "Token tidak valid atau expired", http.StatusUnauthorized)
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)

		c.Next()
	}
}

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleValue, exists := c.Get("role")
		if !exists {
			utils.Error(c.Writer, nil, "Gagal mengambil role dari context", http.StatusInternalServerError)
			c.Abort()
			return
		}

		role, ok := roleValue.(string)
		if !ok {
			utils.Error(c.Writer, nil, "Role di context formatnya salah", http.StatusInternalServerError)
			c.Abort()
			return
		}

		if role != "admin" {
			utils.Error(c.Writer, nil, "Akses ditolak. Hanya untuk admin.", http.StatusForbidden)
			c.Abort()
			return
		}

		c.Next()
	}
}

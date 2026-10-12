package middleware

import (
	"net/http"
	"strings"

	"taskflow/internal/auth"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return authMiddleware(jwtSecret, false)
}

// WebSocketAuthMiddleware permits a query token only for the
// WebSocket handshake endpoint.
func WebSocketAuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return authMiddleware(jwtSecret, true)
}

func authMiddleware(jwtSecret string, allowQueryToken bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenStr string

		header := c.GetHeader("Authorization")

		if header != "" {
			const prefix = "Bearer "
			if !strings.HasPrefix(header, prefix) {
				c.AbortWithStatusJSON(
					http.StatusUnauthorized,
					gin.H{"error": "invalid authorization header"},
				)
				return
			}
			tokenStr = strings.TrimSpace(
				strings.TrimPrefix(header, prefix),
			)
		}

		// Query tokens are forbidden on ordinary REST endpoints.
		if tokenStr == "" && allowQueryToken {
			tokenStr = c.Query("token")
		}

		if tokenStr == "" {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{"error": "missing token"},
			)
			return
		}

		claims, err := auth.ValidateToken(tokenStr, jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{"error": "invalid token"},
			)
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_name", claims.Name)
		c.Next()
	}
}

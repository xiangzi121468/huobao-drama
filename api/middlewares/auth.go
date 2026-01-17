package middlewares

import (
	"crypto/subtle"
	"strings"

	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware is a non-breaking auth guard:
// - If token is empty, all requests pass through.
// - If token is set, require Authorization: Bearer <token> or X-API-Key.
func AuthMiddleware(token string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if token == "" {
			c.Next()
			return
		}

		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		apiKey := strings.TrimSpace(c.GetHeader("X-API-Key"))
		provided := ""

		if apiKey != "" {
			provided = apiKey
		} else if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			provided = strings.TrimSpace(authHeader[7:])
		}

		if provided == "" || subtle.ConstantTimeCompare([]byte(provided), []byte(token)) != 1 {
			response.Unauthorized(c, "未授权")
			c.Abort()
			return
		}

		c.Next()
	}
}

package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func ExtractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)

		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			token := strings.TrimSpace(parts[1])
			if token != "" {
				return token
			}
		}
	}

	cookie, err := c.Cookie("access_token")
	if err == nil && cookie != "" {
		return cookie
	}

	if token := strings.TrimSpace(c.Query("token")); token != "" {
		return token
	}

	return ""
}

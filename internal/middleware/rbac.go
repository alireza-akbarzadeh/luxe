package middleware

import (
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/gin-gonic/gin"
)

func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Original role check
		role, ok := GetUserRole(c)
		if !ok {
			utils.UnauthorizedResponse(c, constants.ErrorUnauthorized)
			c.Abort()
			return
		}
		for _, allowed := range allowedRoles {
			if role == allowed {
				c.Next()
				return
			}
		}
		utils.ForbiddenResponse(c, constants.ErrorForbidden)
		c.Abort()
	}
}

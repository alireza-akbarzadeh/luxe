package middleware

import (
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/gin-gonic/gin"
)

// RequireRole ensures the authenticated user has one of the allowed roles.
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

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

// RequireAdmin restricts access to users with the admin role.
func RequireAdmin() gin.HandlerFunc {
	return RequireRole(constants.RoleAdmin)
}

// IsAdmin reports whether the current request context belongs to an admin user.
func IsAdmin(c *gin.Context) bool {
	role, ok := GetUserRole(c)
	return ok && role == constants.RoleAdmin
}

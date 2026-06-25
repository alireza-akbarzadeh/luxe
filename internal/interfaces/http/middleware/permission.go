package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
)

// RolePermissionChecker resolves whether a role slug grants a permission key.
type RolePermissionChecker interface {
	HasPermission(ctx context.Context, roleSlug, permissionKey string) (bool, error)
}

var rolePermissions RolePermissionChecker

// SetRolePermissionChecker wires the permission resolver (call once at startup).
func SetRolePermissionChecker(checker RolePermissionChecker) {
	rolePermissions = checker
}

// RequireStaff blocks storefront customer accounts from admin APIs.
func RequireStaff() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}
		role, ok := GetUserRole(c)
		if !ok {
			utils.UnauthorizedResponse(c, constants.ErrorUnauthorized)
			c.Abort()
			return
		}
		if role == constants.RoleUser {
			utils.ForbiddenResponse(c, "You do not have access to the admin dashboard")
			c.Abort()
			return
		}
		c.Next()
	}
}

// ModuleGuard enforces module.read (GET/HEAD) or module.write (mutations) for the user's role.
// Admin role bypasses all checks.
func ModuleGuard(module string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		role, ok := GetUserRole(c)
		if !ok {
			utils.UnauthorizedResponse(c, constants.ErrorUnauthorized)
			c.Abort()
			return
		}

		if role == constants.RoleAdmin {
			c.Next()
			return
		}

		if role == constants.RoleUser {
			utils.ForbiddenResponse(c, "You do not have access to the admin dashboard")
			c.Abort()
			return
		}

		permissionKey := readPermissionKey(module)
		if !isReadMethod(c.Request.Method) {
			permissionKey = writePermissionKey(module)
		}

		if rolePermissions == nil {
			utils.InternalServerErrorResponse(c, nil, "permission service not configured")
			c.Abort()
			return
		}

		allowed, err := rolePermissions.HasPermission(c.Request.Context(), role, permissionKey)
		if err != nil {
			utils.HandleServiceError(c, err, "failed to check permissions")
			c.Abort()
			return
		}
		if !allowed {
			utils.ForbiddenResponse(c, PermissionDeniedMessage(permissionKey))
			c.Abort()
			return
		}

		c.Next()
	}
}

func isReadMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead:
		return true
	default:
		return false
	}
}

func readPermissionKey(module string) string {
	return module + ".read"
}

func writePermissionKey(module string) string {
	return module + ".write"
}

// PermissionDeniedMessage turns a permission key into a user-facing 403 message.
func PermissionDeniedMessage(permissionKey string) string {
	parts := strings.SplitN(permissionKey, ".", 2)
	if len(parts) != 2 {
		return constants.ErrForbidden
	}
	module, action := parts[0], parts[1]
	switch action {
	case "read":
		return "You do not have read access for " + module
	case "write":
		return "You do not have write access for " + module
	default:
		return "You do not have permission for " + permissionKey
	}
}

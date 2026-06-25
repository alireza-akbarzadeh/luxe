package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/gin-gonic/gin"
)

// AuditLogger records admin audit entries (implemented by apps.Applications).
type AuditLogger interface {
	Log(ctx context.Context, entry *models.AuditLog) error
}

// AuditMiddleware records admin mutations after successful responses.
func AuditMiddleware(auditSvc AuditLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if c.Writer.Status() >= http.StatusBadRequest {
			return
		}

		method := c.Request.Method
		if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions {
			return
		}

		role, ok := GetUserRole(c)
		if !ok || role == constants.RoleUser {
			return
		}

		userID, ok := GetUserID(c)
		if !ok {
			return
		}

		resourceID := c.Param("id")
		resource := c.FullPath()
		if resource == "" {
			resource = c.Request.URL.Path
		}

		requestID := c.GetString(string(constants.RequestIDKey))
		if requestID == "" {
			requestID = c.GetHeader("X-Request-ID")
		}

		entry := &models.AuditLog{
			UserID:     userID,
			Action:     strings.ToUpper(method),
			Resource:   resource,
			ResourceID: resourceID,
			Path:       c.Request.URL.Path,
			IPAddress:  c.ClientIP(),
			RequestID:  requestID,
		}

		go auditSvc.Log(context.Background(), entry)
	}
}

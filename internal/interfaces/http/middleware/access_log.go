package middleware

import (
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
)

var accessLogSkipPaths = map[string]bool{
	"/api/v1/health":       true,
	"/api/v1/health/live":  true,
	"/api/v1/health/ready": true,
	"/openapi":             true,
}

// AccessLog writes one structured JSON line per HTTP request for log aggregation (Loki, ELK, CloudWatch).
func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		if accessLogSkipPaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()

		requestID, _ := c.Get(string(constants.RequestIDKey))
		userID, hasUser := GetUserID(c)

		fields := map[string]interface{}{
			"event":       "http_request",
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"status":      c.Writer.Status(),
			"duration_ms": time.Since(start).Milliseconds(),
			"client_ip":   c.ClientIP(),
			"request_id":  requestID,
			"bytes_out":   c.Writer.Size(),
		}
		if hasUser {
			fields["user_id"] = userID
		}
		if query := c.Request.URL.RawQuery; query != "" {
			fields["query"] = query
		}

		entry := utils.Log.WithFields(fields)
		status := c.Writer.Status()
		switch {
		case status >= 500:
			entry.Error("request completed")
		case status >= 400:
			entry.Warn("request completed")
		default:
			entry.Info("request completed")
		}
	}
}

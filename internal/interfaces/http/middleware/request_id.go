package middleware

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID assigns or propagates X-Request-ID for tracing.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = uuid.New().String()
		}
		c.Set(string(constants.RequestIDKey), id)
		c.Header("X-Request-ID", id)
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), constants.RequestIDKey, id))
		c.Next()
	}
}

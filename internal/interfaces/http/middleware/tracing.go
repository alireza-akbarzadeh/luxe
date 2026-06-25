package middleware

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// OTELMiddleware traces HTTP requests when OpenTelemetry is enabled.
func OTELMiddleware(serviceName string) gin.HandlerFunc {
	return otelgin.Middleware(serviceName)
}

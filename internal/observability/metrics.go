// Package observability provides metrics for the application.
package observability

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsHandler exposes Prometheus metrics (HTTP requests, Go runtime, etc.).
func MetricsHandler() gin.HandlerFunc {
	return gin.WrapH(promhttp.Handler())
}

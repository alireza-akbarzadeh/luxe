// Package routes contains all route definitions and middleware registrations for the shopping platform API
package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/middleware"
)

// RegisterMiddlewares attaches any custom middleware not already applied globally
func (r *Router) RegisterMiddlewares() {
	r.engine.Use(middleware.RequestID())
	if r.cfg.Observability.OTELEnabled {
		name := r.cfg.Observability.ServiceName
		if name == "" {
			name = "luxe-api"
		}
		r.engine.Use(middleware.OTELMiddleware(name))
	}
	if r.cfg.Observability.SentryEnabled {
		r.engine.Use(middleware.SentryMiddleware())
		r.engine.Use(middleware.SentryScope())
	}
	r.engine.Use(middleware.AccessLog())
	r.engine.Use(middleware.SecurityHeaders())
	r.engine.Use(middleware.AuditMiddleware(r.auditSvc))
	r.engine.Use(middleware.CORS())
	r.engine.Use(middleware.RateLimitMiddleware(100, 200))
}

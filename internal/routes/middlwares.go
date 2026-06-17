// Package routes contains all route definitions and middleware registrations for the shopping platform API
package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/middleware"
)

// RegisterMiddlewares attaches any custom middleware not already applied globally
func (r *Router) RegisterMiddlewares() {
	r.engine.Use(middleware.RequestID())
	r.engine.Use(middleware.SecurityHeaders())
	r.engine.Use(middleware.CORS())
	r.engine.Use(middleware.RateLimitMiddleware(100, 200))
}

package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
)

func SetupAuditRoutes(protected *gin.RouterGroup, ctrl *handlers.Container) {
	settings := protected.Group("/admin")
	settings.Use(middleware.ModuleGuard("settings"))
	{
		settings.GET("/audit-logs", ctrl.Audit.List)
		settings.GET("/audit-logs/summary", ctrl.Audit.Summary)
	}
}

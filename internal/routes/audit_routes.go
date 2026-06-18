package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/controllers"
	"github.com/alireza-akbarzadeh/luxe/internal/middleware"
	"github.com/gin-gonic/gin"
)

func SetupAuditRoutes(protected *gin.RouterGroup, ctrl *controllers.Container) {
	settings := protected.Group("/admin")
	settings.Use(middleware.ModuleGuard("settings"))
	{
		settings.GET("/audit-logs", ctrl.Audit.List)
		settings.GET("/audit-logs/summary", ctrl.Audit.Summary)
	}
}

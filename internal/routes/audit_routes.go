package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/controllers"
	"github.com/alireza-akbarzadeh/luxe/internal/middleware"
	"github.com/gin-gonic/gin"
)

func SetupAuditRoutes(protected *gin.RouterGroup, ctrl *controllers.Container) {
	admin := protected.Group("/admin/audit-logs")
	admin.Use(middleware.ModuleGuard("settings"))
	{
		admin.GET("/", ctrl.Audit.List)
	}
}

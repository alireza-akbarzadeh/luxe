package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/controllers"
	"github.com/alireza-akbarzadeh/luxe/internal/middleware"
	"github.com/gin-gonic/gin"
)

func SetupAdminRoutes(protected *gin.RouterGroup, ctrl *controllers.Container) {
	admin := protected.Group("/admin")
	admin.Use(middleware.RequireAdmin())
	{
		admin.GET("/stats", ctrl.Admin.GetStats)
	}
}

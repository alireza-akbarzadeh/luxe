package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/controllers"
	"github.com/alireza-akbarzadeh/luxe/internal/middleware"
	"github.com/gin-gonic/gin"
)

func SetupSettingRoutes(public, protected *gin.RouterGroup, ctrl *controllers.Container) {
	// Public read
	public.GET("/settings", ctrl.Settings.ListSettings)
	public.GET("/settings/:key", ctrl.Settings.GetSetting)

	// Admin only for write/delete
	admin := protected.Group("/settings")
	admin.Use(middleware.ModuleGuard("settings"))
	{
		admin.PUT("/:key", ctrl.Settings.SetSetting)
		admin.DELETE("/:key", ctrl.Settings.DeleteSetting)
	}
}

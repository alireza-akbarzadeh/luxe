package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
)

// SetupNavMenuRoutes registers public read and admin write nav menu endpoints.
func SetupNavMenuRoutes(public, protected *gin.RouterGroup, ctrl *handlers.Container) {
	navPublic := public.Group("/nav-menus")
	{
		navPublic.GET("", ctrl.NavMenu.GetAll)
		navPublic.GET("/:id", ctrl.NavMenu.GetByID)
	}

	navAdmin := protected.Group("/nav-menus")
	navAdmin.Use(middleware.ModuleGuard("menus"))
	{
		navAdmin.PUT("/reorder", ctrl.NavMenu.Reorder)
		navAdmin.POST("", ctrl.NavMenu.Create)
		navAdmin.PUT("/:id", ctrl.NavMenu.Update)
		navAdmin.DELETE("/:id", ctrl.NavMenu.Delete)
	}
}

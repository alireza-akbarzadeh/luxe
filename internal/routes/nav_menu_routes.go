package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/controllers"
	"github.com/gin-gonic/gin"
)

// SetupNavMenuRoutes handle nav menu api routes
func SetupNavMenuRoutes(public *gin.RouterGroup, ctrl *controllers.Container) {

	navGroup := public.Group("/nav-menus")
	{
		navGroup.GET("", ctrl.NavMenu.GetAll)
		navGroup.GET("/:id", ctrl.NavMenu.GetByID)
		navGroup.POST("", ctrl.NavMenu.Create)
		navGroup.PUT("/:id", ctrl.NavMenu.Update)
		navGroup.DELETE("/:id", ctrl.NavMenu.Delete)
	}
}

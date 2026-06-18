package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/controllers"
	"github.com/alireza-akbarzadeh/luxe/internal/middleware"
	"github.com/gin-gonic/gin"
)

func SetupInventoryRoutes(admin *gin.RouterGroup, ctrl *controllers.Container) {
	inventory := admin.Group("")
	inventory.Use(middleware.ModuleGuard("products"))
	{
		inventory.GET("/inventory/overview", ctrl.Inventory.GetOverview)
		inventory.GET("/inventory", ctrl.Inventory.List)
		inventory.POST("/inventory/adjust", ctrl.Inventory.Adjust)
		inventory.GET("/inventory/products/:id/history", ctrl.Inventory.ListHistory)
		inventory.GET("/inventory/adjustments/recent", ctrl.Inventory.ListRecent)
	}
}

package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
)

func SetupInventoryRoutes(admin *gin.RouterGroup, ctrl *handlers.Container) {
	inventory := admin.Group("")
	inventory.Use(middleware.ModuleGuard("products"))
	{
		inventory.GET("/inventory/overview", ctrl.Inventory.GetOverview)
		inventory.GET("/inventory", ctrl.Inventory.List)
		inventory.POST("/inventory/adjust", ctrl.Inventory.Adjust)
		inventory.POST("/inventory/bulk-adjust", ctrl.Inventory.BulkAdjust)
		inventory.GET("/inventory/products/:id/history", ctrl.Inventory.ListHistory)
		inventory.GET("/inventory/adjustments/recent", ctrl.Inventory.ListRecent)
	}
}

package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/controllers"
	"github.com/alireza-akbarzadeh/luxe/internal/middleware"
	"github.com/gin-gonic/gin"
)

func SetupShipmentRoutes(public, protected *gin.RouterGroup, ctrl *controllers.Container) {
	// Public (or user) shipping provider endpoints – no authentication required
	public.GET("/shipping-providers", ctrl.Shipment.GetShippingProviders)
	public.GET("/shipping-providers/:id", ctrl.Shipment.GetShippingProviderByID)

	// Authenticated user shipment endpoints
	protected.GET("/shipments/:id", ctrl.Shipment.GetShipment)
	protected.GET("/shipments", ctrl.Shipment.GetShipmentsByOrder)

	// Admin endpoints for shipments
	adminShipments := protected.Group("/admin/shipments")
	adminShipments.Use(middleware.ModuleGuard("orders"))
	{
		adminShipments.GET("", ctrl.Shipment.ListShipmentsAdmin)
		adminShipments.POST("", ctrl.Shipment.CreateShipment)
		adminShipments.GET("/:id/available-transitions", ctrl.Shipment.GetAvailableTransitions)
		adminShipments.POST("/:id/transition", ctrl.Shipment.PerformTransition)
		adminShipments.PUT("/:id/status", ctrl.Shipment.UpdateShipmentStatus)
	}

	// Admin endpoints for shipping providers (CRUD)
	adminProviders := protected.Group("/shipping-providers")
	adminProviders.Use(middleware.ModuleGuard("orders"))
	{
		adminProviders.POST("/", ctrl.Shipment.CreateShippingProvider)
		adminProviders.PUT("/:id", ctrl.Shipment.UpdateShippingProvider)
		adminProviders.DELETE("/:id", ctrl.Shipment.DeleteShippingProvider)
	}
}

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
	adminShipments := protected.Group("/shipments")
	adminShipments.Use(middleware.RequireAdmin())
	{
		adminShipments.POST("/", ctrl.Shipment.CreateShipment)
		adminShipments.PUT("/:id/status", ctrl.Shipment.UpdateShipmentStatus)
		// No DELETE /shipments – that doesn't make sense; shipments are usually not deleted.
	}

	// Admin endpoints for shipping providers (CRUD)
	adminProviders := protected.Group("/shipping-providers")
	adminProviders.Use(middleware.RequireAdmin())
	{
		adminProviders.POST("/", ctrl.Shipment.CreateShippingProvider)
		adminProviders.PUT("/:id", ctrl.Shipment.UpdateShippingProvider)
		adminProviders.DELETE("/:id", ctrl.Shipment.DeleteShippingProvider)
	}
}

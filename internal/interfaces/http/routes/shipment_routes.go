package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
)

func SetupShipmentRoutes(public, protected *gin.RouterGroup, ctrl *handlers.Container) {
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

	// Admin list — includes inactive providers (public GET is active-only).
	adminProviderList := protected.Group("/admin/shipping-providers")
	adminProviderList.Use(middleware.ModuleGuard("orders"))
	{
		adminProviderList.GET("", ctrl.Shipment.ListShippingProvidersAdmin)
	}

	// Admin mutations for shipping providers (CRUD)
	adminProviders := protected.Group("/shipping-providers")
	adminProviders.Use(middleware.ModuleGuard("orders"))
	{
		adminProviders.POST("", ctrl.Shipment.CreateShippingProvider)
		adminProviders.PUT("/:id", ctrl.Shipment.UpdateShippingProvider)
		adminProviders.DELETE("/:id", ctrl.Shipment.DeleteShippingProvider)
	}
}

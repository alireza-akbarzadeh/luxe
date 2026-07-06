package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/gin-gonic/gin"
)

// SetupReverseMarketplaceRoutes registers reverse marketplace endpoints.
func SetupReverseMarketplaceRoutes(public, protected *gin.RouterGroup, ctrl *handlers.Container) {
	public.GET("/reverse-marketplace/requests", ctrl.ReverseMarketplace.ListReverseMarketplaceRequests)
	public.GET("/reverse-marketplace/requests/:id", ctrl.ReverseMarketplace.GetReverseMarketplaceRequest)

	protected.POST("/reverse-marketplace/requests", ctrl.ReverseMarketplace.CreateReverseMarketplaceRequest)
	protected.GET("/reverse-marketplace/my-requests", ctrl.ReverseMarketplace.ListMyReverseMarketplaceRequests)

	protected.GET("/vendor/stores/:id/reverse-marketplace/requests", ctrl.ReverseMarketplace.ListVendorReverseMarketplaceRequests)
	protected.POST("/vendor/stores/:id/reverse-marketplace/requests/:requestId/offers", ctrl.ReverseMarketplace.CreateVendorReverseMarketplaceOffer)
}

package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/gin-gonic/gin"
)

// SetupBundleRoutes registers smart bundle endpoints.
func SetupBundleRoutes(public *gin.RouterGroup, ctrl *handlers.Container) {
	public.GET("/products/:id/smart-bundles", ctrl.Bundle.GetProductSmartBundles)
	public.POST("/bundles/suggest", ctrl.Bundle.SuggestSmartBundles)
}

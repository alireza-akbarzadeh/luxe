package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/gin-gonic/gin"
)

// SetupCreatorStorefrontRoutes registers public creator storefront endpoints.
func SetupCreatorStorefrontRoutes(public *gin.RouterGroup, ctrl *handlers.Container) {
	public.GET("/creators", ctrl.CreatorStorefront.ListCreators)
	public.GET("/creators/:slug", ctrl.CreatorStorefront.GetCreator)
}

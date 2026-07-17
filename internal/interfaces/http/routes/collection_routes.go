package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
)

func SetupCollectionRoutes(public, protected *gin.RouterGroup, ctrl *handlers.Container) {
	public.GET("/collections", ctrl.Collection.ListCollections)
	public.GET("/collections/slug/:slug", ctrl.Collection.GetCollectionBySlug)
	public.GET("/collections/slug/:slug/products", ctrl.Collection.GetCollectionProductsBySlug)
	public.GET("/collections/:id/preview-products", ctrl.Collection.PreviewCollectionProducts)
	public.GET("/collections/:id", ctrl.Collection.GetCollection)

	protected.POST("/collections", ctrl.Collection.CreateCollection)
	protected.POST("/collections/validate-rules", ctrl.Collection.ValidateCollectionRulesTransient)
	protected.POST("/collections/:id/validate-rules", ctrl.Collection.ValidateCollectionRules)
	protected.PUT("/collections/:id", ctrl.Collection.UpdateCollection)

	admin := protected.Group("/collections")
	admin.Use(middleware.ModuleGuard("products"))
	{
		admin.DELETE("/:id", ctrl.Collection.DeleteCollection)
	}
}

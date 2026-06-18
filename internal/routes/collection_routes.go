package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/controllers"
	"github.com/alireza-akbarzadeh/luxe/internal/middleware"
	"github.com/gin-gonic/gin"
)

func SetupCollectionRoutes(public, protected *gin.RouterGroup, ctrl *controllers.Container) {
	public.GET("/collections", ctrl.Collection.ListCollections)
	public.GET("/collections/:id", ctrl.Collection.GetCollection)

	protected.POST("/collections", ctrl.Collection.CreateCollection)
	protected.PUT("/collections/:id", ctrl.Collection.UpdateCollection)

	admin := protected.Group("/collections")
	admin.Use(middleware.ModuleGuard("products"))
	{
		admin.DELETE("/:id", ctrl.Collection.DeleteCollection)
	}
}

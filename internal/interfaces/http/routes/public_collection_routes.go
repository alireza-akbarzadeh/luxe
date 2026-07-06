package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/gin-gonic/gin"
)

// SetupPublicCollectionRoutes registers public community collection endpoints.
func SetupPublicCollectionRoutes(public *gin.RouterGroup, ctrl *handlers.Container) {
	public.GET("/public-collections", ctrl.PublicCollection.ListPublicCollections)
	public.GET("/public-collections/:slug", ctrl.PublicCollection.GetPublicCollection)
}

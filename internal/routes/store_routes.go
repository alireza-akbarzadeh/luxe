package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/controllers"
	"github.com/alireza-akbarzadeh/luxe/internal/middleware"
	"github.com/gin-gonic/gin"
)

func SetupStoreRoutes(public, protected *gin.RouterGroup, ctrl *controllers.Container) {
	// Public routes
	stores := public.Group("/stores")
	{
		stores.GET("", ctrl.Store.ListStores)
		stores.GET("/:slug", ctrl.Store.GetStore)
		stores.GET("/:slug/products", ctrl.Store.GetStoreProducts)
	}

	// Protected (authenticated) routes for follow/unfollow
	protectedStores := protected.Group("/stores")
	{
		protectedStores.POST("/:slug/follow", ctrl.Store.FollowStore)
		protectedStores.DELETE("/:slug/follow", ctrl.Store.UnfollowStore)
	}

	// Admin routes – use a separate prefix to avoid wildcard conflict
	adminStores := protected.Group("/admin/stores")
	adminStores.Use(middleware.RequireRole("admin"))
	{
		adminStores.POST("", ctrl.Store.CreateStore)
		adminStores.PUT("/:id", ctrl.Store.UpdateStore)
		adminStores.DELETE("/:id", ctrl.Store.DeleteStore)
	}
}

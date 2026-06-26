package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
)

func SetupStoreRoutes(public, protected *gin.RouterGroup, ctrl *handlers.Container) {
	// Public routes
	stores := public.Group("/stores")
	{
		stores.GET("", ctrl.Store.ListStores)
		stores.GET("/:slug", ctrl.Store.GetStore)
		stores.GET("/:slug/products", ctrl.Store.GetStoreProducts)
		stores.GET("/:slug/reviews", ctrl.Store.GetStoreReviews)
	}

	// Protected (authenticated) routes for follow/unfollow and reviews
	protectedStores := protected.Group("/stores")
	{
		protectedStores.POST("/:slug/follow", ctrl.Store.FollowStore)
		protectedStores.DELETE("/:slug/follow", ctrl.Store.UnfollowStore)
		protectedStores.GET("/:slug/reviews/me", ctrl.Store.GetMyStoreReview)
		protectedStores.POST("/:slug/reviews", ctrl.Store.CreateStoreReview)
		protectedStores.PUT("/:slug/reviews/:reviewId", ctrl.Store.UpdateStoreReview)
		protectedStores.DELETE("/:slug/reviews/:reviewId", ctrl.Store.DeleteStoreReview)
	}

	// Admin routes – use a separate prefix to avoid wildcard conflict
	adminStores := protected.Group("/admin/stores")
	adminStores.Use(middleware.ModuleGuard("products"))
	{
		adminStores.GET("/:id", ctrl.Store.GetStoreAdmin)
		adminStores.POST("", ctrl.Store.CreateStore)
		adminStores.PUT("/:id", ctrl.Store.UpdateStore)
		adminStores.DELETE("/:id", ctrl.Store.DeleteStore)
	}

	vendorStores := protected.Group("/vendor/stores")
	{
		vendorStores.GET("", ctrl.Store.ListVendorStores)
		vendorStores.POST("", ctrl.Store.CreateVendorStore)
	}
}

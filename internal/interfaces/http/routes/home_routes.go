package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/gin-gonic/gin"
)

// SetupHomeRoutes registers storefront homepage aggregation endpoints.
func SetupHomeRoutes(public, protected *gin.RouterGroup, ctrl *handlers.Container) {
	public.GET("/home", ctrl.Home.Manifest)
	public.GET("/home/categories", ctrl.Home.GetCategories)
	public.GET("/home/top-brands", ctrl.Home.GetTopBrands)
	public.GET("/home/top-products", ctrl.Home.GetTopProducts)
	public.GET("/home/trending-products", ctrl.Home.GetTrendingProducts)
	public.GET("/home/new-arrivals", ctrl.Home.GetNewArrivals)
	public.GET("/home/flash-deals", ctrl.Home.GetFlashDeals)
	public.GET("/home/hero-slides", ctrl.Home.GetHeroSlides)
	public.GET("/home/most-wishlisted", ctrl.Home.GetMostWishlisted)
	public.GET("/home/customer-favorites", ctrl.Home.GetCustomerFavorites)
	public.GET("/home/popular-collections", ctrl.Home.GetPopularCollections)
	public.GET("/home/featured-stores", ctrl.Home.GetFeaturedStores)
	public.GET("/home/shop-by-price", ctrl.Home.GetShopByPrice)
	public.GET("/home/seasonal-picks", ctrl.Home.GetSeasonalPicks)
	public.GET("/home/recently-restocked", ctrl.Home.GetRecentlyRestocked)

	protected.GET("/home/recommended", ctrl.Home.GetRecommended)
	protected.GET("/home/recently-viewed", ctrl.Home.GetRecentlyViewed)

	userMe := protected.Group("/users/me")
	{
		userMe.GET("/favorite-categories", ctrl.Home.GetFavoriteCategories)
		userMe.POST("/favorite-categories", ctrl.Home.SetFavoriteCategories)
	}

	protected.POST("/products/:id/view", ctrl.Home.RecordProductView)
}

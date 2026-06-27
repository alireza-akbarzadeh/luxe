package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/gin-gonic/gin"
)

// SetupStorefrontRoutes registers public storefront landing feeds.
func SetupStorefrontRoutes(public *gin.RouterGroup, ctrl *handlers.Container) {
	public.GET("/storefront/discounted-products", ctrl.Storefront.GetDiscountedProducts)
	public.GET("/storefront/best-sellers", ctrl.Storefront.GetBestSellers)
	public.GET("/storefront/popular-brands", ctrl.Storefront.GetPopularBrands)
	public.GET("/storefront/categories", ctrl.Storefront.GetHomeCategories)
}

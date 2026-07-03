package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/gin-gonic/gin"
)

// SetupShopLookRoutes registers storefront shop-the-look endpoints.
func SetupShopLookRoutes(public *gin.RouterGroup, ctrl *handlers.Container) {
	public.GET("/shop-looks", ctrl.ShopLook.ListShopLooks)
	public.GET("/shop-looks/:slug", ctrl.ShopLook.GetShopLook)
}

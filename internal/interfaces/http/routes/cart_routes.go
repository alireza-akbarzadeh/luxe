package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/gin-gonic/gin"
)

// SetupCartRoutes registers all cart endpoints (require JWT authentication).
func SetupCartRoutes(public, protected *gin.RouterGroup, ctrl *handlers.Container) {
	public.GET("/cart", ctrl.Cart.GetCart)
	protected.POST("/cart/items", ctrl.Cart.AddItem)
	protected.PUT("/cart/items/:id", ctrl.Cart.UpdateItem)
	protected.DELETE("/cart/items/:id", ctrl.Cart.RemoveItem)
	protected.DELETE("/cart/items", ctrl.Cart.ClearCart)
}

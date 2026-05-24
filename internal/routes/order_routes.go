package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/controllers"
	"github.com/alireza-akbarzadeh/luxe/internal/middleware"
	"github.com/gin-gonic/gin"
)

func SetupOrderRoutes(protected *gin.RouterGroup, ctrl *controllers.Container) {
	// User order endpoints (authenticated)
	protected.POST("/checkout", ctrl.Order.Checkout)
	protected.GET(constants.RouteOrders+constants.RouteOrdersMy, ctrl.Order.GetUserOrders)
	protected.GET(constants.RouteOrders+"/:id", ctrl.Order.GetOrder)

	// Admin order endpoints (require admin role)
	admin := protected.Group(constants.RouteOrders)
	admin.Use(middleware.RequireRole("admin"))
	{
		admin.GET("/", ctrl.Order.ListAllOrders)
		admin.PUT("/:id/status", ctrl.Order.UpdateOrderStatus)
	}
}

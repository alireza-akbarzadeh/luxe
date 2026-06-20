package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/controllers"
	"github.com/alireza-akbarzadeh/luxe/internal/middleware"
	"github.com/gin-gonic/gin"
)

func SetupOrderRoutes(protected *gin.RouterGroup, ctrl *controllers.Container) {
	protected.POST("/checkout", ctrl.Order.Checkout)
	protected.GET(constants.RouteOrders+constants.RouteOrdersMy, ctrl.Order.GetUserOrders)

	// Admin list must use GET "" (not GET "/") so /orders matches without a trailing slash.
	admin := protected.Group(constants.RouteOrders)
	admin.Use(middleware.ModuleGuard("orders"))
	{
		admin.GET("", ctrl.Order.ListAllOrders)
		admin.GET("/:id/available-transitions", ctrl.Order.GetAvailableTransitions)
		admin.POST("/:id/transition", ctrl.Order.PerformTransition)
		admin.PUT("/:id/status", ctrl.Order.UpdateOrderStatus)
	}

	protected.GET(constants.RouteOrders+"/:id", ctrl.Order.GetOrder)
	protected.POST(constants.RouteOrders+"/:id/cancel", ctrl.Order.CancelOrder)
}

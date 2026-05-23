package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/controllers"
	"github.com/alireza-akbarzadeh/luxe/internal/middleware"
	"github.com/gin-gonic/gin"
)

func SetupCouponRoutes(public, protected *gin.RouterGroup, ctrl *controllers.Container) {
	public.GET("/coupons", ctrl.Coupon.List)

	user := protected.Group("/coupons")
	{
		user.POST("/validate", ctrl.Coupon.Validate)
		user.GET("/my", ctrl.Coupon.GetMyCoupons)
	}

	admin := protected.Group("/coupons")
	admin.Use(middleware.RequireRole("admin"))
	{
		admin.POST("/", ctrl.Coupon.Create)
		admin.PUT("/:id", ctrl.Coupon.Update)
		admin.DELETE("/:id", ctrl.Coupon.Delete)
	}
}

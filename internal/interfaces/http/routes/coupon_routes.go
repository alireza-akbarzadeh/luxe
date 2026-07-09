package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
)

func SetupCouponRoutes(public, protected *gin.RouterGroup, ctrl *handlers.Container) {
	public.GET("/coupons", ctrl.Coupon.List)

	user := protected.Group("/coupons")
	{
		user.POST("/validate", ctrl.Coupon.Validate)
		user.GET("/my", ctrl.Coupon.GetMyCoupons)
		user.GET("/best-automatic", ctrl.Coupon.GetBestAutomatic)
	}

	admin := protected.Group("/coupons")
	admin.Use(middleware.ModuleGuard("products"))
	{
		admin.POST("", ctrl.Coupon.Create)
		admin.PUT("/:id", ctrl.Coupon.Update)
		admin.DELETE("/:id", ctrl.Coupon.Delete)
		admin.GET("/:id", ctrl.Coupon.GetCouponByID)
	}

	adminList := protected.Group("/admin/coupons")
	adminList.Use(middleware.ModuleGuard("products"))
	{
		adminList.GET("", ctrl.Coupon.ListAdmin)
	}
}

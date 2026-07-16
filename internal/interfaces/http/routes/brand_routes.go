package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
)

func SetupBrandRoutes(public, protected *gin.RouterGroup, ctrl *handlers.Container) {
	// Public (no auth needed)
	public.GET("/brands", ctrl.Brand.ListBrands)
	public.GET("/brands/slug/:slug", ctrl.Brand.GetBrandBySlug)
	public.GET("/brands/:id", ctrl.Brand.GetBrand)

	// Protected (any authenticated user)
	protected.POST("/brands", ctrl.Brand.CreateBrand)
	protected.PUT("/brands/:id", ctrl.Brand.UpdateBrand)

	// Admin only
	admin := protected.Group("/brands")
	admin.Use(middleware.ModuleGuard("products"))
	{
		admin.DELETE("/:id", ctrl.Brand.DeleteBrand)
	}
}

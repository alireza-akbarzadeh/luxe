package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/controllers"
	"github.com/alireza-akbarzadeh/luxe/internal/middleware"
	"github.com/gin-gonic/gin"
)

func SetupProductRoutes(public, protected *gin.RouterGroup, ctrl *controllers.Container) {
	// Public routes – order matters: more specific first
	public.GET("/products", ctrl.Product.List)
	public.GET("/products/:id/related", ctrl.Product.GetRelated)
	public.GET("/products/:id", ctrl.Product.GetOne)
	protected.POST("/products/:id/like", ctrl.UserLike.ToggleLike)
	protected.GET("/products/:id/liked", ctrl.UserLike.IsLikedByUser)
	protected.POST("/products/suggestions", ctrl.Product.GetProductSuggestions)

	// Admin routes (protected + admin role)
	admin := protected.Group("/products")
	admin.Use(middleware.RequireRole("admin"))
	{
		admin.POST("/", ctrl.Product.Create)
		admin.POST("/bulk", ctrl.Product.BulkCreate)
		admin.PUT("/:id", ctrl.Product.Update)
		admin.DELETE("/:id", ctrl.Product.Delete)
		admin.DELETE("/bulk", ctrl.Product.BulkDelete)
	}
}

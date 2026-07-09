package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
)

// SetupCategoryRoutes registers all category routes (public + admin)
func SetupCategoryRoutes(public, protected *gin.RouterGroup, ctrl *handlers.Container) {
	// Public category endpoints (accessible without authentication)
	public.GET(constants.RouteCategories, ctrl.Category.List)
	public.GET(constants.RouteCategories+"/:identifier", ctrl.Category.GetOne) // handles both slug and id

	// Admin category endpoints (require JWT + admin role)
	// Use a dedicated admin group to avoid route conflicts
	admin := protected.Group("/admin" + constants.RouteCategories)
	admin.Use(middleware.ModuleGuard("products"))
	{
		admin.POST("/", ctrl.Category.Create)
		admin.POST("/bulk", ctrl.Category.BulkCreate)
		admin.PUT("/reorder", ctrl.Category.Reorder)
		admin.PUT("/:id", ctrl.Category.Update)
		admin.GET("/:id", ctrl.Category.GetCategoryByID)
		admin.DELETE("/:id", ctrl.Category.Delete)
		admin.DELETE("/bulk", ctrl.Category.BulkDelete)
	}
}

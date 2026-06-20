package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/controllers"
	"github.com/alireza-akbarzadeh/luxe/internal/middleware"
	"github.com/gin-gonic/gin"
)

// SetupImportRoutes mounts Excel bulk-import endpoints under /admin/import (admin only).
func SetupImportRoutes(protected *gin.RouterGroup, ctrl *controllers.Container) {
	imp := protected.Group("/admin/import")
	imp.Use(middleware.ModuleGuard("products"))
	imp.Use(middleware.StrictRateLimit())
	{
		imp.GET("/template/:entity", ctrl.Import.DownloadTemplate)
		imp.POST("/products", ctrl.Import.ImportProducts)
		imp.POST("/categories", ctrl.Import.ImportCategories)
	}
}

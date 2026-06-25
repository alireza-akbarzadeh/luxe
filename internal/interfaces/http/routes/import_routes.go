package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
)

// SetupImportRoutes mounts Excel bulk-import endpoints under /admin/import (admin only).
func SetupImportRoutes(protected *gin.RouterGroup, ctrl *handlers.Container) {
	imp := protected.Group("/admin/import")
	imp.Use(middleware.ModuleGuard("products"))
	imp.Use(middleware.StrictRateLimit())
	{
		imp.GET("/template/:entity", ctrl.Import.DownloadTemplate)
		imp.POST("/products", ctrl.Import.ImportProducts)
		imp.POST("/categories", ctrl.Import.ImportCategories)
	}
}

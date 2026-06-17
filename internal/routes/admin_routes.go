package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/controllers"
	"github.com/alireza-akbarzadeh/luxe/internal/middleware"
	"github.com/gin-gonic/gin"
)

func SetupAdminRoutes(protected *gin.RouterGroup, ctrl *controllers.Container) {
	admin := protected.Group("/admin")
	admin.Use(middleware.RequireAdmin())
	{
			admin.GET("/stats", ctrl.Admin.GetStats)
		admin.GET("/dashboard/overview", ctrl.Admin.GetDashboardOverview)
		admin.GET("/sales-feed/snapshot", ctrl.Admin.GetSalesFeedSnapshot)
		admin.GET("/users", ctrl.Admin.ListUsers)
		admin.PATCH("/users/:id/role", ctrl.Admin.UpdateUserRole)
		admin.PATCH("/users/:id/active", ctrl.Admin.ToggleUserActive)
		admin.POST("/orders/bulk-status", ctrl.Admin.BulkUpdateOrderStatus)
		admin.GET("/orders/export", ctrl.Admin.ExportOrdersCSV)
		admin.GET("/webhooks", ctrl.Admin.ListWebhookEvents)
	}
}

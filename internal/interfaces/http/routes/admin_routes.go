package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
)

func SetupAdminRoutes(protected *gin.RouterGroup, ctrl *handlers.Container) {
	admin := protected.Group("/admin")
	admin.Use(middleware.RequireStaff())

	dashboard := admin.Group("")
	dashboard.Use(middleware.ModuleGuard("orders"))
	{
		dashboard.GET("/stats", ctrl.Admin.GetStats)
		dashboard.GET("/dashboard/overview", ctrl.Admin.GetDashboardOverview)
		dashboard.GET("/dashboard/health", ctrl.Admin.GetDashboardHealth)
		dashboard.GET("/dashboard/export", ctrl.Admin.ExportDashboardCSV)
		dashboard.GET("/reports/revenue", ctrl.Admin.GetRevenueReport)
		dashboard.GET("/sales-feed/snapshot", ctrl.Admin.GetSalesFeedSnapshot)
		dashboard.GET("/nav/preferences", ctrl.Admin.GetNavPreferences)
		dashboard.PUT("/nav/preferences", ctrl.Admin.UpdateNavPreferences)
	}

	users := admin.Group("")
	users.Use(middleware.ModuleGuard("users"))
	{
		users.GET("/users", ctrl.Admin.ListUsers)
		users.PATCH("/users/:id/role", ctrl.Admin.UpdateUserRole)
		users.PATCH("/users/:id/active", ctrl.Admin.ToggleUserActive)
	}

	orders := admin.Group("")
	orders.Use(middleware.ModuleGuard("orders"))
	{
		orders.POST("/orders/bulk-status", ctrl.Admin.BulkUpdateOrderStatus)
		orders.GET("/orders/export", ctrl.Admin.ExportOrdersCSV)
	}

	settings := admin.Group("")
	settings.Use(middleware.ModuleGuard("settings"))
	{
		settings.GET("/webhooks", ctrl.Admin.ListWebhookEvents)
	}

	SetupInventoryRoutes(admin, ctrl)

	roles := admin.Group("")
	roles.Use(middleware.ModuleGuard("roles"))
	setupRoleRoutes(roles, ctrl)
}

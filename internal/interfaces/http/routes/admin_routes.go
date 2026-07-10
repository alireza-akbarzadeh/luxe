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
		dashboard.GET("/analytics/sales", ctrl.Admin.GetSalesAnalytics)
		dashboard.GET("/ai/business-insights", ctrl.Admin.GetBusinessInsights)
		dashboard.GET("/sales-feed/snapshot", ctrl.Admin.GetSalesFeedSnapshot)
		dashboard.GET("/nav/preferences", ctrl.Admin.GetNavPreferences)
		dashboard.PUT("/nav/preferences", ctrl.Admin.UpdateNavPreferences)
	}

	users := admin.Group("")
	users.Use(middleware.ModuleGuard("users"))
	{
		users.GET("/users", ctrl.Admin.ListUsers)
		users.GET("/users/:id", ctrl.Admin.GetCustomerDetail)
		users.GET("/users/:id/addresses", ctrl.Admin.ListCustomerAddresses)
		users.PATCH("/users/:id/role", ctrl.Admin.UpdateUserRole)
		users.PATCH("/users/:id/active", ctrl.Admin.ToggleUserActive)
		users.PATCH("/users/:id/notes", ctrl.Admin.UpdateCustomerNotes)
		users.PATCH("/users/:id/segment", ctrl.Admin.UpdateCustomerSegment)
		users.GET("/customers/stats", ctrl.Admin.GetCustomerStats)
	}

	orders := admin.Group("")
	orders.Use(middleware.ModuleGuard("orders"))
	{
		orders.POST("/orders/bulk-status", ctrl.Admin.BulkUpdateOrderStatus)
		orders.GET("/orders/export", ctrl.Admin.ExportOrdersCSV)
	}

	products := admin.Group("")
	products.Use(middleware.ModuleGuard("products"))
	{
		products.GET("/products/export", ctrl.Admin.ExportProductsCSV)
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

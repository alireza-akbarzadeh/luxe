package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/controllers"
	"github.com/alireza-akbarzadeh/luxe/internal/middleware"
	"github.com/gin-gonic/gin"
)

func SetupInvoiceRoutes(protected *gin.RouterGroup, ctrl *controllers.Container) {
	admin := protected.Group("/admin/invoices")
	admin.Use(middleware.ModuleGuard("orders"))
	{
		admin.GET("", ctrl.Invoice.ListInvoicesAdmin)
		admin.GET("/:id", ctrl.Invoice.GetInvoiceAdmin)
		admin.PUT("/:id/status", ctrl.Invoice.UpdateInvoiceStatus)
	}
}

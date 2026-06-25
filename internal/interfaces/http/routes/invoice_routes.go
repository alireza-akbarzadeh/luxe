package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
)

func SetupInvoiceRoutes(protected *gin.RouterGroup, ctrl *handlers.Container) {
	admin := protected.Group("/admin/invoices")
	admin.Use(middleware.ModuleGuard("orders"))
	{
		admin.GET("", ctrl.Invoice.ListInvoicesAdmin)
		admin.GET("/:id", ctrl.Invoice.GetInvoiceAdmin)
		admin.GET("/:id/pdf", ctrl.Invoice.DownloadInvoicePDF)
		admin.POST("/:id/send", ctrl.Invoice.SendInvoiceEmail)
		admin.PUT("/:id/status", ctrl.Invoice.UpdateInvoiceStatus)
	}
}

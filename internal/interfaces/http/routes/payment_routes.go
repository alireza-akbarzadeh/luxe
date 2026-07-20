package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
)

func SetupPaymentRoutes(public, protected *gin.RouterGroup, ctrl *handlers.Container) {
	public.GET("/payments/stripe-config", ctrl.Payment.GetStripeConfig)
	protected.GET("/payment-providers", ctrl.Payment.GetPaymentProviders)

	// Admin payments management
	admin := protected.Group("/admin/payments")
	admin.Use(middleware.RequireAdmin())
	{
		admin.GET("", ctrl.Payment.ListPaymentsAdmin)
		admin.GET("/summary", ctrl.Payment.GetPaymentsSummaryAdmin)
		admin.GET("/:id", ctrl.Payment.GetPaymentAdmin)
	}

	// Admin transactions hub (combined payments + wallet ledger KPIs)
	adminTx := protected.Group("/admin/transactions")
	adminTx.Use(middleware.RequireAdmin())
	{
		adminTx.GET("/summary", ctrl.Payment.GetTransactionsHubSummary)
	}
}

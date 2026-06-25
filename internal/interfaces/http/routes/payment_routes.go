package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/gin-gonic/gin"
)

func SetupPaymentRoutes(public, protected *gin.RouterGroup, ctrl *handlers.Container) {
	public.GET("/payments/stripe-config", ctrl.Payment.GetStripeConfig)
	protected.GET("/payment-providers", ctrl.Payment.GetPaymentProviders)
}

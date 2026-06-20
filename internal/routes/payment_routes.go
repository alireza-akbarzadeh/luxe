package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/controllers"
	"github.com/gin-gonic/gin"
)

func SetupPaymentRoutes(public, protected *gin.RouterGroup, ctrl *controllers.Container) {
	public.GET("/payments/stripe-config", ctrl.Payment.GetStripeConfig)
	protected.GET("/payment-providers", ctrl.Payment.GetPaymentProviders)
}

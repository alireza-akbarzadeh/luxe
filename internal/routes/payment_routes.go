package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/controllers"
	"github.com/gin-gonic/gin"
)

func SetupPaymentRoutes(protected *gin.RouterGroup, ctrl *controllers.Container) {
	// User order endpoints (authenticated)
	protected.GET("/payment-providers", ctrl.Payment.GetPaymentProviders)

}

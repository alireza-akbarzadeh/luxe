package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/controllers"
	"github.com/gin-gonic/gin"
)

func SetupStripeRoutes(v1 *gin.RouterGroup, ctrl *controllers.Container) {
	v1.POST("/webhooks/stripe", ctrl.Stripe.Handle)
}

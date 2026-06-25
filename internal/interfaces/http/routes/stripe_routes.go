package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/gin-gonic/gin"
)

func SetupStripeRoutes(v1 *gin.RouterGroup, ctrl *handlers.Container) {
	v1.POST("/webhooks/stripe", ctrl.Stripe.Handle)
}

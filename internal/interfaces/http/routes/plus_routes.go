package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/gin-gonic/gin"
)

// SetupPlusRoutes registers Luxe Plus membership routes.
func SetupPlusRoutes(public, protected *gin.RouterGroup, ctrl *handlers.Container) {
	plus := public.Group("/plus")
	{
		plus.GET("/benefits", ctrl.Plus.GetBenefits)
	}

	protectedPlus := protected.Group("/plus")
	{
		protectedPlus.GET("/membership", ctrl.Plus.GetMembership)
		protectedPlus.POST("/subscribe", ctrl.Plus.Subscribe)
		protectedPlus.POST("/subscribe/confirm-stripe", ctrl.Plus.ConfirmStripe)
	}
}

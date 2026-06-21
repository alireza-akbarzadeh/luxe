package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/controllers"
	"github.com/gin-gonic/gin"
)

// SetupPushRoutes registers Web Push subscription endpoints.
func SetupPushRoutes(v1 *gin.RouterGroup, protected *gin.RouterGroup, ctrl *controllers.Container) {
	v1.GET("/push/vapid-public-key", ctrl.Push.GetVapidPublicKey)

	push := protected.Group("/account/push")
	{
		push.POST("/subscriptions", ctrl.Push.RegisterPushSubscription)
		push.DELETE("/subscriptions", ctrl.Push.DeletePushSubscription)
		push.POST("/test", ctrl.Push.SendTestPush)
	}
}

package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/controllers"
	"github.com/alireza-akbarzadeh/luxe/internal/middleware"
	"github.com/gin-gonic/gin"
)

// SetupWebSocketRoutes registers the realtime WebSocket upgrade and notification REST endpoints.
func SetupWebSocketRoutes(v1 *gin.RouterGroup, protected *gin.RouterGroup, ctrl *controllers.Container, cfg *config.Config) {
	v1.GET("/ws/connect", middleware.AuthMiddleware(cfg), ctrl.WebSocket.Connect)

	ws := protected.Group("/ws")
	{
		ws.GET("/notifications", ctrl.WebSocket.GetNotifications)
		ws.PUT("/notifications/:id/read", ctrl.WebSocket.MarkNotificationAsRead)
		ws.PUT("/notifications/read-all", ctrl.WebSocket.MarkAllNotificationsAsRead)
	}
}

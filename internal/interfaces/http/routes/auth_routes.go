package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
)

// SetupAuthRoutes registers all authentication routes (public + protected)
func SetupAuthRoutes(public, protected *gin.RouterGroup, ctrl *handlers.Container) {
	authPublic := public.Group("/auth")
	authPublic.Use(middleware.StrictRateLimit())
	{
		authPublic.POST("/register", ctrl.Auth.Register)
		authPublic.POST("/login", ctrl.Auth.Login)
		authPublic.POST("/forgot-password", ctrl.Auth.ForgotPassword)
		authPublic.POST("/reset-password", ctrl.Auth.ResetPassword)
		authPublic.POST("/refresh", ctrl.Auth.Refresh)
		authPublic.GET("/verify-email", ctrl.Auth.VerifyEmail)
		authPublic.POST("/logout", ctrl.Auth.Logout)
	}

	authProtected := protected.Group("/auth")
	{
		authProtected.GET("/sessions", ctrl.Auth.ListSessions)
		authProtected.DELETE("/sessions/:id", ctrl.Auth.RevokeSession)
		authProtected.DELETE("/sessions", ctrl.Auth.RevokeOtherSessions)
		authProtected.POST("/send-verification", ctrl.Auth.SendVerificationEmail)
		authProtected.POST("/change-password", ctrl.Auth.ChangePassword)
	}
}

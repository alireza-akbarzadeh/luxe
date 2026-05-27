package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/controllers"
	"github.com/gin-gonic/gin"
)

// SetupAuthRoutes registers all authentication routes (public + protected)
func SetupAuthRoutes(public, protected *gin.RouterGroup, ctrl *controllers.Container) {
	// 1. Public auth endpoints (no auth required)
	authGroup := public.Group("/auth")
	{
		authGroup.POST("/register", ctrl.Auth.Register)
		authGroup.POST("/login", ctrl.Auth.Login)
	}
	authGroup.POST("/forgot-password", ctrl.Auth.ForgotPassword)
	authGroup.POST("/reset-password", ctrl.Auth.ResetPassword)
	authGroup.POST("/refresh", ctrl.Auth.Refresh)

	authGroup.GET("/verify-email", ctrl.Auth.VerifyEmail)
	// 2. Protected auth endpoints (require a valid JWT token)
	authGroup.POST("/send-verification", ctrl.Auth.SendVerificationEmail)
	authGroup.POST("/logout", ctrl.Auth.Logout)
	authGroup.POST("/change-password", ctrl.Auth.ChangePassword)
}

package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/controllers"
	"github.com/alireza-akbarzadeh/luxe/internal/middleware"
	"github.com/gin-gonic/gin"
)

// SetupUserRoutes registers user profile endpoints (authenticated users)
// and admin user management endpoints (admin only).
func SetupUserRoutes(protected *gin.RouterGroup, ctrl *controllers.Container) {
	// Profile endpoints for authenticated users
	protected.GET("/profile", ctrl.User.GetProfile)
	protected.PUT("/profile", ctrl.User.UpdateProfile)

	// Admin user management (no "/admin" prefix – apply role middleware directly)
	adminUsers := protected.Group("/users")
	adminUsers.Use(middleware.RequireAdmin())
	{
		adminUsers.GET("/", ctrl.User.GetAllUsers)
	}

	userMe := protected.Group("/users/me")
	{
		userMe.GET("/liked-products", ctrl.UserLike.GetUserLikedProductIDs) // GET /api/v1/users/me/liked-products
	}
}

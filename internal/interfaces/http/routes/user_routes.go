package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
)

// SetupUserRoutes registers user profile endpoints (authenticated users)
// and admin user management endpoints (admin only).
func SetupUserRoutes(protected *gin.RouterGroup, ctrl *handlers.Container) {
	// Profile endpoints for authenticated users
	protected.GET("/profile", ctrl.User.GetProfile)
	protected.PUT("/profile", ctrl.User.UpdateProfile)

	// Admin user management (no "/admin" prefix – apply role middleware directly)
	adminUsers := protected.Group("/users")
	adminUsers.Use(middleware.ModuleGuard("users"))
	{
		adminUsers.GET("/", ctrl.User.GetAllUsers)
	}

	userMe := protected.Group("/users/me")
	{
		userMe.GET("/liked-products", ctrl.UserLike.GetUserLikedProductIDs) // GET /api/v1/users/me/liked-products
		userMe.GET("/reviews", ctrl.Review.GetMyReviews)
		// userMe.GET("/questions", ctrl.Pdp.GetMyQuestions)
	}
}

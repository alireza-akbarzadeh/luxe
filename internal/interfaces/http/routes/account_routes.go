package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/gin-gonic/gin"
)

// SetupAccountRoutes handle all user acounitng info
func SetupAccountRoutes(protected *gin.RouterGroup, ctrl *handlers.Container) {
	// account endpoints for authenticated users
	accountGroup := protected.Group("/account")
	{
		accountGroup.GET("/summary", ctrl.Account.GetAccountSummary)
		accountGroup.GET("/orders", ctrl.Account.GetUserOrderAccount)
		accountGroup.GET("/wishlist", ctrl.Account.GetUserWishlist)

	}
}

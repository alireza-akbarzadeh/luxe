package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/controllers"
	"github.com/alireza-akbarzadeh/luxe/internal/middleware"
	"github.com/gin-gonic/gin"
)

// SetupWalletRoutes registers wallet endpoints.
func SetupWalletRoutes(protected *gin.RouterGroup, ctrl *controllers.Container) {
	wallet := protected.Group("/wallet")
	{
		wallet.GET("/", ctrl.Wallet.GetWallet)
		wallet.POST("/deposit", ctrl.Wallet.Deposit)
		wallet.POST("/withdraw", ctrl.Wallet.Withdraw)
		wallet.GET("/transactions/:id", ctrl.Wallet.GetTransaction)
		wallet.POST("/deposit/:id/cancel", ctrl.Wallet.CancelPendingDeposit)
	}

	// Admin wallet management
	admin := protected.Group("/admin/wallet")
	admin.Use(middleware.RequireRole("admin"))
	{
		admin.POST("/adjust", ctrl.Wallet.AdminAdjust)
	}
}

package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
)

// SetupWalletRoutes registers wallet endpoints.
func SetupWalletRoutes(protected *gin.RouterGroup, ctrl *handlers.Container) {
	// Register without trailing slashes — Gin's RedirectTrailingSlash 301 breaks browser/proxy calls.
	protected.GET("/wallet", ctrl.Wallet.GetWallet)
	protected.POST("/wallet/deposit", ctrl.Wallet.Deposit)
	protected.POST("/wallet/withdraw", ctrl.Wallet.Withdraw)
	protected.GET("/wallet/transactions/:id", ctrl.Wallet.GetTransaction)
	protected.POST("/wallet/deposit/:id/cancel", ctrl.Wallet.CancelPendingDeposit)

	// Admin wallet management
	admin := protected.Group("/admin/wallet")
	admin.Use(middleware.ModuleGuard("users"))
	{
		admin.POST("/adjust", ctrl.Wallet.AdminAdjust)
	}
}

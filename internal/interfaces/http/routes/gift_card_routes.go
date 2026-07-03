package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/gin-gonic/gin"
)

// SetupGiftCardRoutes registers gift card endpoints.
func SetupGiftCardRoutes(protected *gin.RouterGroup, ctrl *handlers.Container) {
	protected.POST("/gift-cards", ctrl.GiftCard.CreateGiftCard)
	protected.POST("/gift-cards/confirm-stripe", ctrl.GiftCard.ConfirmStripePurchase)
	protected.GET("/gift-cards/sent", ctrl.GiftCard.ListSentGiftCards)
	protected.GET("/gift-cards/received", ctrl.GiftCard.ListReceivedGiftCards)
	protected.POST("/gift-cards/:code/claim", ctrl.GiftCard.ClaimGiftCard)
	protected.GET("/gift-cards/recipient-lookup", ctrl.GiftCard.LookupGiftRecipients)
	protected.POST("/gift-cards/:code/transfer", ctrl.GiftCard.TransferGiftCard)
}

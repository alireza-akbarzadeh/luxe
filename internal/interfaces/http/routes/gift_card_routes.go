package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/gin-gonic/gin"
)

// SetupGiftCardRoutes registers gift card endpoints.
func SetupGiftCardRoutes(protected *gin.RouterGroup, ctrl *handlers.Container) {
	protected.POST("/gift-cards", ctrl.GiftCard.CreateGiftCard)
	protected.GET("/gift-cards/sent", ctrl.GiftCard.ListSentGiftCards)
	protected.GET("/gift-cards/received", ctrl.GiftCard.ListReceivedGiftCards)
	protected.POST("/gift-cards/:code/claim", ctrl.GiftCard.ClaimGiftCard)
}

package handlers

import (
	"net/http"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	appgiftcard "github.com/alireza-akbarzadeh/luxe/internal/application/giftcard"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type GiftCardHandler struct {
	service  *appgiftcard.Service
	validate *validator.Validate
}

func NewGiftCardHandler(service *appgiftcard.Service) *GiftCardHandler {
	return &GiftCardHandler{service: service, validate: validator.New()}
}

// CreateGiftCard issues a new gift card giveaway.
// @Summary      Create gift card giveaway
// @Description  Creates a digital gift card sent to a recipient email
// @Tags         GiftCards
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CreateGiftCardRequest true "Gift card details"
// @Success      201 {object} utils.Response{data=dto.CreateGiftCardResponse}
// @Router       /gift-cards [post]
func (ctrl *GiftCardHandler) CreateGiftCard(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	var req dto.CreateGiftCardRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	email, _ := middleware.GetUserEmail(c)
	card, err := ctrl.service.Create(c.Request.Context(), userID, email, req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to create gift card")
		return
	}
	message := "gift card created"
	if card.CheckoutURL != "" {
		message = "complete payment at Stripe checkout"
	}
	utils.CreatedResponse(c, message, card)
}

// ConfirmStripePurchase confirms a gift card purchase after returning from Stripe Checkout.
// @Summary      Confirm Stripe gift card purchase
// @Description  Activates a pending gift card using checkout session_id from the Stripe success redirect.
// @Tags         GiftCards
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.ConfirmGiftCardStripeRequest true "Stripe session ID"
// @Success      200 {object} utils.Response{data=dto.GiftCardResponse}
// @Router       /gift-cards/confirm-stripe [post]
func (ctrl *GiftCardHandler) ConfirmStripePurchase(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	var req dto.ConfirmGiftCardStripeRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	card, err := ctrl.service.ConfirmBySessionID(c.Request.Context(), userID, req.SessionID)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to confirm gift card purchase")
		return
	}
	utils.SuccessResponse(c, "gift card purchase confirmed", card)
}

// ListSentGiftCards returns gift cards the user has given away.
// @Summary      List sent gift cards
// @Tags         GiftCards
// @Produce      json
// @Security     BearerAuth
// @Param        limit  query int false "Items per page"
// @Param        offset query int false "Offset"
// @Success      200 {object} utils.Response
// @Router       /gift-cards/sent [get]
func (ctrl *GiftCardHandler) ListSentGiftCards(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	limit, offset := paginationParams(c, constants.DefaultLimit)
	cards, total, err := ctrl.service.ListSent(c.Request.Context(), userID, limit, offset)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list sent gift cards")
		return
	}
	items := make([]dto.GiftCardResponse, len(cards))
	for i := range cards {
		items[i] = dto.ToGiftCardResponse(&cards[i])
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, gin.H{
		"gift_cards": items,
		"total":      total,
		"limit":      limit,
		"offset":     offset,
	})
}

// ListReceivedGiftCards returns gift cards addressed to the user.
// @Summary      List received gift cards
// @Tags         GiftCards
// @Produce      json
// @Security     BearerAuth
// @Param        limit  query int false "Items per page"
// @Param        offset query int false "Offset"
// @Success      200 {object} utils.Response
// @Router       /gift-cards/received [get]
func (ctrl *GiftCardHandler) ListReceivedGiftCards(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	email, _ := middleware.GetUserEmail(c)
	limit, offset := paginationParams(c, constants.DefaultLimit)
	cards, total, err := ctrl.service.ListReceived(c.Request.Context(), userID, email, limit, offset)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list received gift cards")
		return
	}
	items := make([]dto.GiftCardResponse, len(cards))
	for i := range cards {
		items[i] = dto.ToGiftCardResponse(&cards[i])
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, gin.H{
		"gift_cards": items,
		"total":      total,
		"limit":      limit,
		"offset":     offset,
	})
}

// ClaimGiftCard links a gift card code to the authenticated recipient.
// @Summary      Claim gift card
// @Tags         GiftCards
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        code path string true "Gift card code"
// @Success      200 {object} utils.Response{data=dto.GiftCardResponse}
// @Router       /gift-cards/{code}/claim [post]
func (ctrl *GiftCardHandler) ClaimGiftCard(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	email, ok := middleware.GetUserEmail(c)
	if !ok || email == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "account email is required to claim gift cards")
		return
	}
	card, err := ctrl.service.Claim(c.Request.Context(), userID, email, c.Param("code"))
	if err != nil {
		utils.HandleServiceError(c, err, "failed to claim gift card")
		return
	}
	utils.SuccessResponse(c, "gift card claimed", dto.ToGiftCardResponse(card))
}

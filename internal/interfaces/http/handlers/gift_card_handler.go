package handlers

import (
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
// @Success      201 {object} utils.Response{data=dto.GiftCardResponse}
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
	card, err := ctrl.service.Create(c.Request.Context(), userID, req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to create gift card")
		return
	}
	utils.CreatedResponse(c, "gift card created", dto.ToGiftCardResponse(card))
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
		utils.BadRequestResponse(c, "account email is required to claim gift cards")
		return
	}
	card, err := ctrl.service.Claim(c.Request.Context(), userID, email, c.Param("code"))
	if err != nil {
		utils.HandleServiceError(c, err, "failed to claim gift card")
		return
	}
	utils.SuccessResponse(c, "gift card claimed", dto.ToGiftCardResponse(card))
}

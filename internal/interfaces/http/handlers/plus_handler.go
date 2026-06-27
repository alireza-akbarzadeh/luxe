package handlers

import (
	appmembership "github.com/alireza-akbarzadeh/luxe/internal/application/membership"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// PlusHandler serves Luxe Plus membership endpoints.
type PlusHandler struct {
	membership *appmembership.Service
	validate   *validator.Validate
}

// NewPlusHandler creates a Plus handler.
func NewPlusHandler(membership *appmembership.Service) *PlusHandler {
	return &PlusHandler{
		membership: membership,
		validate:   validator.New(),
	}
}

// GetBenefits returns the public Luxe Plus benefits catalog.
// @Summary      Luxe Plus benefits
// @Description  Public catalog of Plus membership perks for the landing page.
// @Tags         Plus
// @Produce      json
// @Success      200 {object} utils.Response{data=dto.PlusBenefitsResponse}
// @Router       /plus/benefits [get]
func (h *PlusHandler) GetBenefits(c *gin.Context) {
	utils.SuccessResponse(c, constants.MsgFetchSuccess, h.membership.BenefitsCatalog())
}

// GetMembership returns the authenticated user's membership status.
// @Summary      Membership status
// @Description  Returns whether the user is on Free or active Luxe Plus, with expiry and perks.
// @Tags         Plus
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} utils.Response{data=dto.MembershipStatusResponse}
// @Failure      401 {object} utils.Response
// @Router       /plus/membership [get]
func (h *PlusHandler) GetMembership(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	status, err := h.membership.Status(c.Request.Context(), userID)
	if err != nil {
		RespondServiceError(c, err, "failed to fetch membership")
		return
	}

	utils.SuccessResponse(c, constants.MsgFetchSuccess, status)
}

// Subscribe activates Luxe Plus for one year via wallet, gift card, or Stripe Checkout.
// @Summary      Subscribe to Luxe Plus
// @Description  Pay with wallet balance, a gift card code, or Stripe Checkout (redirect to checkout_url when pending).
// @Tags         Plus
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.SubscribePlusRequest true "Payment method"
// @Success      200 {object} utils.Response{data=dto.SubscribePlusResponse}
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      409 {object} utils.Response
// @Router       /plus/subscribe [post]
func (h *PlusHandler) Subscribe(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	var req dto.SubscribePlusRequest
	if c.Request.ContentLength == 0 {
		req.PaymentMethod = constants.PlusPaymentWallet
	} else if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	email, _ := middleware.GetUserEmail(c)

	result, err := h.membership.Subscribe(c.Request.Context(), userID, email, req)
	if err != nil {
		RespondServiceError(c, err, "failed to subscribe to Luxe Plus")
		return
	}

	message := "Luxe Plus activated"
	if result.PaymentStatus == constants.PlusPaymentStatusPending {
		message = "complete payment at checkout_url to activate Luxe Plus"
	}

	utils.SuccessResponse(c, message, result)
}

package handlers

import (
	appmembership "github.com/alireza-akbarzadeh/luxe/internal/application/membership"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
)

// PlusHandler serves Luxe Plus membership endpoints.
type PlusHandler struct {
	membership *appmembership.Service
}

// NewPlusHandler creates a Plus handler.
func NewPlusHandler(membership *appmembership.Service) *PlusHandler {
	return &PlusHandler{membership: membership}
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

// Subscribe activates Luxe Plus for one year (wallet charge).
// @Summary      Subscribe to Luxe Plus
// @Description  Charges the user's wallet and activates Plus membership for one year.
// @Tags         Plus
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} utils.Response{data=dto.MembershipStatusResponse}
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

	status, err := h.membership.Subscribe(c.Request.Context(), userID)
	if err != nil {
		RespondServiceError(c, err, "failed to subscribe to Luxe Plus")
		return
	}

	utils.SuccessResponse(c, "Luxe Plus activated", status)
}

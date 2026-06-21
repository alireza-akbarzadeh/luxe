package controllers

import (
	"net/http"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/middleware"
	"github.com/alireza-akbarzadeh/luxe/internal/services"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/gin-gonic/gin"
)

type PushController struct {
	pushService services.PushServiceInterface
}

func NewPushController(pushService services.PushServiceInterface) *PushController {
	return &PushController{pushService: pushService}
}

// GetVapidPublicKey returns the VAPID public key for Web Push subscribe().
// @Summary      Get VAPID public key
// @Description  Returns the public VAPID key used by the browser PushManager.subscribe()
// @Tags         Push
// @Produce      json
// @Success      200 {object} utils.Response{data=dto.VapidPublicKeyResponse}
// @Router       /push/vapid-public-key [get]
func (pc *PushController) GetVapidPublicKey(c *gin.Context) {
	utils.SuccessResponse(c, "vapid public key", dto.VapidPublicKeyResponse{
		PublicKey: pc.pushService.GetVapidPublicKey(),
		Enabled:   pc.pushService.Enabled(),
	})
}

// RegisterPushSubscription saves a browser push subscription for the authenticated user.
// @Summary      Register push subscription
// @Description  Stores a Web Push subscription endpoint and keys for the current user
// @Tags         Push
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body dto.RegisterPushSubscriptionRequest true "Push subscription"
// @Success      201 {object} utils.Response
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Router       /account/push/subscriptions [post]
func (pc *PushController) RegisterPushSubscription(c *gin.Context) {
	if !pc.pushService.Enabled() {
		utils.ErrorResponse(c, http.StatusServiceUnavailable, "web push is not configured")
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	var req dto.RegisterPushSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	if err := pc.pushService.RegisterSubscription(
		c.Request.Context(),
		userID,
		req,
		c.Request.UserAgent(),
	); err != nil {
		utils.HandleServiceError(c, err, "failed to register push subscription")
		return
	}

	utils.CreatedResponse(c, "push subscription registered", nil)
}

// DeletePushSubscription removes a browser push subscription for the authenticated user.
// @Summary      Delete push subscription
// @Description  Removes a Web Push subscription by endpoint for the current user
// @Tags         Push
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body dto.DeletePushSubscriptionRequest true "Push subscription endpoint"
// @Success      200 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Router       /account/push/subscriptions [delete]
func (pc *PushController) DeletePushSubscription(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	var req dto.DeletePushSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	if err := pc.pushService.DeleteSubscription(c.Request.Context(), userID, req.Endpoint); err != nil {
		utils.HandleServiceError(c, err, "failed to delete push subscription")
		return
	}

	utils.SuccessResponse(c, "push subscription removed", nil)
}

// SendTestPush sends a test Web Push notification to the authenticated user's devices.
// @Summary      Send test push notification
// @Description  Sends a test push notification to all registered devices for the current user
// @Tags         Push
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} utils.Response
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Router       /account/push/test [post]
func (pc *PushController) SendTestPush(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	if err := pc.pushService.SendTestPush(c.Request.Context(), userID); err != nil {
		utils.HandleServiceError(c, err, "failed to send test push")
		return
	}

	utils.SuccessResponse(c, "test push sent", nil)
}

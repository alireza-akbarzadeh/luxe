package controllers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/services"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
)

type StripeWebhookController struct {
	paymentService  services.PaymentServiceInterface
	checkoutService services.CheckoutServiceInterface
	webhookSecret   string
}

func NewStripeWebhookController(
	paymentService services.PaymentServiceInterface,
	checkoutService services.CheckoutServiceInterface,
	cfg *config.Config,
) *StripeWebhookController {
	return &StripeWebhookController{
		paymentService:  paymentService,
		checkoutService: checkoutService,
		webhookSecret:   cfg.Stripe.WebhookSecret,
	}
}

// Handle processes Stripe webhook events (checkout.session.completed).
func (ctrl *StripeWebhookController) Handle(c *gin.Context) {
	if ctrl.webhookSecret == "" {
		utils.ErrorResponse(c, http.StatusServiceUnavailable, "stripe webhooks are not configured")
		return
	}

	const maxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes)
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "failed to read webhook body")
		return
	}

	sigHeader := c.GetHeader("Stripe-Signature")
	event, err := webhook.ConstructEvent(payload, sigHeader, ctrl.webhookSecret)
	if err != nil {
		utils.UnauthorizedResponse(c, "invalid stripe signature")
		return
	}

	switch event.Type {
	case stripe.EventTypeCheckoutSessionCompleted:
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			utils.Log.WithError(err).Error("stripe webhook: failed to parse checkout session")
			c.Status(http.StatusOK)
			return
		}

		paymentIntentID := ""
		if session.PaymentIntent != nil {
			paymentIntentID = session.PaymentIntent.ID
		}

		orderID, err := ctrl.paymentService.ConfirmStripeSession(session.ID, paymentIntentID)
		if err != nil {
			utils.Log.WithError(err).Error("stripe webhook: failed to confirm payment")
			c.Status(http.StatusOK)
			return
		}

		if err := ctrl.checkoutService.CompletePaidOrder(c.Request.Context(), orderID); err != nil {
			utils.Log.WithError(err).WithField("order_id", orderID).Error("stripe webhook: failed to complete order")
		}
	}

	c.Status(http.StatusOK)
}

package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	appcheckout "github.com/alireza-akbarzadeh/luxe/internal/application/checkout"
	appmembership "github.com/alireza-akbarzadeh/luxe/internal/application/membership"
	apppayment "github.com/alireza-akbarzadeh/luxe/internal/application/payment"
	appwallet "github.com/alireza-akbarzadeh/luxe/internal/application/wallet"
	appwebhook "github.com/alireza-akbarzadeh/luxe/internal/application/webhookevent"
	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
)

type StripeWebhookHandler struct {
	paymentService    *apppayment.Service
	checkoutService   *appcheckout.Service
	walletService     *appwallet.Service
	membershipService *appmembership.Service
	webhookCommands   *appwebhook.Commands
	webhookSecret     string
}

func NewStripeWebhookHandler(
	paymentService *apppayment.Service,
	checkoutService *appcheckout.Service,
	walletService *appwallet.Service,
	membershipService *appmembership.Service,
	webhookCommands *appwebhook.Commands,
	cfg *config.Config,
) *StripeWebhookHandler {
	return &StripeWebhookHandler{
		paymentService:    paymentService,
		checkoutService:   checkoutService,
		walletService:     walletService,
		membershipService: membershipService,
		webhookCommands:   webhookCommands,
		webhookSecret:     cfg.Stripe.WebhookSecret,
	}
}

// Handle processes Stripe webhook events with idempotency and event logging.
func (ctrl *StripeWebhookHandler) Handle(c *gin.Context) {
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

	ctx := c.Request.Context()

	ev, recordErr := ctrl.webhookCommands.Record(ctx, event.ID, string(event.Type), "stripe", payload)
	if recordErr != nil {
		utils.Log.WithField("event_id", event.ID).Info("stripe webhook: duplicate event ignored")
		c.Status(http.StatusOK)
		return
	}
	_ = ev

	var handlerErr error

	switch event.Type {
	case stripe.EventTypeCheckoutSessionCompleted:
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			utils.Log.WithError(err).Error("stripe webhook: failed to parse checkout session")
			_ = ctrl.webhookCommands.MarkFailed(ctx, event.ID, err.Error())
			c.Status(http.StatusOK)
			return
		}

		paymentIntentID := ""
		if session.PaymentIntent != nil {
			paymentIntentID = session.PaymentIntent.ID
		}

		if isWalletDepositSession(session) {
			handlerErr = ctrl.walletService.ConfirmDepositByStripeSession(ctx, session.ID, paymentIntentID)
		} else if isPlusMembershipSession(session) {
			handlerErr = ctrl.membershipService.ConfirmStripeSubscription(ctx, session)
		} else {
			orderID, err := ctrl.paymentService.ConfirmStripeSession(ctx, session.ID, paymentIntentID)
			if err == nil {
				handlerErr = ctrl.checkoutService.CompletePaidOrder(ctx, orderID)
			} else {
				handlerErr = err
			}
		}

	case stripe.EventTypeCheckoutSessionExpired:
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			utils.Log.WithError(err).Error("stripe webhook: failed to parse expired checkout session")
			_ = ctrl.webhookCommands.MarkFailed(ctx, event.ID, err.Error())
			c.Status(http.StatusOK)
			return
		}
		if isWalletDepositSession(session) {
			handlerErr = ctrl.walletService.FailDepositByStripeSession(ctx, session.ID)
		}

	default:
		// Unhandled event type — record as processed to prevent repeated delivery.
	}

	if handlerErr != nil {
		utils.Log.WithError(handlerErr).WithField("event_id", event.ID).Error("stripe webhook: processing error")
		_ = ctrl.webhookCommands.MarkFailed(ctx, event.ID, handlerErr.Error())
	} else {
		_ = ctrl.webhookCommands.MarkProcessed(ctx, event.ID)
	}

	c.Status(http.StatusOK)
}

func isWalletDepositSession(session stripe.CheckoutSession) bool {
	return session.Metadata != nil && session.Metadata["type"] == "wallet_deposit"
}

func isPlusMembershipSession(session stripe.CheckoutSession) bool {
	return session.Metadata != nil && session.Metadata["type"] == "plus_membership"
}

package stripeintegration

import (
	"fmt"
	"math"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
)

// Gateway creates Stripe Checkout Sessions for Luxe orders.
type Gateway struct {
	secretKey   string
	frontendURL string
}

func NewGateway(secretKey, frontendURL string) *Gateway {
	return &Gateway{
		secretKey:   secretKey,
		frontendURL: strings.TrimRight(frontendURL, "/"),
	}
}

// CreateCheckoutSession builds a one-time payment session for an order total.
func (g *Gateway) CreateCheckoutSession(order *models.Order, paymentID uint, customerEmail string) (checkoutURL, sessionID string, err error) {
	if g.secretKey == "" {
		return "", "", fmt.Errorf("stripe secret key is not configured")
	}

	stripe.Key = g.secretKey

	currency := strings.ToLower(order.Currency)
	if currency == "" {
		currency = "usd"
	}

	amountCents := int64(math.Round(order.TotalAmount * 100))
	if amountCents < 1 {
		return "", "", fmt.Errorf("order total must be greater than zero")
	}

	successURL := fmt.Sprintf("%s/orders/%d?payment=success&session_id={CHECKOUT_SESSION_ID}", g.frontendURL, order.ID)
	cancelURL := fmt.Sprintf("%s/checkout?payment=cancelled&order_id=%d", g.frontendURL, order.ID)

	params := &stripe.CheckoutSessionParams{
		Mode:              stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:        stripe.String(successURL),
		CancelURL:         stripe.String(cancelURL),
		ClientReferenceID: stripe.String(fmt.Sprint(order.ID)),
		CustomerEmail:     stripe.String(customerEmail),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(currency),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(fmt.Sprintf("Order %s", order.OrderNumber)),
					},
					UnitAmount: stripe.Int64(amountCents),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Metadata: map[string]string{
			"order_id":   fmt.Sprint(order.ID),
			"user_id":    fmt.Sprint(order.UserID),
			"payment_id": fmt.Sprint(paymentID),
		},
	}

	sess, err := session.New(params)
	if err != nil {
		return "", "", fmt.Errorf("stripe checkout session: %w", err)
	}

	return sess.URL, sess.ID, nil
}

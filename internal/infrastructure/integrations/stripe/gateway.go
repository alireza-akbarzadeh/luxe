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

// CreateWalletDepositSession builds a one-time Stripe Checkout session for a wallet top-up.
func (g *Gateway) CreateWalletDepositSession(userID, walletTxID uint, amount float64, currency, customerEmail string) (checkoutURL, sessionID string, err error) {
	if g.secretKey == "" {
		return "", "", fmt.Errorf("stripe secret key is not configured")
	}

	stripe.Key = g.secretKey

	currency = strings.ToLower(currency)
	if currency == "" {
		currency = "usd"
	}

	amountCents := int64(math.Round(amount * 100))
	if amountCents < 1 {
		return "", "", fmt.Errorf("deposit amount must be greater than zero")
	}

	successURL := fmt.Sprintf("%s/wallet?deposit=success&session_id={CHECKOUT_SESSION_ID}", g.frontendURL)
	cancelURL := fmt.Sprintf("%s/wallet?deposit=cancelled", g.frontendURL)

	params := &stripe.CheckoutSessionParams{
		Mode:              stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:        stripe.String(successURL),
		CancelURL:         stripe.String(cancelURL),
		ClientReferenceID: stripe.String(fmt.Sprint(walletTxID)),
		CustomerEmail:     stripe.String(customerEmail),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(currency),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("Wallet deposit"),
					},
					UnitAmount: stripe.Int64(amountCents),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Metadata: map[string]string{
			"type":                  "wallet_deposit",
			"wallet_transaction_id": fmt.Sprint(walletTxID),
			"user_id":               fmt.Sprint(userID),
		},
	}

	sess, err := session.New(params)
	if err != nil {
		return "", "", fmt.Errorf("stripe wallet deposit session: %w", err)
	}

	return sess.URL, sess.ID, nil
}

// CreatePlusMembershipSession builds a Stripe Checkout session for Luxe Plus annual membership.
func (g *Gateway) CreatePlusMembershipSession(userID uint, amount float64, currency, customerEmail string) (checkoutURL, sessionID string, err error) {
	if g.secretKey == "" {
		return "", "", fmt.Errorf("stripe secret key is not configured")
	}

	stripe.Key = g.secretKey

	currency = strings.ToLower(currency)
	if currency == "" {
		currency = "usd"
	}

	amountCents := int64(math.Round(amount * 100))
	if amountCents < 1 {
		return "", "", fmt.Errorf("membership price must be greater than zero")
	}

	successURL := fmt.Sprintf("%s/plus/landing?plus=success&session_id={CHECKOUT_SESSION_ID}", g.frontendURL)
	cancelURL := fmt.Sprintf("%s/plus/landing?plus=cancelled", g.frontendURL)

	params := &stripe.CheckoutSessionParams{
		Mode:              stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:        stripe.String(successURL),
		CancelURL:         stripe.String(cancelURL),
		ClientReferenceID: stripe.String(fmt.Sprint(userID)),
		CustomerEmail:     stripe.String(customerEmail),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(currency),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("Luxe Plus — annual membership"),
					},
					UnitAmount: stripe.Int64(amountCents),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Metadata: map[string]string{
			"type":    "plus_membership",
			"user_id": fmt.Sprint(userID),
		},
	}

	sess, err := session.New(params)
	if err != nil {
		return "", "", fmt.Errorf("stripe plus membership session: %w", err)
	}

	return sess.URL, sess.ID, nil
}

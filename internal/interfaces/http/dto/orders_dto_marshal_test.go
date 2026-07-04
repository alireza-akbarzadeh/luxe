package dto

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/stretchr/testify/require"
)

func TestCheckoutResultEnvelopeMarshal(t *testing.T) {
	order := &models.Order{ID: 42, OrderNumber: "ORD-1", TotalAmount: 99.99, Currency: "USD", Status: "pending", CreatedAt: time.Now()}
	result := &CheckoutResult{Order: order, CheckoutURL: "https://checkout.stripe.com/test", StripeSessionID: "cs_test"}
	resp := utils.Response{Success: true, Message: "order created", Data: result}
	b, err := json.Marshal(resp)
	require.NoError(t, err)

	var parsed map[string]interface{}
	require.NoError(t, json.Unmarshal(b, &parsed))
	require.Equal(t, true, parsed["success"])

	data, ok := parsed["data"].(map[string]interface{})
	require.True(t, ok, "data should be object: %s", string(b))
	require.Equal(t, float64(42), data["id"])
	require.Equal(t, "https://checkout.stripe.com/test", data["checkout_url"])
}

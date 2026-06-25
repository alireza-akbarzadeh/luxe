package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/services"
	"github.com/stretchr/testify/require"
)

func TestCheckout_MockPayment_EndToEnd(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	email := fmt.Sprintf("checkout-%s@integration.test", suffix)
	token := registerUser(t, server, email)
	product := seedProduct(t, suffix)

	addResp, err := authRequest(http.MethodPost, server.URL+"/api/v1/cart/items", token, services.AddItemRequest{
		ProductID: product.ID,
		Quantity:  2,
	})
	require.NoError(t, err)
	defer addResp.Body.Close()
	require.Equal(t, http.StatusOK, addResp.StatusCode)

	checkoutBody := dto.CheckoutRequest{
		Email:         email,
		FirstName:     "Test",
		LastName:      "User",
		AddressLine1:  "123 Test Street",
		City:          "Testville",
		State:         "CA",
		Zip:           "90210",
		Country:       "US",
		Phone:         "+14155550100",
		PaymentMethod: "mock",
		CardNumber:    "4111111111111111",
		ExpiryMonth:   12,
		ExpiryYear:    2028,
		CVV:           "123",
	}

	checkoutResp, err := authRequest(http.MethodPost, server.URL+"/api/v1/checkout", token, checkoutBody)
	require.NoError(t, err)
	defer checkoutResp.Body.Close()
	require.Equal(t, http.StatusCreated, checkoutResp.StatusCode)

	var checkoutEnvelope apiEnvelope
	require.NoError(t, json.NewDecoder(checkoutResp.Body).Decode(&checkoutEnvelope))
	require.True(t, checkoutEnvelope.Success)

	var orderPayload map[string]interface{}
	require.NoError(t, json.Unmarshal(checkoutEnvelope.Data, &orderPayload))
	orderID, ok := orderPayload["id"].(float64)
	require.True(t, ok, "checkout response should include order id")
	require.NotContains(t, orderPayload, "checkout_url", "mock checkout should not return stripe checkout_url")

	deadline := time.Now().Add(5 * time.Second)
	var paidOrder models.Order
	for time.Now().Before(deadline) {
		err := testDB.First(&paidOrder, uint(orderID)).Error
		require.NoError(t, err)
		if paidOrder.Status == constants.OrderStatusPaid {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	require.Equal(t, constants.OrderStatusPaid, paidOrder.Status)

	var payment models.Payment
	err = testDB.Where("order_id = ?", uint(orderID)).First(&payment).Error
	require.NoError(t, err)
	require.Equal(t, "succeeded", payment.Status)
	require.Equal(t, "mock", payment.Method)

	var cart models.Cart
	err = testDB.Where("user_id = ? AND status = ?", paidOrder.UserID, constants.CartStatusConverted).First(&cart).Error
	require.NoError(t, err)
}

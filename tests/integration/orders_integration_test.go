package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	appcart "github.com/alireza-akbarzadeh/luxe/internal/application/cart"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/stretchr/testify/require"
)

func TestOrders_ListAfterCheckout(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	email := fmt.Sprintf("orders-%s@integration.test", suffix)
	token := registerUser(t, server, email)
	product := seedProduct(t, suffix)

	_, err := authRequest(http.MethodPost, server.URL+"/api/v1/cart/items", token, appcart.AddItemRequest{
		ProductID: product.ID,
		Quantity:  1,
	})
	require.NoError(t, err)

	checkoutResp, err := authRequest(http.MethodPost, server.URL+"/api/v1/checkout", token, dto.CheckoutRequest{
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
	})
	require.NoError(t, err)
	defer checkoutResp.Body.Close()
	require.Equal(t, http.StatusCreated, checkoutResp.StatusCode)

	listResp, err := authRequest(http.MethodGet, server.URL+"/api/v1/orders/my?limit=10&offset=0", token, nil)
	require.NoError(t, err)
	defer listResp.Body.Close()
	require.Equal(t, http.StatusOK, listResp.StatusCode)

	var envelope apiEnvelope
	require.NoError(t, json.NewDecoder(listResp.Body).Decode(&envelope))
	require.True(t, envelope.Success)

	var data struct {
		Orders []struct {
			ID uint `json:"id"`
		} `json:"orders"`
		Total int64 `json:"total"`
	}
	require.NoError(t, json.Unmarshal(envelope.Data, &data))
	require.GreaterOrEqual(t, data.Total, int64(1))
	require.NotEmpty(t, data.Orders)
}

func TestOrders_UnauthorizedWithoutToken(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/v1/orders/my")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

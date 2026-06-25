package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	appcart "github.com/alireza-akbarzadeh/luxe/internal/application/cart"
	"github.com/stretchr/testify/require"
)

func TestCart_AddItemAndGetCart(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	token := registerUser(t, server, fmt.Sprintf("cart-%s@integration.test", suffix))
	product := seedProduct(t, suffix)

	addResp, err := authRequest(http.MethodPost, server.URL+"/api/v1/cart/items", token, appcart.AddItemRequest{
		ProductID: product.ID,
		Quantity:  2,
	})
	require.NoError(t, err)
	defer addResp.Body.Close()
	require.Equal(t, http.StatusOK, addResp.StatusCode)

	cartResp, err := authRequest(http.MethodGet, server.URL+"/api/v1/cart", token, nil)
	require.NoError(t, err)
	defer cartResp.Body.Close()
	require.Equal(t, http.StatusOK, cartResp.StatusCode)

	var envelope apiEnvelope
	require.NoError(t, json.NewDecoder(cartResp.Body).Decode(&envelope))
	require.True(t, envelope.Success)

	var cartData struct {
		Items []struct {
			ProductID uint `json:"product_id"`
			Quantity  int  `json:"quantity"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal(envelope.Data, &cartData))
	require.Len(t, cartData.Items, 1)
	require.Equal(t, product.ID, cartData.Items[0].ProductID)
	require.Equal(t, 2, cartData.Items[0].Quantity)
}

func TestCart_UpdateAndRemoveItem(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	token := registerUser(t, server, fmt.Sprintf("cart-crud-%s@integration.test", suffix))
	product := seedProduct(t, suffix)

	addResp, err := authRequest(http.MethodPost, server.URL+"/api/v1/cart/items", token, appcart.AddItemRequest{
		ProductID: product.ID,
		Quantity:  1,
	})
	require.NoError(t, err)
	defer addResp.Body.Close()
	require.Equal(t, http.StatusOK, addResp.StatusCode)

	var addEnvelope dto.AddItemResponse
	require.NoError(t, json.NewDecoder(addResp.Body).Decode(&addEnvelope))
	require.True(t, addEnvelope.Success)
	require.NotZero(t, addEnvelope.Data.ID)
	itemID := addEnvelope.Data.ID

	updateResp, err := authRequest(http.MethodPut,
		fmt.Sprintf("%s/api/v1/cart/items/%d", server.URL, itemID),
		token,
		appcart.UpdateCartItemRequest{Quantity: 3},
	)
	require.NoError(t, err)
	defer updateResp.Body.Close()
	require.Equal(t, http.StatusOK, updateResp.StatusCode)

	removeResp, err := authRequest(http.MethodDelete,
		fmt.Sprintf("%s/api/v1/cart/items/%d", server.URL, itemID),
		token,
		nil,
	)
	require.NoError(t, err)
	defer removeResp.Body.Close()
	require.Equal(t, http.StatusOK, removeResp.StatusCode)

	cartResp, err := authRequest(http.MethodGet, server.URL+"/api/v1/cart", token, nil)
	require.NoError(t, err)
	defer cartResp.Body.Close()

	var envelope apiEnvelope
	require.NoError(t, json.NewDecoder(cartResp.Body).Decode(&envelope))
	var cartData struct {
		Items []interface{} `json:"items"`
	}
	require.NoError(t, json.Unmarshal(envelope.Data, &cartData))
	require.Empty(t, cartData.Items)
}

func TestCart_UnauthorizedWithoutToken(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	resp, err := authRequest(http.MethodPost, server.URL+"/api/v1/cart/items", "", appcart.AddItemRequest{
		ProductID: 1,
		Quantity:  1,
	})
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

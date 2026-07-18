package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	appcart "github.com/alireza-akbarzadeh/luxe/internal/application/cart"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/stretchr/testify/require"
)

func TestInvoice_CreatedForPaidOrder(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	email := fmt.Sprintf("invoice-%s@integration.test", suffix)
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

	var checkoutEnvelope apiEnvelope
	require.NoError(t, json.NewDecoder(checkoutResp.Body).Decode(&checkoutEnvelope))
	require.True(t, checkoutEnvelope.Success)

	var orderPayload map[string]interface{}
	require.NoError(t, json.Unmarshal(checkoutEnvelope.Data, &orderPayload))
	orderID, ok := orderPayload["id"].(float64)
	require.True(t, ok)

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

	deadline = time.Now().Add(5 * time.Second)
	var invoice models.Invoice
	for time.Now().Before(deadline) {
		err := testDB.Where("order_id = ?", uint(orderID)).First(&invoice).Error
		if err == nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	err = testDB.Where("order_id = ?", uint(orderID)).First(&invoice).Error
	require.NoError(t, err)
	require.NotEmpty(t, invoice.InvoiceNumber)
	require.Equal(t, constants.InvoiceStatusPaid, invoice.Status)

	adminEmail := fmt.Sprintf("invoice-admin-%s@integration.test", suffix)
	adminToken := registerAdmin(t, server, adminEmail)

	listResp, err := authRequest(
		http.MethodGet,
		fmt.Sprintf("%s/api/v1/admin/invoices?order_id=%d&limit=10&offset=0", server.URL, uint(orderID)),
		adminToken,
		nil,
	)
	require.NoError(t, err)
	defer listResp.Body.Close()
	require.Equal(t, http.StatusOK, listResp.StatusCode)

	var listEnvelope apiEnvelope
	require.NoError(t, json.NewDecoder(listResp.Body).Decode(&listEnvelope))
	require.True(t, listEnvelope.Success)

	var listData struct {
		Invoices []struct {
			ID            uint   `json:"id"`
			OrderID       uint   `json:"order_id"`
			InvoiceNumber string `json:"invoice_number"`
			Status        string `json:"status"`
		} `json:"invoices"`
		Total int64 `json:"total"`
	}
	require.NoError(t, json.Unmarshal(listEnvelope.Data, &listData))
	require.GreaterOrEqual(t, listData.Total, int64(1))
	require.NotEmpty(t, listData.Invoices)
	require.Equal(t, uint(orderID), listData.Invoices[0].OrderID)
	require.Equal(t, invoice.InvoiceNumber, listData.Invoices[0].InvoiceNumber)
	require.Equal(t, constants.InvoiceStatusPaid, listData.Invoices[0].Status)
}

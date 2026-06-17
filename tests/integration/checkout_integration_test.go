package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/controllers"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/routes"
	"github.com/alireza-akbarzadeh/luxe/internal/services"
	"github.com/alireza-akbarzadeh/luxe/internal/tasks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type registerResponse struct {
	Success bool `json:"success"`
	Data    struct {
		AccessToken string `json:"access_token"`
	} `json:"data"`
}

type apiEnvelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Message string          `json:"message"`
}

func newTestServer(t *testing.T) *httptest.Server {
	jobQueue, err := tasks.NewJobQueue(testCfg, tasks.Handlers{})
	require.NoError(t, err)

	svc := services.NewServices(testDB, testCfg, jobQueue)
	tasks.BindHandlers(jobQueue, svc.JobHandlers())
	require.NoError(t, jobQueue.Start())
	t.Cleanup(func() { jobQueue.Shutdown() })

	ctrl := controllers.NewContainer(testDB, svc, testCfg)

	engine := gin.New()
	engine.Use(gin.Recovery())
	router := routes.NewRouter(engine, ctrl, testCfg, svc.Audit)
	router.Setup()

	return httptest.NewServer(engine)
}

func seedProduct(t *testing.T, suffix string) *models.Product {
	store := models.Store{
		Name:   "Integration Store " + suffix,
		Slug:   "integration-store-" + suffix,
		Status: "active",
	}
	require.NoError(t, testDB.Create(&store).Error)

	product := models.Product{
		Name:           "Integration Product " + suffix,
		Price:          29.99,
		Stock:          100,
		SKU:            "SKU-" + suffix,
		Slug:           "product-" + suffix,
		Status:         "active",
		StoreID:        store.ID,
		TrackInventory: true,
		AllowBackorder: false,
	}
	require.NoError(t, testDB.Create(&product).Error)
	return &product
}

func registerUser(t *testing.T, server *httptest.Server, email string) string {
	body, err := json.Marshal(dto.RegisterRequest{
		Email:     email,
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
	})
	require.NoError(t, err)

	resp, err := http.Post(server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var reg registerResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&reg))
	require.True(t, reg.Success)
	require.NotEmpty(t, reg.Data.AccessToken)
	return reg.Data.AccessToken
}

func authRequest(t *testing.T, method, url, token string, payload interface{}) *http.Response {
	var body *bytes.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		require.NoError(t, err)
		body = bytes.NewReader(raw)
	} else {
		body = bytes.NewReader(nil)
	}

	req, err := http.NewRequest(method, url, body)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func TestCheckout_MockPayment_EndToEnd(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	email := fmt.Sprintf("checkout-%s@integration.test", suffix)
	token := registerUser(t, server, email)
	product := seedProduct(t, suffix)

	addResp := authRequest(t, http.MethodPost, server.URL+"/api/v1/cart/items", token, services.AddItemRequest{
		ProductID: product.ID,
		Quantity:  2,
	})
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

	checkoutResp := authRequest(t, http.MethodPost, server.URL+"/api/v1/checkout", token, checkoutBody)
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
	err := testDB.Where("order_id = ?", uint(orderID)).First(&payment).Error
	require.NoError(t, err)
	require.Equal(t, "succeeded", payment.Status)
	require.Equal(t, "mock", payment.Method)

	var cart models.Cart
	err = testDB.Where("user_id = ? AND status = ?", paidOrder.UserID, constants.CartStatusConverted).First(&cart).Error
	require.NoError(t, err)
}

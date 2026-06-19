package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/controllers"
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
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		User         struct {
			ID    uint   `json:"id"`
			Email string `json:"email"`
		} `json:"user"`
	} `json:"data"`
}

type loginResponse struct {
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
	router := routes.NewRouter(engine, ctrl, testCfg, svc.Audit, svc.Role)
	router.Setup()

	return httptest.NewServer(engine)
}

func postJSON(url string, payload interface{}) (*http.Response, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return http.Post(url, "application/json", bytes.NewReader(body))
}

func authRequest(method, url, token string, payload interface{}) (*http.Response, error) {
	var body *bytes.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(raw)
	} else {
		body = bytes.NewReader(nil)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return http.DefaultClient.Do(req)
}

func registerUser(t *testing.T, server *httptest.Server, email string) string {
	token, _ := registerUserWithID(t, server, email)
	return token
}

func registerUserWithID(t *testing.T, server *httptest.Server, email string) (token string, userID uint) {
	t.Helper()
	resp, err := postJSON(server.URL+"/api/v1/auth/register", dto.RegisterRequest{
		Email:     email,
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
	})
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var reg registerResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&reg))
	require.True(t, reg.Success)
	require.NotEmpty(t, reg.Data.AccessToken)
	return reg.Data.AccessToken, reg.Data.User.ID
}

func loginAdminToken(t *testing.T, server *httptest.Server, email string) string {
	t.Helper()
	loginResp, err := postJSON(server.URL+"/api/v1/auth/login", map[string]string{
		"email":    email,
		"password": "password123",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, loginResp.StatusCode)

	var login loginResponse
	require.NoError(t, json.NewDecoder(loginResp.Body).Decode(&login))
	loginResp.Body.Close()
	require.NotEmpty(t, login.Data.AccessToken)
	return login.Data.AccessToken
}

func promoteToAdmin(t *testing.T, email string) {
	t.Helper()
	require.NoError(t, testDB.Model(&models.User{}).
		Where("email = ?", email).
		Update("role", constants.RoleAdmin).Error)
}

func registerAdmin(t *testing.T, server *httptest.Server, email string) string {
	t.Helper()
	registerUser(t, server, email)
	promoteToAdmin(t, email)
	return loginAdminToken(t, server, email)
}

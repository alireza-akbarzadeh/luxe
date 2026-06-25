package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/stretchr/testify/require"
)

func TestAuth_RegisterAndLogin(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	email := fmt.Sprintf("auth-%s@integration.test", suffix)
	password := "password123"

	regResp, err := postJSON(server.URL+"/api/v1/auth/register", dto.RegisterRequest{
		Email:     email,
		Password:  password,
		FirstName: "Auth",
		LastName:  "Tester",
	})
	require.NoError(t, err)
	defer regResp.Body.Close()
	require.Equal(t, http.StatusCreated, regResp.StatusCode)

	var reg registerResponse
	require.NoError(t, json.NewDecoder(regResp.Body).Decode(&reg))
	require.True(t, reg.Success)
	require.NotEmpty(t, reg.Data.AccessToken)
	require.NotEmpty(t, reg.Data.RefreshToken)
	require.Equal(t, email, reg.Data.User.Email)

	loginResp, err := postJSON(server.URL+"/api/v1/auth/login", dto.LoginRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)
	defer loginResp.Body.Close()
	require.Equal(t, http.StatusOK, loginResp.StatusCode)

	var login loginResponse
	require.NoError(t, json.NewDecoder(loginResp.Body).Decode(&login))
	require.True(t, login.Success)
	require.NotEmpty(t, login.Data.AccessToken)
}

func TestAuth_Login_InvalidPassword(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	email := fmt.Sprintf("auth-bad-%s@integration.test", suffix)
	registerUser(t, server, email)

	loginResp, err := postJSON(server.URL+"/api/v1/auth/login", dto.LoginRequest{
		Email:    email,
		Password: "wrongpassword",
	})
	require.NoError(t, err)
	defer loginResp.Body.Close()
	require.Equal(t, http.StatusUnauthorized, loginResp.StatusCode)
}

func TestAuth_RegisterDuplicateEmail(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	email := fmt.Sprintf("dup-%s@integration.test", suffix)

	first, err := postJSON(server.URL+"/api/v1/auth/register", dto.RegisterRequest{
		Email:     email,
		Password:  "password123",
		FirstName: "Dup",
		LastName:  "User",
	})
	require.NoError(t, err)
	defer first.Body.Close()
	require.Equal(t, http.StatusCreated, first.StatusCode)

	second, err := postJSON(server.URL+"/api/v1/auth/register", dto.RegisterRequest{
		Email:     email,
		Password:  "password123",
		FirstName: "Dup",
		LastName:  "Again",
	})
	require.NoError(t, err)
	defer second.Body.Close()
	require.Equal(t, http.StatusConflict, second.StatusCode)
}

func TestAuth_Login_MissingPassword(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	resp, err := postJSON(server.URL+"/api/v1/auth/login", dto.LoginRequest{
		Email: "nobody@integration.test",
	})
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

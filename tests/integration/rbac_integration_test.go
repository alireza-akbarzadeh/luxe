package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/stretchr/testify/require"
)

func decodeJSON(r *http.Response, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

// TestRBAC_AdminStats_NoToken returns 401 when there is no auth token.
func TestRBAC_AdminStats_NoToken(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/v1/admin/stats")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// TestRBAC_AdminStats_RegularUser returns 403 when a non-admin calls an admin endpoint.
func TestRBAC_AdminStats_RegularUser(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	email := fmt.Sprintf("rbac-user-%s@integration.test", suffix)
	token := registerUser(t, server, email)

	resp, err := authRequest(http.MethodGet, server.URL+"/api/v1/admin/stats", token, nil)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

// TestRBAC_AdminStats_AdminUser returns 200 when an admin calls the endpoint.
// The test registers a user, elevates its role directly in the DB, then logs in
// again so the issued JWT carries the admin role claim.
func TestRBAC_AdminStats_AdminUser(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	email := fmt.Sprintf("rbac-admin-%s@integration.test", suffix)
	password := "password123"

	// Register → regular user.
	registerUser(t, server, email)

	// Elevate role in DB directly.
	require.NoError(t, testDB.Model(&models.User{}).
		Where("email = ?", email).
		Update("role", constants.RoleAdmin).Error)

	// Login again so the JWT carries the admin role.
	loginResp, err := postJSON(server.URL+"/api/v1/auth/login", map[string]string{
		"email":    email,
		"password": password,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, loginResp.StatusCode)

	var login loginResponse
	require.NoError(t, json.NewDecoder(loginResp.Body).Decode(&login))
	loginResp.Body.Close()
	adminToken := login.Data.AccessToken

	resp, err := authRequest(http.MethodGet, server.URL+"/api/v1/admin/stats", adminToken, nil)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var envelope apiEnvelope
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&envelope))
	require.True(t, envelope.Success)
}

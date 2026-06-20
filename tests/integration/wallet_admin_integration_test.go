package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWallet_AdminAdjust_CreditsUserBalance(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	userEmail := fmt.Sprintf("wallet-target-%s@integration.test", suffix)
	userToken, userID := registerUserWithID(t, server, userEmail)

	adminEmail := fmt.Sprintf("wallet-admin-%s@integration.test", suffix)
	adminToken := registerAdmin(t, server, adminEmail)

	adjustResp, err := authRequest(http.MethodPost, server.URL+"/api/v1/admin/wallet/adjust", adminToken, map[string]interface{}{
		"user_id":     userID,
		"amount":      42.5,
		"description": "integration test credit",
	})
	require.NoError(t, err)
	defer adjustResp.Body.Close()
	require.Equal(t, http.StatusOK, adjustResp.StatusCode)

	var adjustEnvelope apiEnvelope
	require.NoError(t, json.NewDecoder(adjustResp.Body).Decode(&adjustEnvelope))
	require.True(t, adjustEnvelope.Success)

	walletResp, err := authRequest(http.MethodGet, server.URL+"/api/v1/wallet", userToken, nil)
	require.NoError(t, err)
	defer walletResp.Body.Close()
	require.Equal(t, http.StatusOK, walletResp.StatusCode)

	var walletEnvelope apiEnvelope
	require.NoError(t, json.NewDecoder(walletResp.Body).Decode(&walletEnvelope))
	require.True(t, walletEnvelope.Success)

	var walletData struct {
		Balance float64 `json:"balance"`
	}
	require.NoError(t, json.Unmarshal(walletEnvelope.Data, &walletData))
	require.Equal(t, 42.5, walletData.Balance)
}

func TestWallet_AdminAdjust_ForbiddenForRegularUser(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	email := fmt.Sprintf("wallet-forbidden-%s@integration.test", suffix)
	token, userID := registerUserWithID(t, server, email)

	resp, err := authRequest(http.MethodPost, server.URL+"/api/v1/admin/wallet/adjust", token, map[string]interface{}{
		"user_id":     userID,
		"amount":      10,
		"description": "should fail",
	})
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

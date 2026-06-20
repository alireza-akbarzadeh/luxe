package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWallet_GetBalance_NewUser(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	email := fmt.Sprintf("wallet-%s@integration.test", suffix)
	token := registerUser(t, server, email)

	resp, err := authRequest(http.MethodGet, server.URL+"/api/v1/wallet", token, nil)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var envelope apiEnvelope
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&envelope))
	require.True(t, envelope.Success)

	var data struct {
		Balance float64 `json:"balance"`
	}
	require.NoError(t, json.Unmarshal(envelope.Data, &data))
	require.Equal(t, 0.0, data.Balance)
}

func TestWallet_Deposit_IncreasesBalance(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	email := fmt.Sprintf("wallet-dep-%s@integration.test", suffix)
	token := registerUser(t, server, email)

	depResp, err := authRequest(http.MethodPost, server.URL+"/api/v1/wallet/deposit", token, map[string]interface{}{
		"amount": 25.50,
	})
	require.NoError(t, err)
	defer depResp.Body.Close()
	require.Equal(t, http.StatusOK, depResp.StatusCode)

	var depEnvelope apiEnvelope
	require.NoError(t, json.NewDecoder(depResp.Body).Decode(&depEnvelope))
	require.True(t, depEnvelope.Success)

	var depData struct {
		Status string `json:"status"`
	}
	require.NoError(t, json.Unmarshal(depEnvelope.Data, &depData))
	require.Equal(t, "completed", depData.Status, "mock deposit when Stripe is disabled")

	walletResp, err := authRequest(http.MethodGet, server.URL+"/api/v1/wallet", token, nil)
	require.NoError(t, err)
	defer walletResp.Body.Close()
	require.Equal(t, http.StatusOK, walletResp.StatusCode)

	var envelope apiEnvelope
	require.NoError(t, json.NewDecoder(walletResp.Body).Decode(&envelope))
	require.True(t, envelope.Success)

	var data struct {
		Balance float64 `json:"balance"`
	}
	require.NoError(t, json.Unmarshal(envelope.Data, &data))
	require.Equal(t, 25.50, data.Balance)
}

func TestWallet_UnauthorizedWithoutToken(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/v1/wallet")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

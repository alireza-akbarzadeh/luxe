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

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

func TestCouponWorkflow_ActivePauseResume(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	adminEmail := fmt.Sprintf("coupon-admin-%s@integration.test", suffix)
	adminToken := registerAdmin(t, server, adminEmail)

	code := fmt.Sprintf("INT-%s", suffix)
	now := time.Now().UTC()
	isActive := true
	createResp, err := authRequest(http.MethodPost, server.URL+"/api/v1/coupons", adminToken, dto.CreateCouponRequest{
		Code:           code,
		Description:    "integration coupon",
		DiscountType:   "percentage",
		DiscountValue:  10,
		UsageLimit:     100,
		IsActive:       &isActive,
		StartDate:      now,
		EndDate:        now.Add(30 * 24 * time.Hour),
	})
	require.NoError(t, err)
	defer createResp.Body.Close()
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var createBody struct {
		Success bool `json:"success"`
		Data    struct {
			Coupon struct {
				ID            uint `json:"id"`
				WorkflowState *struct {
					Code string `json:"code"`
				} `json:"workflow_state"`
			} `json:"coupon"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(createResp.Body).Decode(&createBody))
	require.True(t, createBody.Success)
	require.NotZero(t, createBody.Data.Coupon.ID)
	require.NotNil(t, createBody.Data.Coupon.WorkflowState)
	require.Equal(t, "active", createBody.Data.Coupon.WorkflowState.Code)

	couponID := createBody.Data.Coupon.ID

	pauseResp, err := authRequest(
		http.MethodPost,
		fmt.Sprintf("%s/api/v1/workflows/coupon/%d/transition", server.URL, couponID),
		adminToken,
		map[string]string{"event": "pause"},
	)
	require.NoError(t, err)
	defer pauseResp.Body.Close()
	require.Equal(t, http.StatusOK, pauseResp.StatusCode)

	var pauseEnvelope apiEnvelope
	require.NoError(t, json.NewDecoder(pauseResp.Body).Decode(&pauseEnvelope))
	require.True(t, pauseEnvelope.Success)

	var pauseData struct {
		ToState struct {
			Code string `json:"code"`
		} `json:"to_state"`
	}
	require.NoError(t, json.Unmarshal(pauseEnvelope.Data, &pauseData))
	require.Equal(t, "paused", pauseData.ToState.Code)

	resumeResp, err := authRequest(
		http.MethodPost,
		fmt.Sprintf("%s/api/v1/workflows/coupon/%d/transition", server.URL, couponID),
		adminToken,
		map[string]string{"event": "resume"},
	)
	require.NoError(t, err)
	defer resumeResp.Body.Close()
	require.Equal(t, http.StatusOK, resumeResp.StatusCode)

	var resumeEnvelope apiEnvelope
	require.NoError(t, json.NewDecoder(resumeResp.Body).Decode(&resumeEnvelope))
	require.True(t, resumeEnvelope.Success)

	var resumeData struct {
		ToState struct {
			Code string `json:"code"`
		} `json:"to_state"`
	}
	require.NoError(t, json.Unmarshal(resumeEnvelope.Data, &resumeData))
	require.Equal(t, "active", resumeData.ToState.Code)
}

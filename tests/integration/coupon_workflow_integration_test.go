package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/stretchr/testify/require"
)

func TestCouponWorkflow_CreateDraftThenActivate(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	adminEmail := fmt.Sprintf("coupon-admin-%s@integration.test", suffix)
	adminToken := registerAdmin(t, server, adminEmail)

	code := fmt.Sprintf("INT-%s", suffix)
	now := time.Now().UTC()
	isActive := false
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
	require.Equal(t, "draft", createBody.Data.Coupon.WorkflowState.Code)

	couponID := createBody.Data.Coupon.ID

	transitionResp, err := authRequest(
		http.MethodPost,
		fmt.Sprintf("%s/api/v1/workflows/coupon/%d/transition", server.URL, couponID),
		adminToken,
		map[string]string{"event": "activate"},
	)
	require.NoError(t, err)
	defer transitionResp.Body.Close()
	require.Equal(t, http.StatusOK, transitionResp.StatusCode)

	var transitionEnvelope apiEnvelope
	require.NoError(t, json.NewDecoder(transitionResp.Body).Decode(&transitionEnvelope))
	require.True(t, transitionEnvelope.Success)

	var transitionData struct {
		ToState struct {
			Code string `json:"code"`
		} `json:"to_state"`
	}
	require.NoError(t, json.Unmarshal(transitionEnvelope.Data, &transitionData))
	require.Equal(t, "active", transitionData.ToState.Code)

	getResp, err := authRequest(http.MethodGet, fmt.Sprintf("%s/api/v1/coupons/%d", server.URL, couponID), adminToken, nil)
	require.NoError(t, err)
	defer getResp.Body.Close()
	require.Equal(t, http.StatusOK, getResp.StatusCode)

	var getBody struct {
		Success bool `json:"success"`
		Data    struct {
			Coupon struct {
				IsActive      bool `json:"is_active"`
				WorkflowState *struct {
					Code string `json:"code"`
				} `json:"workflow_state"`
			} `json:"coupon"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(getResp.Body).Decode(&getBody))
	require.True(t, getBody.Success)
	require.True(t, getBody.Data.Coupon.IsActive)
	require.NotNil(t, getBody.Data.Coupon.WorkflowState)
	require.Equal(t, "active", getBody.Data.Coupon.WorkflowState.Code)
}

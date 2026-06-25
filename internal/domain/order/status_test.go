package order

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCanCancel(t *testing.T) {
	require.ErrorIs(t, CanCancel(Order{Status: StatusDelivered}), ErrCannotCancel)
	require.NoError(t, CanCancel(Order{Status: StatusPending}))
	require.NoError(t, CanCancel(Order{Status: StatusPaid}))
}

func TestIsValidBulkAdminStatus(t *testing.T) {
	assert.True(t, IsValidBulkAdminStatus(StatusShipped))
	assert.False(t, IsValidBulkAdminStatus(StatusPending))
}

func TestShouldRefundWalletOnCancel(t *testing.T) {
	assert.True(t, ShouldRefundWalletOnCancel(PaymentMethodWallet, PaymentStatusSucceeded))
	assert.False(t, ShouldRefundWalletOnCancel("stripe", PaymentStatusSucceeded))
}

func TestWorkflowStateCode(t *testing.T) {
	assert.Equal(t, "paid", WorkflowStateCode(StatusPaid))
	assert.Equal(t, "", WorkflowStateCode("unknown"))
}

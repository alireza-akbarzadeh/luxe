package services

import (
	"context"
	"testing"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/stretchr/testify/assert"
)

func TestApplyOrderWorkflow_NilEngine(t *testing.T) {
	t.Parallel()
	ok := applyOrderWorkflow(context.Background(), nil, 1, constants.OrderStatusShipped, constants.RoleAdmin, nil)
	assert.False(t, ok)
}

func TestApplyProductWorkflow_NilEngine(t *testing.T) {
	t.Parallel()
	ok := applyProductWorkflow(context.Background(), nil, 1, constants.ProductStatusActive, constants.RoleAdmin, nil)
	assert.False(t, ok)
}

func TestApplyShipmentWorkflow_NilEngine(t *testing.T) {
	t.Parallel()
	assert.NotPanics(t, func() {
		applyShipmentWorkflow(context.Background(), nil, 1, "delivered")
	})
}

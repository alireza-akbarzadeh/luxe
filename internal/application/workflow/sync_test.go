package workflow

import (
	"context"
	"testing"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/stretchr/testify/assert"
)

func TestApplyOrderWorkflow_NilEngine(t *testing.T) {
	t.Parallel()
	ok := ApplyOrderWorkflow(context.Background(), nil, 1, constants.OrderStatusShipped, constants.RoleAdmin, nil)
	assert.False(t, ok)
}

func TestApplyProductWorkflow_NilEngine(t *testing.T) {
	t.Parallel()
	ok := ApplyProductWorkflow(context.Background(), nil, 1, constants.ProductStatusActive, constants.RoleAdmin, nil)
	assert.False(t, ok)
}

func TestApplyShipmentWorkflow_NilEngine(t *testing.T) {
	t.Parallel()
	assert.NotPanics(t, func() {
		ApplyShipmentWorkflow(context.Background(), nil, 1, "delivered")
	})
}

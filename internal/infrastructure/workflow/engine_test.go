package workflow

import (
	"testing"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestMirrorStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		entityType string
		code       string
		want       string
	}{
		{constants.WorkflowEntityOrder, "created", "pending"},
		{constants.WorkflowEntityOrder, "pending_payment", "pending"},
		{constants.WorkflowEntityOrder, "paid", "paid"},
		{constants.WorkflowEntityProduct, "published", "active"},
		{constants.WorkflowEntityProduct, "under_review", "inactive"},
		{constants.WorkflowEntityShipment, "in_transit", "shipped"},
		{constants.WorkflowEntityReturn, "requested", "requested"},
	}

	for _, tt := range tests {
		t.Run(tt.entityType+"_"+tt.code, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, mirrorStatus(tt.entityType, tt.code))
		})
	}
}

func TestEngineCheckRole(t *testing.T) {
	t.Parallel()

	engine := NewEngine(nil)
	trans := &models.WorkflowTransition{RequiredRole: constants.RoleAdmin, Event: "approve"}

	assert.NoError(t, engine.checkRole(trans, constants.RoleAdmin))
	assert.Error(t, engine.checkRole(trans, constants.RoleUser))

	open := &models.WorkflowTransition{Event: "cancel"}
	assert.NoError(t, engine.checkRole(open, constants.RoleUser))
}

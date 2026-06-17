package services

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/services/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
)

// applyWorkflowEvent runs a validated transition. Returns the engine error when
// the transition is rejected; nil when the engine is unset.
func applyWorkflowEvent(ctx context.Context, engine *workflow.Engine, req workflow.TransitionRequest) error {
	if engine == nil {
		return nil
	}
	_, err := engine.Transition(ctx, req)
	return err
}

// syncWorkflowState forces an entity to a target state (audit-only, no guards/hooks).
func syncWorkflowState(ctx context.Context, engine *workflow.Engine, workflowKey string, entityID uint, stateCode, event string, actorID *uint) {
	if engine == nil {
		return
	}
	if _, err := engine.SetState(ctx, workflow.SetStateRequest{
		WorkflowKey:     workflowKey,
		EntityID:        entityID,
		TargetStateCode: stateCode,
		Event:           event,
		ActorID:         actorID,
	}); err != nil {
		utils.Log.WithError(err).
			WithField("workflow", workflowKey).
			WithField("entity_id", entityID).
			WithField("state", stateCode).
			Warn("failed to sync workflow state")
	}
}

// syncUserWorkflowState updates the user account workflow pointer (best-effort).
func syncUserWorkflowState(ctx context.Context, engine *workflow.Engine, userID uint, stateCode, event string, actorID *uint) {
	syncWorkflowState(ctx, engine, constants.WorkflowEntityUser, userID, stateCode, event, actorID)
}

// orderStatusToEvent maps legacy admin status strings to seeded workflow events.
var orderStatusToEvent = map[string]string{
	constants.OrderStatusPaid:      "payment_succeeded",
	"processing":                   "start_processing",
	constants.OrderStatusShipped:   "ship",
	constants.OrderStatusDelivered: "deliver",
	"completed":                    "complete",
	constants.OrderStatusCancelled: "cancel",
	constants.OrderStatusRefunded:  "refund",
}

// applyOrderWorkflow tries a validated transition first, then falls back to SetState.
// Returns true when the engine updated the order (status mirror included).
func applyOrderWorkflow(
	ctx context.Context,
	engine *workflow.Engine,
	orderID uint,
	status, actorRole string,
	actorID *uint,
) bool {
	if engine == nil {
		return false
	}

	if event, ok := orderStatusToEvent[status]; ok {
		if err := applyWorkflowEvent(ctx, engine, workflow.TransitionRequest{
			WorkflowKey: constants.WorkflowEntityOrder,
			EntityID:    orderID,
			Event:       event,
			ActorID:     actorID,
			ActorRole:   actorRole,
		}); err == nil {
			return true
		}
	}

	code, ok := map[string]string{
		constants.OrderStatusPending:   "pending_payment",
		constants.OrderStatusPaid:      "paid",
		"processing":                   "processing",
		constants.OrderStatusShipped:   "shipped",
		constants.OrderStatusDelivered: "delivered",
		"completed":                    "completed",
		constants.OrderStatusCancelled: "cancelled",
		constants.OrderStatusRefunded:  "refunded",
	}[status]
	if !ok {
		return false
	}

	_, err := engine.SetState(ctx, workflow.SetStateRequest{
		WorkflowKey:     constants.WorkflowEntityOrder,
		EntityID:        orderID,
		TargetStateCode: code,
		Event:           "admin_set",
		ActorID:         actorID,
	})
	return err == nil
}

// shipmentStatusToEvent maps legacy shipment statuses to workflow events.
var shipmentStatusToEvent = map[string]string{
	"processing": "ready",
	"shipped":    "depart",
	"delivered":  "deliver",
}

// applyShipmentWorkflow tries transition then SetState for a legacy shipment status.
func applyShipmentWorkflow(ctx context.Context, engine *workflow.Engine, shipmentID uint, status string) {
	if engine == nil {
		return
	}
	if event, ok := shipmentStatusToEvent[status]; ok {
		if err := applyWorkflowEvent(ctx, engine, workflow.TransitionRequest{
			WorkflowKey: constants.WorkflowEntityShipment,
			EntityID:    shipmentID,
			Event:       event,
		}); err == nil {
			return
		}
	}
	code, ok := map[string]string{
		"pending":    "pending",
		"processing": "ready_for_pickup",
		"shipped":    "in_transit",
		"delivered":  "delivered",
		"cancelled":  "returned",
	}[status]
	if !ok {
		return
	}
	syncWorkflowState(ctx, engine, constants.WorkflowEntityShipment, shipmentID, code, "status_update", nil)
}

// productStatusToEvent maps legacy product status changes to workflow events (admin actions).
var productStatusToEvent = map[string]string{
	constants.ProductStatusActive: "publish",
}

// productStatusToStateCode maps legacy product status strings to workflow state codes.
var productStatusToStateCode = map[string]string{
	"draft":                            "draft",
	constants.ProductStatusActive:      "published",
	constants.ProductStatusInactive:    "discontinued",
	constants.ProductStatusArchived:    "archived",
}

// applyProductWorkflow tries a validated transition first, then falls back to SetState.
func applyProductWorkflow(
	ctx context.Context,
	engine *workflow.Engine,
	productID uint,
	status, actorRole string,
	actorID *uint,
) bool {
	if engine == nil {
		return false
	}

	if event, ok := productStatusToEvent[status]; ok {
		if err := applyWorkflowEvent(ctx, engine, workflow.TransitionRequest{
			WorkflowKey: constants.WorkflowEntityProduct,
			EntityID:    productID,
			Event:       event,
			ActorID:     actorID,
			ActorRole:   actorRole,
		}); err == nil {
			return true
		}
	}

	code, ok := productStatusToStateCode[status]
	if !ok {
		return false
	}
	syncWorkflowState(ctx, engine, constants.WorkflowEntityProduct, productID, code, "status_update", actorID)
	return true
}

package workflow

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	infraworkflow "github.com/alireza-akbarzadeh/luxe/internal/infrastructure/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

// ApplyEvent runs a validated transition. Returns the engine error when rejected; nil when engine is unset.
func ApplyEvent(ctx context.Context, engine *infraworkflow.Engine, req infraworkflow.TransitionRequest) error {
	if engine == nil {
		return nil
	}
	_, err := engine.Transition(ctx, req)
	return err
}

// SyncState forces an entity to a target state (audit-only, no guards/hooks).
func SyncState(ctx context.Context, engine *infraworkflow.Engine, workflowKey string, entityID uint, stateCode, event string, actorID *uint) {
	if engine == nil {
		return
	}
	if _, err := engine.SetState(ctx, infraworkflow.SetStateRequest{
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

// SyncUserState updates the user account workflow pointer (best-effort).
func SyncUserState(ctx context.Context, engine *infraworkflow.Engine, userID uint, stateCode, event string, actorID *uint) {
	SyncState(ctx, engine, constants.WorkflowEntityUser, userID, stateCode, event, actorID)
}

var orderStatusToEvent = map[string]string{
	constants.OrderStatusPaid:      "payment_succeeded",
	"processing":                   "start_processing",
	constants.OrderStatusShipped:   "ship",
	constants.OrderStatusDelivered: "deliver",
	"completed":                    "complete",
	constants.OrderStatusCancelled: "cancel",
	constants.OrderStatusRefunded:  "refund",
}

// ApplyOrderWorkflow tries a validated transition first, then falls back to SetState.
func ApplyOrderWorkflow(
	ctx context.Context,
	engine *infraworkflow.Engine,
	orderID uint,
	status, actorRole string,
	actorID *uint,
) bool {
	if engine == nil {
		return false
	}

	if event, ok := orderStatusToEvent[status]; ok {
		if err := ApplyEvent(ctx, engine, infraworkflow.TransitionRequest{
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

	_, err := engine.SetState(ctx, infraworkflow.SetStateRequest{
		WorkflowKey:     constants.WorkflowEntityOrder,
		EntityID:        orderID,
		TargetStateCode: code,
		Event:           "admin_set",
		ActorID:         actorID,
	})
	return err == nil
}

var shipmentStatusToEvent = map[string]string{
	"processing": "ready",
	"shipped":    "depart",
	"delivered":  "deliver",
}

// ApplyShipmentWorkflow tries transition then SetState for a legacy shipment status.
func ApplyShipmentWorkflow(ctx context.Context, engine *infraworkflow.Engine, shipmentID uint, status string) {
	if engine == nil {
		return
	}
	if event, ok := shipmentStatusToEvent[status]; ok {
		if err := ApplyEvent(ctx, engine, infraworkflow.TransitionRequest{
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
	SyncState(ctx, engine, constants.WorkflowEntityShipment, shipmentID, code, "status_update", nil)
}

var productStatusToEvent = map[string]string{
	constants.ProductStatusActive: "publish",
}

var productStatusToStateCode = map[string]string{
	"draft":                         "draft",
	constants.ProductStatusActive:   "published",
	constants.ProductStatusInactive: "discontinued",
	constants.ProductStatusArchived: "archived",
}

// ApplyProductWorkflow tries a validated transition first, then falls back to SetState.
func ApplyProductWorkflow(
	ctx context.Context,
	engine *infraworkflow.Engine,
	productID uint,
	status, actorRole string,
	actorID *uint,
) bool {
	if engine == nil {
		return false
	}

	if event, ok := productStatusToEvent[status]; ok {
		if err := ApplyEvent(ctx, engine, infraworkflow.TransitionRequest{
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
	SyncState(ctx, engine, constants.WorkflowEntityProduct, productID, code, "status_update", actorID)
	return true
}

// ApplyCategoryWorkflow syncs category is_active into the workflow engine (best-effort).
func ApplyCategoryWorkflow(
	ctx context.Context,
	engine *infraworkflow.Engine,
	categoryID uint,
	isActive bool,
	actorID *uint,
) bool {
	if engine == nil {
		return false
	}

	if isActive {
		for _, event := range []string{"activate", "reactivate"} {
			if err := ApplyEvent(ctx, engine, infraworkflow.TransitionRequest{
				WorkflowKey: constants.WorkflowEntityCategory,
				EntityID:    categoryID,
				Event:       event,
				ActorID:     actorID,
				ActorRole:   constants.RoleAdmin,
			}); err == nil {
				return true
			}
		}
		SyncState(ctx, engine, constants.WorkflowEntityCategory, categoryID, "active", "status_update", actorID)
		return true
	}

	if err := ApplyEvent(ctx, engine, infraworkflow.TransitionRequest{
		WorkflowKey: constants.WorkflowEntityCategory,
		EntityID:    categoryID,
		Event:       "deactivate",
		ActorID:     actorID,
		ActorRole:   constants.RoleAdmin,
	}); err == nil {
		return true
	}

	SyncState(ctx, engine, constants.WorkflowEntityCategory, categoryID, "inactive", "status_update", actorID)
	return true
}

var brandStatusToStateCode = map[string]string{
	"draft":    "draft",
	"active":   "active",
	"inactive": "inactive",
	"archived": "archived",
}

// ApplyBrandWorkflow syncs legacy brand status into the workflow engine (best-effort).
func ApplyBrandWorkflow(
	ctx context.Context,
	engine *infraworkflow.Engine,
	brandID uint,
	status string,
	actorID *uint,
) bool {
	if engine == nil {
		return false
	}

	eventByStatus := map[string][]string{
		"active":   {"activate", "reactivate"},
		"inactive": {"deactivate"},
		"archived": {"archive"},
	}
	if events, ok := eventByStatus[status]; ok {
		for _, event := range events {
			if err := ApplyEvent(ctx, engine, infraworkflow.TransitionRequest{
				WorkflowKey: constants.WorkflowEntityBrand,
				EntityID:    brandID,
				Event:       event,
				ActorID:     actorID,
				ActorRole:   constants.RoleAdmin,
			}); err == nil {
				return true
			}
		}
	}

	code, ok := brandStatusToStateCode[status]
	if !ok {
		code = "draft"
	}
	SyncState(ctx, engine, constants.WorkflowEntityBrand, brandID, code, "status_update", actorID)
	return true
}

// ApplyCollectionWorkflow syncs legacy collection status into the workflow engine (best-effort).
func ApplyCollectionWorkflow(
	ctx context.Context,
	engine *infraworkflow.Engine,
	collectionID uint,
	status string,
	actorID *uint,
) bool {
	if engine == nil {
		return false
	}

	eventByStatus := map[string][]string{
		"active":   {"activate", "reactivate"},
		"inactive": {"deactivate"},
		"archived": {"archive"},
	}
	if events, ok := eventByStatus[status]; ok {
		for _, event := range events {
			if err := ApplyEvent(ctx, engine, infraworkflow.TransitionRequest{
				WorkflowKey: constants.WorkflowEntityCollection,
				EntityID:    collectionID,
				Event:       event,
				ActorID:     actorID,
				ActorRole:   constants.RoleAdmin,
			}); err == nil {
				return true
			}
		}
	}

	code, ok := brandStatusToStateCode[status]
	if !ok {
		code = "draft"
	}
	SyncState(ctx, engine, constants.WorkflowEntityCollection, collectionID, code, "status_update", actorID)
	return true
}

// ApplyCouponWorkflow syncs coupon is_active into the workflow engine (best-effort).
func ApplyCouponWorkflow(
	ctx context.Context,
	engine *infraworkflow.Engine,
	couponID uint,
	isActive bool,
	actorID *uint,
) bool {
	if engine == nil {
		return false
	}

	if isActive {
		for _, event := range []string{"activate", "resume"} {
			if err := ApplyEvent(ctx, engine, infraworkflow.TransitionRequest{
				WorkflowKey: constants.WorkflowEntityCoupon,
				EntityID:    couponID,
				Event:       event,
				ActorID:     actorID,
				ActorRole:   constants.RoleAdmin,
			}); err == nil {
				return true
			}
		}
		SyncState(ctx, engine, constants.WorkflowEntityCoupon, couponID, "active", "status_update", actorID)
		return true
	}

	for _, event := range []string{"pause"} {
		if err := ApplyEvent(ctx, engine, infraworkflow.TransitionRequest{
			WorkflowKey: constants.WorkflowEntityCoupon,
			EntityID:    couponID,
			Event:       event,
			ActorID:     actorID,
			ActorRole:   constants.RoleAdmin,
		}); err == nil {
			return true
		}
	}
	SyncState(ctx, engine, constants.WorkflowEntityCoupon, couponID, "paused", "status_update", actorID)
	return true
}

// ApplyCouponExhausted moves a coupon to the exhausted workflow state when usage limit is reached.
func ApplyCouponExhausted(ctx context.Context, engine *infraworkflow.Engine, couponID uint) {
	if engine == nil {
		return
	}
	if err := ApplyEvent(ctx, engine, infraworkflow.TransitionRequest{
		WorkflowKey: constants.WorkflowEntityCoupon,
		EntityID:    couponID,
		Event:       "mark_exhausted",
		ActorRole:   constants.RoleAdmin,
	}); err != nil {
		SyncState(ctx, engine, constants.WorkflowEntityCoupon, couponID, "exhausted", "usage_limit_reached", nil)
	}
}

var blogPostStatusToStateCode = map[string]string{
	"draft":      "draft",
	"in_review":  "in_review",
	"scheduled":  "scheduled",
	"published":  "published",
	"archived":   "archived",
}

// ApplyBlogPostWorkflow syncs blog post status into the workflow engine (best-effort).
func ApplyBlogPostWorkflow(
	ctx context.Context,
	engine *infraworkflow.Engine,
	postID uint,
	status string,
	actorID *uint,
) bool {
	if engine == nil {
		return false
	}

	eventByStatus := map[string][]string{
		"in_review": {"submit_review"},
		"scheduled": {"approve_schedule", "schedule"},
		"published": {"publish", "release"},
		"archived":  {"archive"},
		"draft":     {"unpublish", "restore"},
	}
	if events, ok := eventByStatus[status]; ok {
		for _, event := range events {
			if err := ApplyEvent(ctx, engine, infraworkflow.TransitionRequest{
				WorkflowKey: constants.WorkflowEntityBlogPost,
				EntityID:    postID,
				Event:       event,
				ActorID:     actorID,
				ActorRole:   constants.RoleAdmin,
			}); err == nil {
				return true
			}
		}
	}

	code, ok := blogPostStatusToStateCode[status]
	if !ok {
		code = "draft"
	}
	SyncState(ctx, engine, constants.WorkflowEntityBlogPost, postID, code, "status_update", actorID)
	return true
}

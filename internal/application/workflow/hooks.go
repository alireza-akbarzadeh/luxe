package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/asynq"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	infraworkflow "github.com/alireza-akbarzadeh/luxe/internal/infrastructure/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

// Notifier sends in-app notifications (implemented by notification service facade).
type Notifier interface {
	CreateNotification(userID uint, notificationType, title, message string, data interface{}) error
}

// WalletRefunder credits customer wallet on refund hooks.
type WalletRefunder interface {
	Refund(ctx context.Context, userID uint, amount float64, orderID uint) error
}

// InventoryRestocker restores stock when a return workflow completes.
type InventoryRestocker interface {
	RestockForReturn(ctx context.Context, returnID uint) error
}

// HookDeps groups dependencies for registering workflow guards and hooks.
type HookDeps struct {
	Engine   *infraworkflow.Engine
	Repo     *postgres.WorkflowHooksRepository
	Notify   Notifier
	Wallet   WalletRefunder
	JobQueue asynq.JobQueue
}

// RegisterGuardsAndHooks wires guard and hook functions referenced by seeded workflow transitions.
func RegisterGuardsAndHooks(deps HookDeps) {
	engine := deps.Engine
	repo := deps.Repo
	notification := deps.Notify
	wallet := deps.Wallet
	jobQueue := deps.JobQueue

	if engine == nil || repo == nil {
		return
	}

	engine.RegisterGuard("product_has_price", func(ctx context.Context, productID uint) error {
		price, err := repo.GetProductPrice(ctx, productID)
		if err != nil {
			return utils.ErrInternal(err)
		}
		if price <= 0 {
			return utils.ErrBadRequest("product must have a price greater than zero before review")
		}
		return nil
	})

	engine.RegisterGuard("order_payment_succeeded", func(ctx context.Context, orderID uint) error {
		status, err := repo.GetLatestPaymentStatusForOrder(ctx, orderID)
		if err != nil {
			return utils.ErrInternal(err)
		}
		if status == constants.PaymentStatusFailed {
			return utils.ErrBadRequest("cannot mark order paid: payment failed")
		}
		return nil
	})

	engine.RegisterGuard("order_cancellable", func(ctx context.Context, orderID uint) error {
		status, err := repo.GetOrderStatus(ctx, orderID)
		if err != nil {
			return utils.ErrInternal(err)
		}
		blocked := map[string]bool{
			constants.OrderStatusDelivered: true,
			"completed":                    true,
			constants.OrderStatusRefunded:  true,
			constants.OrderStatusCancelled: true,
		}
		if blocked[status] {
			return utils.ErrBadRequest("order cannot be cancelled from its current state")
		}
		return nil
	})

	engine.RegisterHook("product_published", func(ctx context.Context, productID uint, _ map[string]interface{}) error {
		return repo.SetProductPublishedAt(ctx, productID, time.Now())
	})

	engine.RegisterHook("blog_post_published", func(ctx context.Context, postID uint, _ map[string]interface{}) error {
		return repo.SetBlogPostPublishedAt(ctx, postID, time.Now())
	})

	engine.RegisterHook("order_paid", orderNotifyHook(repo, notification, jobQueue,
		"order_paid", "Payment Confirmed", "Payment for order #%s has been confirmed."))

	engine.RegisterHook("order_shipped", orderNotifyHook(repo, notification, jobQueue,
		"order_shipped", "Order Shipped", "Your order #%s has been shipped."))

	engine.RegisterHook("order_refunded", orderNotifyHook(repo, notification, jobQueue,
		"order_refunded", "Order Refunded", "Your order #%s has been refunded."))

	engine.RegisterHook("order_cancelled", orderNotifyHook(repo, notification, jobQueue,
		"order_cancelled", "Order Cancelled", "Your order #%s has been cancelled."))

	engine.RegisterHook("shipment_delivered", func(ctx context.Context, shipmentID uint, _ map[string]interface{}) error {
		return repo.SetShipmentDeliveredAt(ctx, shipmentID, time.Now())
	})

	if wallet != nil {
		engine.RegisterHook("return_refunded", func(ctx context.Context, returnID uint, _ map[string]interface{}) error {
			row, err := repo.GetReturnRefundInfo(ctx, returnID)
			if err != nil {
				return utils.ErrInternal(err)
			}
			if row.RefundAmount <= 0 {
				return nil
			}
			return wallet.Refund(ctx, row.UserID, row.RefundAmount, row.OrderID)
		})
	}
}

// RegisterInventoryHooks wires inventory side-effects into return workflow transitions.
func RegisterInventoryHooks(engine *infraworkflow.Engine, inventory InventoryRestocker) {
	if engine == nil || inventory == nil {
		return
	}
	engine.RegisterHook("return_restock_inventory", func(ctx context.Context, returnID uint, _ map[string]interface{}) error {
		return inventory.RestockForReturn(ctx, returnID)
	})
}

func orderNotifyHook(
	repo *postgres.WorkflowHooksRepository,
	notification Notifier,
	jobQueue asynq.JobQueue,
	notifType, title, bodyTemplate string,
) infraworkflow.HookFunc {
	return func(ctx context.Context, orderID uint, _ map[string]interface{}) error {
		row, err := repo.GetOrderNotifyRow(ctx, orderID)
		if err != nil {
			return utils.ErrInternal(err)
		}
		message := fmt.Sprintf(bodyTemplate, row.OrderNumber)

		if notification != nil {
			_ = notification.CreateNotification(row.UserID, notifType, title, message, map[string]interface{}{
				"order_id":     orderID,
				"order_number": row.OrderNumber,
			})
		}

		if jobQueue != nil {
			email, err := repo.GetUserEmail(ctx, row.UserID)
			if err == nil && email != "" {
				_ = jobQueue.EnqueueSendEmail(ctx, email, title, message)
			}
		}
		return nil
	}
}

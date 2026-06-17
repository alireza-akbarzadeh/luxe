package services

import (
	"context"
	"fmt"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/services/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/tasks"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"gorm.io/gorm"
)

// RegisterWorkflowGuardsAndHooks wires the guard and hook functions referenced by
// the seeded workflow transitions. New workflows can reference these keys (or new
// ones registered here) without modifying the engine itself.
func RegisterWorkflowGuardsAndHooks(
	engine *workflow.Engine,
	db *gorm.DB,
	notification NotificationServiceInterface,
	wallet WalletServiceInterface,
	jobQueue tasks.JobQueue,
) {
	// ─── Guards ─────────────────────────────────────────────────────────────

	// A product must have a price before it can enter review.
	engine.RegisterGuard("product_has_price", func(ctx context.Context, productID uint) error {
		var price float64
		if err := db.WithContext(ctx).Table("products").
			Select("price").Where("id = ?", productID).Scan(&price).Error; err != nil {
			return utils.ErrInternal(err)
		}
		if price <= 0 {
			return utils.ErrBadRequest("product must have a price greater than zero before review")
		}
		return nil
	})

	// An order may move to paid only if its payment did not explicitly fail.
	engine.RegisterGuard("order_payment_succeeded", func(ctx context.Context, orderID uint) error {
		var status string
		err := db.WithContext(ctx).Table("payments").
			Select("status").Where("order_id = ?", orderID).
			Order("id DESC").Limit(1).Scan(&status).Error
		if err != nil {
			return utils.ErrInternal(err)
		}
		if status == constants.PaymentStatusFailed {
			return utils.ErrBadRequest("cannot mark order paid: payment failed")
		}
		return nil
	})

	// An order can be cancelled unless it is already in a terminal/fulfilled state.
	engine.RegisterGuard("order_cancellable", func(ctx context.Context, orderID uint) error {
		var status string
		if err := db.WithContext(ctx).Table("orders").
			Select("status").Where("id = ?", orderID).Scan(&status).Error; err != nil {
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

	// ─── Hooks ──────────────────────────────────────────────────────────────

	engine.RegisterHook("product_published", func(ctx context.Context, productID uint, _ map[string]interface{}) error {
		now := time.Now()
		return db.WithContext(ctx).Table("products").
			Where("id = ?", productID).Update("published_at", now).Error
	})

	engine.RegisterHook("order_paid", orderNotifyHook(db, notification, jobQueue,
		"order_paid", "Payment Confirmed", "Payment for order #%s has been confirmed."))

	engine.RegisterHook("order_shipped", orderNotifyHook(db, notification, jobQueue,
		"order_shipped", "Order Shipped", "Your order #%s has been shipped."))

	engine.RegisterHook("order_refunded", orderNotifyHook(db, notification, jobQueue,
		"order_refunded", "Order Refunded", "Your order #%s has been refunded."))

	engine.RegisterHook("order_cancelled", orderNotifyHook(db, notification, jobQueue,
		"order_cancelled", "Order Cancelled", "Your order #%s has been cancelled."))

	engine.RegisterHook("shipment_delivered", func(ctx context.Context, shipmentID uint, _ map[string]interface{}) error {
		now := time.Now()
		return db.WithContext(ctx).Table("shipments").
			Where("id = ?", shipmentID).Update("delivered_at", now).Error
	})

	// On completed refund, credit the customer's wallet.
	engine.RegisterHook("return_refunded", func(ctx context.Context, returnID uint, _ map[string]interface{}) error {
		var row struct {
			OrderID      uint
			UserID       uint
			RefundAmount float64
		}
		if err := db.WithContext(ctx).Table("returns").
			Select("order_id, user_id, refund_amount").
			Where("id = ?", returnID).Scan(&row).Error; err != nil {
			return utils.ErrInternal(err)
		}
		if row.RefundAmount <= 0 {
			return nil
		}
		return wallet.Refund(row.UserID, row.RefundAmount, row.OrderID)
	})
}

// orderNotifyHook builds a hook that loads the order's user + number and sends an
// in-app notification plus an async email.
func orderNotifyHook(
	db *gorm.DB,
	notification NotificationServiceInterface,
	jobQueue tasks.JobQueue,
	notifType, title, bodyTemplate string,
) workflow.HookFunc {
	return func(ctx context.Context, orderID uint, _ map[string]interface{}) error {
		var row struct {
			UserID      uint
			OrderNumber string
		}
		if err := db.WithContext(ctx).Table("orders").
			Select("user_id, order_number").
			Where("id = ?", orderID).Scan(&row).Error; err != nil {
			return utils.ErrInternal(err)
		}
		message := fmt.Sprintf(bodyTemplate, row.OrderNumber)

		_ = notification.CreateNotification(row.UserID, notifType, title, message, map[string]interface{}{
			"order_id":     orderID,
			"order_number": row.OrderNumber,
		})

		if jobQueue != nil {
			var email string
			if err := db.WithContext(ctx).Table("users").
				Select("email").Where("id = ?", row.UserID).Scan(&email).Error; err == nil && email != "" {
				_ = jobQueue.EnqueueSendEmail(ctx, email, title, message)
			}
		}
		return nil
	}
}

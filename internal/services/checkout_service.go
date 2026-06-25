package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/asynq"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	appcheckout "github.com/alireza-akbarzadeh/luxe/internal/application/checkout"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/alireza-akbarzadeh/luxe/internal/websocket"
	domaincart "github.com/alireza-akbarzadeh/luxe/internal/domain/cart"
	domaincheckout "github.com/alireza-akbarzadeh/luxe/internal/domain/checkout"
	domainorder "github.com/alireza-akbarzadeh/luxe/internal/domain/order"
	"gorm.io/gorm"
)

const defaultShippingProvider = "standard"

type CheckoutServiceInterface interface {
	Checkout(ctx context.Context, userID uint, req dto.CheckoutRequest) (*dto.CheckoutResult, error)
	CompletePaidOrder(ctx context.Context, orderID uint) error
	ProcessOrder(ctx context.Context, orderID uint, cardInfo dto.CardInfo) error
	// CancelOrder cancels an order owned by userID, restores stock, and refunds wallet payments.
	CancelOrder(ctx context.Context, orderID, userID uint) error
}

type checkoutService struct {
	db                  *gorm.DB
	checkoutRepo        *postgres.CheckoutRepository
	cartRepo            *postgres.CartRepository
	notificationService NotificationServiceInterface
	couponService       CouponServiceInterface
	paymentService      PaymentServiceInterface
	shipmentService     ShipmentServiceInterface
	walletService       WalletServiceInterface
	invoiceService      InvoiceServiceInterface
	workerPool          asynq.JobQueue
	hub                 *websocket.Hub
	salesFeed           *SalesFeedService
	engine              *workflow.Engine
	inventoryService    InventoryServiceInterface
	stripeEnabled       bool
	cartDomain          *domaincart.Service
	checkoutDomain      *domaincheckout.Service
	orderDomain         *domainorder.Service
}

func NewCheckoutService(
	db *gorm.DB,
	notificationService NotificationServiceInterface,
	couponService CouponServiceInterface,
	paymentService PaymentServiceInterface,
	shipmentService ShipmentServiceInterface,
	walletService WalletServiceInterface,
	invoiceService InvoiceServiceInterface,
	workerPool asynq.JobQueue,
	hub *websocket.Hub,
	salesFeed *SalesFeedService,
	engine *workflow.Engine,
	inventoryService InventoryServiceInterface,
	stripeEnabled bool,
) CheckoutServiceInterface {
	return &checkoutService{
		db:                  db,
		checkoutRepo:        postgres.NewCheckoutRepository(db),
		cartRepo:            postgres.NewCartRepository(db),
		notificationService: notificationService,
		couponService:       couponService,
		paymentService:      paymentService,
		shipmentService:     shipmentService,
		walletService:       walletService,
		invoiceService:      invoiceService,
		workerPool:          workerPool,
		hub:                 hub,
		salesFeed:           salesFeed,
		engine:              engine,
		inventoryService:    inventoryService,
		stripeEnabled:       stripeEnabled,
		cartDomain:          domaincart.NewService(),
		checkoutDomain:      domaincheckout.NewService(),
		orderDomain:         domainorder.NewService(),
	}
}

// setOrderState moves an order to a workflow state via the engine (best-effort:
// failures are logged but never block the core checkout/payment flow).
func (s *checkoutService) setOrderState(ctx context.Context, orderID uint, stateCode, event string) {
	s.setOrderStateActor(ctx, orderID, stateCode, event, nil)
}

func (s *checkoutService) setOrderStateActor(ctx context.Context, orderID uint, stateCode, event string, actorID *uint) {
	if s.engine == nil {
		return
	}
	if stateCode == "paid" {
		if applyOrderWorkflow(ctx, s.engine, orderID, constants.OrderStatusPaid, constants.RoleUser, actorID) {
			return
		}
	}
	syncWorkflowState(ctx, s.engine, constants.WorkflowEntityOrder, orderID, stateCode, event, actorID)
}

// Checkout converts the user's active cart into an order.
func (s *checkoutService) Checkout(ctx context.Context, userID uint, req dto.CheckoutRequest) (*dto.CheckoutResult, error) {
	req.NormalizePaymentMethod(s.stripeEnabled)

	cart, err := appcheckout.LoadActiveCart(ctx, s.checkoutRepo, userID)
	if err != nil {
		return nil, err
	}
	if err := appcheckout.ValidateCartForCheckout(s.cartDomain, userID, cart); err != nil {
		return nil, utils.ErrBadRequest(err.Error())
	}
	if err := appcheckout.ValidateCheckoutInput(s.checkoutDomain, domaincheckout.CheckoutInput{
		UserID:         userID,
		CartTotalCents: int64(appcheckout.CartSubtotal(cart.Items) * 100),
		Currency:       "USD",
		PaymentMethod:  req.PaymentMethod,
	}); err != nil {
		return nil, utils.ErrBadRequest(err.Error())
	}
	address, err := s.checkoutRepo.ResolveOrCreateAddress(ctx, userID, req)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	carrier := s.getCarrier(req.ShippingProviderID)

	var order *models.Order
	var stockChanges []deltaResult
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		subtotal := appcheckout.CartSubtotal(cart.Items)
		discount, couponID, err := s.applyCoupon(tx, userID, req.CouponCode, subtotal)
		if err != nil {
			return err
		}
		totalAmount := subtotal - discount
		if totalAmount < 0 {
			totalAmount = 0
		}

		order = appcheckout.BuildOrderModel(userID, totalAmount, address.ID, generateOrderNumber(userID))
		if err := s.checkoutRepo.CreateOrder(tx, order); err != nil {
			return utils.ErrInternal(err)
		}

		stockChanges, err = s.reserveCartStock(ctx, tx, order.ID, cart.Items)
		if err != nil {
			return err
		}

		if err := s.checkoutRepo.CreateOrderItems(tx, order.ID, cart.Items); err != nil {
			return utils.ErrInternal(err)
		}

		if err := s.createPayment(tx, order.ID, userID, totalAmount, req.PaymentMethod, "USD"); err != nil {
			return err
		}

		if err := s.createShipment(tx, order.ID, userID, carrier, req); err != nil {
			return err
		}

		if couponID != nil {
			if err := s.couponService.ApplyCoupon(tx, userID, order.ID, req.CouponCode, subtotal); err != nil {
				return err
			}
		}

		return s.cartRepo.MarkConvertedTx(tx, cart.ID)
	})
	if err != nil {
		return nil, err
	}

	for _, change := range stockChanges {
		if change.QuantityBefore != change.QuantityAfter {
			s.inventoryService.RunStockSideEffects(ctx, change.Product, change.QuantityBefore, change.QuantityAfter)
		}
	}

	if err := s.checkoutRepo.PreloadOrderDetails(ctx, order); err != nil {
		return nil, utils.ErrInternal(err)
	}
	s.sendOrderCreatedNotification(userID, order)
	s.setOrderState(ctx, order.ID, "pending_payment", "order_created")

	result := &dto.CheckoutResult{Order: order}

	if req.PaymentMethod == "stripe" {
		payment, err := s.checkoutRepo.FindPaymentByOrderID(ctx, order.ID)
		if err != nil {
			return nil, utils.ErrInternal(err)
		}
		checkoutURL, sessionID, err := s.paymentService.CreateStripeCheckoutSession(order, payment, req.Email)
		if err != nil {
			return nil, err
		}
		result.CheckoutURL = checkoutURL
		result.StripeSessionID = sessionID
		return result, nil
	}

	s.enqueueFulfillmentJob(ctx, order.ID, req)
	return result, nil
}

// CompletePaidOrder finalizes an order after external payment confirmation (Stripe webhook).
func (s *checkoutService) CompletePaidOrder(ctx context.Context, orderID uint) error {
	order, err := s.checkoutRepo.FindOrderWithPayment(ctx, orderID)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}

	if order.Status == constants.OrderStatusPaid {
		s.ensureInvoice(ctx, orderID)
		return nil
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return s.checkoutRepo.UpdateOrderStatusTx(tx, orderID, constants.OrderStatusPaid)
	})
	if err != nil {
		return err
	}

	s.setOrderState(ctx, orderID, "paid", "payment_succeeded")

	s.ensureInvoice(ctx, orderID)

	txnID := ""
	if order.Payment != nil {
		txnID = order.Payment.TransactionID
	}
	s.broadcastOrderUpdate(order.ID, order.UserID, "payment_succeeded", map[string]interface{}{
		"title":          "Payment Confirmed",
		"message":        fmt.Sprintf("Order %s paid successfully", order.OrderNumber),
		"order_id":       order.ID,
		"order_number":   order.OrderNumber,
		"total_amount":   order.TotalAmount,
		"transaction_id": txnID,
		"status":         constants.OrderStatusPaid,
	})

	return s.processShipment(ctx, orderID)
}

// ProcessOrder is the background job handler that orchestrates payment and shipment.
func (s *checkoutService) ProcessOrder(ctx context.Context, orderID uint, cardInfo dto.CardInfo) error {
	// Use a transaction for the payment + order status update
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Load order with payment
		var order models.Order
		if err := tx.Preload("Payment").First(&order, orderID).Error; err != nil {
			return fmt.Errorf("order not found: %w", err)
		}
		if order.Payment == nil {
			return fmt.Errorf("payment record missing for order %d", orderID)
		}

		// 2. Process payment — wallet deducts balance; mock/card goes through gateway.
		if order.Payment.Method == "wallet" {
			if err := s.walletService.DeductForOrder(order.UserID, order.TotalAmount, order.ID); err != nil {
				tx.Model(&order).Update("status", "payment_failed")
				tx.Model(&models.Shipment{}).Where("order_id = ?", order.ID).Update("status", "cancelled")
				s.broadcastOrderUpdate(order.ID, order.UserID, "payment_failed", map[string]interface{}{
					"title":    "Payment Failed",
					"message":  fmt.Sprintf("Wallet payment failed: %s", err.Error()),
					"order_id": order.ID,
					"status":   "payment_failed",
				})
				return err
			}
			tx.Model(&models.Payment{}).Where("id = ?", order.Payment.ID).Updates(map[string]interface{}{
				"status":         constants.PaymentStatusSucceeded,
				"transaction_id": fmt.Sprintf("wallet_%d_%d", order.UserID, order.ID),
			})
		} else if err := s.paymentService.ProcessPayment(tx, order.Payment.ID, cardInfo); err != nil {
			tx.Model(&order).Update("status", "payment_failed")
			tx.Model(&models.Shipment{}).Where("order_id = ?", order.ID).Update("status", "cancelled")
			s.broadcastOrderUpdate(order.ID, order.UserID, "payment_failed", map[string]interface{}{
				"title":    "Payment Failed",
				"message":  fmt.Sprintf("Payment failed: %s", err.Error()),
				"order_id": order.ID,
				"status":   "payment_failed",
			})
			return err
		}

		// Payment succeeded – update order to pay
		tx.Model(&order).Update("status", constants.OrderStatusPaid)

		// Broadcast success
		s.broadcastOrderUpdate(order.ID, order.UserID, "payment_succeeded", map[string]interface{}{
			"title":          "Payment Confirmed",
			"message":        fmt.Sprintf("Order %s paid successfully", order.OrderNumber),
			"order_id":       order.ID,
			"order_number":   order.OrderNumber,
			"total_amount":   order.TotalAmount,
			"transaction_id": order.Payment.TransactionID,
			"status":         constants.OrderStatusPaid,
		})
		return nil
	})
	if err != nil {
		return err
	}

	s.setOrderState(ctx, orderID, "paid", "payment_succeeded")

	s.ensureInvoice(ctx, orderID)

	// Transaction committed – now handle shipment (outside transaction for performance)
	return s.processShipment(ctx, orderID)
}

func (s *checkoutService) broadcastOrderUpdate(orderID, userID uint, eventType string, data map[string]interface{}) {
	msg := websocket.Message{
		Type:      eventType,
		UserID:    userID,
		Data:      data,
		Timestamp: time.Now(),
	}
	roomID := fmt.Sprintf("order_%d", orderID)
	s.hub.BroadcastToRoom(roomID, msg)

	go func() {
		_ = s.notificationService.CreateNotification(
			userID,
			eventType,
			data["title"].(string),
			data["message"].(string),
			data,
		)
	}()

	if s.salesFeed != nil {
		switch eventType {
		case "payment_succeeded":
			totalAmount, _ := data["total_amount"].(float64)
			orderNumber, _ := data["order_number"].(string)
			title, _ := data["title"].(string)
			message, _ := data["message"].(string)
			s.salesFeed.PublishOrderEvent("payment", title, message, totalAmount)
			s.salesFeed.PublishRevenueSnapshot(totalAmount, 1)
			if orderNumber != "" {
				s.salesFeed.PublishOrderEvent(
					"new_order",
					fmt.Sprintf("New order %s", orderNumber),
					message,
					totalAmount,
				)
			}
		case "payment_failed":
			title, _ := data["title"].(string)
			message, _ := data["message"].(string)
			s.salesFeed.PublishOrderEvent("cancellation", title, message, 0)
		}
	}
}

// processShipment handles the shipping steps (called after payment success)
func (s *checkoutService) processShipment(ctx context.Context, orderID uint) error {
	shipment, err := s.checkoutRepo.FindShipmentByOrderID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("shipment not found: %w", err)
	}

	oldStatus := shipment.Status
	shipment.Status = "processing"
	if err := s.checkoutRepo.SaveShipment(ctx, shipment); err != nil {
		return err
	}

	s.broadcastOrderUpdate(orderID, shipment.UserID, "shipment_processing", map[string]interface{}{
		"title":       "Preparing Shipment",
		"message":     "Your order is being prepared for shipping.",
		"order_id":    orderID,
		"shipment_id": shipment.ID,
		"old_status":  oldStatus,
		"new_status":  "processing",
	})

	// 2. Simulate carrier API – generate tracking & mark shipped
	time.Sleep(2 * time.Second) // simulate carrier delay

	trackingNumber := fmt.Sprintf("TRK-%d-%d", orderID, time.Now().Unix())
	now := time.Now()
	if err := s.checkoutRepo.UpdateShipmentFields(ctx, orderID, map[string]interface{}{
		"status":          constants.ShipmentStatusShipped,
		"tracking_number": trackingNumber,
		"shipped_at":      now,
	}); err != nil {
		return err
	}

	s.broadcastOrderUpdate(orderID, shipment.UserID, "shipment_shipped", map[string]interface{}{
		"title":           "Package Shipped",
		"message":         fmt.Sprintf("Your package is on the way! Tracking: %s", trackingNumber),
		"order_id":        orderID,
		"shipment_id":     shipment.ID,
		"tracking_number": trackingNumber,
		"carrier":         shipment.Carrier,
		"shipped_at":      now,
		"status":          constants.ShipmentStatusShipped,
	})

	return nil
}

// getCarrier resolves the shipping carrier name from the request or falls back to the first active provider.
func (s *checkoutService) getCarrier(shippingProviderID *uint) string {
	if shippingProviderID != nil {
		provider, err := s.shipmentService.GetShippingProviderByID(*shippingProviderID)
		if err == nil {
			return provider.Name
		}
	}
	providers, _ := s.shipmentService.GetShippingProviders()
	if len(providers) > 0 {
		return providers[0].Name
	}
	return defaultShippingProvider
}

// reserveCartStock locks product rows, validates stock, and decrements inventory.
func (s *checkoutService) reserveCartStock(ctx context.Context, tx *gorm.DB, orderID uint, cartItems []models.CartItem) ([]deltaResult, error) {
	var changes []deltaResult
	for _, item := range cartItems {
		var product models.Product
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&product, item.ProductID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, utils.ErrBadRequest("product not found")
			}
			return nil, utils.ErrInternal(err)
		}
		if !appcheckout.ProductStockAvailable(product, item.Quantity) {
			return nil, utils.ErrBadRequest(
				fmt.Sprintf("insufficient stock for product: %s", product.Name))
		}
		if appcheckout.ShouldDecrementProductStock(product) {
			if s.inventoryService != nil {
				change, err := s.inventoryService.DecrementForSale(ctx, tx, orderID, item.ProductID, item.Quantity)
				if err != nil {
					return nil, err
				}
				changes = append(changes, change)
				continue
			}
			product.Stock -= item.Quantity
			if err := tx.Save(&product).Error; err != nil {
				return nil, utils.ErrInternal(err)
			}
		}
	}
	return changes, nil
}

// applyCoupon validates a coupon code (if provided) and returns the discount amount and coupon ID.
func (s *checkoutService) applyCoupon(tx *gorm.DB, userID uint, code string, subtotal float64) (float64, *uint, error) {
	// No coupon code → no discount
	if code == "" {
		return 0, nil, nil
	}

	// Validate the coupon (ideally pass tx for atomicity)
	coupon, discount, err := s.couponService.ValidateCoupon(code, userID, subtotal)
	if err != nil {
		return 0, nil, err
	}

	return discount, &coupon.ID, nil
}

// createPayment uses the PaymentService to insert a pending payment record (within tx).
func (s *checkoutService) createPayment(tx *gorm.DB, orderID, userID uint, amount float64, method, currency string) error {
	req := dto.PaymentRequest{
		OrderID:  orderID,
		UserID:   userID,
		Amount:   amount,
		Method:   method,
		Currency: currency,
	}
	_, err := s.paymentService.CreatePayment(tx, req)
	return err
}

// createShipment uses the ShipmentService to insert a pending shipment record (within tx).
func (s *checkoutService) createShipment(tx *gorm.DB, orderID, userID uint, carrier string, req dto.CheckoutRequest) error {
	shipReq := CreateShipmentRequest{
		OrderID:      orderID,
		Carrier:      carrier,
		AddressLine1: req.AddressLine1,
		AddressLine2: req.AddressLine2,
		City:         req.City,
		State:        req.State,
		PostalCode:   req.Zip,
		Country:      req.Country,
	}
	_, err := s.shipmentService.CreateShipmentRecord(tx, shipReq)
	return err
}

// enqueueFulfillmentJob submits the ProcessOrder job to the background queue.
func (s *checkoutService) enqueueFulfillmentJob(ctx context.Context, orderID uint, req dto.CheckoutRequest) {
	cardInfo := dto.CardInfo{
		CardNumber:  req.CardNumber,
		ExpiryMonth: req.ExpiryMonth,
		ExpiryYear:  req.ExpiryYear,
		CVV:         req.CVV,
	}
	if err := s.workerPool.EnqueueProcessOrder(ctx, orderID, cardInfo); err != nil {
		utils.Log.WithError(err).WithField("order_id", orderID).Error("failed to enqueue order fulfillment job")
	}
}

// sendOrderCreatedNotification sends the initial "order placed" notification asynchronously.
func (s *checkoutService) sendOrderCreatedNotification(userID uint, order *models.Order) {
	go func() {
		_ = s.notificationService.CreateNotification(
			userID,
			"order_created",
			"Order Placed Successfully",
			fmt.Sprintf("Your order #%s has been placed and is being processed.", order.OrderNumber),
			map[string]interface{}{
				"order_id":     order.ID,
				"order_number": order.OrderNumber,
				"status":       order.Status,
				"total_amount": order.TotalAmount,
				"currency":     order.Currency,
			},
		)
	}()
}

// cancellableStatuses are the order statuses a customer may cancel from.
var cancellableStatuses = map[string]bool{
	constants.OrderStatusPending: true,
	constants.OrderStatusPaid:    true,
}

// CancelOrder cancels an order belonging to userID, restores stock, and refunds
// wallet payments. Stripe orders are cancelled without an automatic refund (requires
// manual processing via the Stripe dashboard).
func (s *checkoutService) CancelOrder(ctx context.Context, orderID, userID uint) error {
	order, err := s.checkoutRepo.FindOrderForCancel(ctx, orderID, userID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return utils.ErrNotFound("order not found")
		}
		return utils.ErrInternal(err)
	}

	if err := s.orderDomain.CanCancel(domainorder.Order{
		ID:     order.ID,
		UserID: order.UserID,
		Status: order.Status,
	}); err != nil {
		return utils.ErrBadRequest(err.Error())
	}

	var restores []deltaResult
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range order.Items {
			if s.inventoryService != nil {
				change, err := s.inventoryService.RestoreForOrderCancel(ctx, tx, order.ID, item.ProductID, item.Quantity)
				if err != nil {
					return err
				}
				restores = append(restores, change)
				continue
			}
			if err := tx.Model(&models.Product{}).
				Where("id = ?", item.ProductID).
				UpdateColumn("stock", gorm.Expr("stock + ?", item.Quantity)).Error; err != nil {
				return utils.ErrInternal(err)
			}
		}

		// Refund wallet if it was the payment method and payment succeeded.
		if order.Payment != nil &&
			order.Payment.Method == "wallet" &&
			order.Payment.Status == constants.PaymentStatusSucceeded {
			if err := s.walletService.Refund(order.UserID, order.TotalAmount, order.ID); err != nil {
				return err
			}
			tx.Model(&models.Payment{}).Where("id = ?", order.Payment.ID).
				Update("status", constants.PaymentStatusRefunded)
		}

		// Cancel any pending/processing shipment.
		return s.checkoutRepo.CancelPendingShipmentsTx(tx, order.ID)
	})
	if err != nil {
		return err
	}

	for _, change := range restores {
		if change.QuantityBefore != change.QuantityAfter && s.inventoryService != nil {
			s.inventoryService.RunStockSideEffects(ctx, change.Product, change.QuantityBefore, change.QuantityAfter)
		}
	}

	if err := applyWorkflowEvent(ctx, s.engine, workflow.TransitionRequest{
		WorkflowKey: constants.WorkflowEntityOrder,
		EntityID:    order.ID,
		Event:       "cancel",
		ActorID:     &userID,
		ActorRole:   constants.RoleUser,
	}); err != nil {
		// Fallback when transition rules reject the move (e.g. stale workflow state).
		s.setOrderStateActor(ctx, order.ID, "cancelled", "cancel", &userID)
	}

	return nil
}

func (s *checkoutService) ensureInvoice(ctx context.Context, orderID uint) {
	if s.invoiceService == nil {
		return
	}
	if _, err := s.invoiceService.CreateFromPaidOrder(ctx, orderID); err != nil {
		utils.Log.WithError(err).WithField("order_id", orderID).Warn("failed to create invoice for paid order")
	}
}

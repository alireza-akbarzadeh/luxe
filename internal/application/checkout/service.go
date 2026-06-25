package checkout

import (
	"context"
	"fmt"
	"time"

	appcoupon "github.com/alireza-akbarzadeh/luxe/internal/application/coupon"
	appinventory "github.com/alireza-akbarzadeh/luxe/internal/application/inventory"
	appinvoice "github.com/alireza-akbarzadeh/luxe/internal/application/invoice"
	apppayment "github.com/alireza-akbarzadeh/luxe/internal/application/payment"
	appsalesfeed "github.com/alireza-akbarzadeh/luxe/internal/application/salesfeed"
	appshipment "github.com/alireza-akbarzadeh/luxe/internal/application/shipment"
	appwallet "github.com/alireza-akbarzadeh/luxe/internal/application/wallet"
	appworkflow "github.com/alireza-akbarzadeh/luxe/internal/application/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/domain/cart"
	"github.com/alireza-akbarzadeh/luxe/internal/domain/checkout"
	"github.com/alireza-akbarzadeh/luxe/internal/domain/order"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/asynq"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/alireza-akbarzadeh/luxe/internal/websocket"
	"gorm.io/gorm"
)

const defaultShippingProvider = "standard"

// Notifier sends in-app notifications during checkout flows.
type Notifier interface {
	CreateNotification(userID uint, notificationType, title, message string, data interface{}) error
}

// Service orchestrates cart-to-order checkout, payment, and fulfillment.
type Service struct {
	checkoutRepo     *postgres.CheckoutRepository
	cartRepo         *postgres.CartRepository
	notifier         Notifier
	couponService    *appcoupon.Service
	paymentService   *apppayment.Service
	shipmentService  *appshipment.Service
	walletService    *appwallet.Service
	invoiceCommands  *appinvoice.Commands
	workerPool       asynq.JobQueue
	hub              *websocket.Hub
	salesFeed        *appsalesfeed.Service
	engine           *workflow.Engine
	inventoryService *appinventory.Service
	stripeEnabled    bool
	cartDomain       *cart.Service
	checkoutDomain   *checkout.Service
	orderDomain      *order.Service
}

// NewService wires checkout orchestration dependencies.
func NewService(
	db *gorm.DB,
	notifier Notifier,
	couponService *appcoupon.Service,
	paymentService *apppayment.Service,
	shipmentService *appshipment.Service,
	walletService *appwallet.Service,
	invoiceCommands *appinvoice.Commands,
	workerPool asynq.JobQueue,
	hub *websocket.Hub,
	salesFeed *appsalesfeed.Service,
	engine *workflow.Engine,
	inventoryService *appinventory.Service,
	stripeEnabled bool,
) *Service {
	return &Service{
		checkoutRepo:     postgres.NewCheckoutRepository(db),
		cartRepo:         postgres.NewCartRepository(db),
		notifier:         notifier,
		couponService:    couponService,
		paymentService:   paymentService,
		shipmentService:  shipmentService,
		walletService:    walletService,
		invoiceCommands:  invoiceCommands,
		workerPool:       workerPool,
		hub:              hub,
		salesFeed:        salesFeed,
		engine:           engine,
		inventoryService: inventoryService,
		stripeEnabled:    stripeEnabled,
		cartDomain:       cart.NewService(),
		checkoutDomain:   checkout.NewService(),
		orderDomain:      order.NewService(),
	}
}

// setOrderState moves an order to a workflow state via the engine (best-effort:
// failures are logged but never block the core checkout/payment flow).
func (s *Service) setOrderState(ctx context.Context, orderID uint, stateCode, event string) {
	s.setOrderStateActor(ctx, orderID, stateCode, event, nil)
}

func (s *Service) setOrderStateActor(ctx context.Context, orderID uint, stateCode, event string, actorID *uint) {
	if s.engine == nil {
		return
	}
	if stateCode == "paid" {
		if appworkflow.ApplyOrderWorkflow(ctx, s.engine, orderID, constants.OrderStatusPaid, constants.RoleUser, actorID) {
			return
		}
	}
	appworkflow.SyncState(ctx, s.engine, constants.WorkflowEntityOrder, orderID, stateCode, event, actorID)
}

// Checkout converts the user's active cart into an order.
func (s *Service) Checkout(ctx context.Context, userID uint, req dto.CheckoutRequest) (*dto.CheckoutResult, error) {
	req.NormalizePaymentMethod(s.stripeEnabled)

	cart, err := LoadActiveCart(ctx, s.checkoutRepo, userID)
	if err != nil {
		return nil, err
	}
	if err := ValidateCartForCheckout(s.cartDomain, userID, cart); err != nil {
		return nil, utils.ErrBadRequest(err.Error())
	}
	if err := ValidateCheckoutInput(s.checkoutDomain, checkout.CheckoutInput{
		UserID:         userID,
		CartTotalCents: int64(CartSubtotal(cart.Items) * 100),
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
	var stockChanges []appinventory.DeltaResult
	err = s.checkoutRepo.Transaction(ctx, func(tx *gorm.DB) error {
		subtotal := CartSubtotal(cart.Items)
		discount, couponID, err := s.applyCoupon(ctx, tx, userID, req.CouponCode, subtotal)
		if err != nil {
			return err
		}
		totalAmount := subtotal - discount
		if totalAmount < 0 {
			totalAmount = 0
		}

		order = BuildOrderModel(userID, totalAmount, address.ID, GenerateOrderNumber(userID))
		if err := s.checkoutRepo.CreateOrder(tx, order); err != nil {
			return utils.ErrInternal(err)
		}

		stockChanges, err = ReserveCartStock(ctx, tx, s.checkoutRepo, order.ID, cart.Items, func(ctx context.Context, tx *gorm.DB, orderID, productID uint, qty int) (appinventory.DeltaResult, error) {
			return s.inventoryService.DecrementForSale(ctx, tx, orderID, productID, qty)
		})
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
func (s *Service) CompletePaidOrder(ctx context.Context, orderID uint) error {
	order, err := s.checkoutRepo.FindOrderWithPayment(ctx, orderID)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}

	if order.Status == constants.OrderStatusPaid {
		s.ensureInvoice(ctx, orderID)
		return nil
	}

	err = s.checkoutRepo.Transaction(ctx, func(tx *gorm.DB) error {
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
func (s *Service) ProcessOrder(ctx context.Context, orderID uint, cardInfo dto.CardInfo) error {
	var order *models.Order
	err := s.checkoutRepo.Transaction(ctx, func(tx *gorm.DB) error {
		var err error
		order, err = s.checkoutRepo.FindOrderWithPaymentTx(tx, orderID)
		if err != nil {
			return fmt.Errorf("order not found: %w", err)
		}
		if order.Payment == nil {
			return fmt.Errorf("payment record missing for order %d", orderID)
		}

		if order.Payment.Method == "wallet" {
			if err := s.walletService.DeductForOrder(ctx, order.UserID, order.TotalAmount, order.ID); err != nil {
				if markErr := s.checkoutRepo.MarkOrderPaymentFailedTx(tx, order.ID); markErr != nil {
					return markErr
				}
				s.broadcastOrderUpdate(order.ID, order.UserID, "payment_failed", map[string]interface{}{
					"title":    "Payment Failed",
					"message":  fmt.Sprintf("Wallet payment failed: %s", err.Error()),
					"order_id": order.ID,
					"status":   "payment_failed",
				})
				return err
			}
			if err := s.checkoutRepo.UpdateWalletPaymentSucceededTx(tx, order.Payment.ID, order.UserID, order.ID); err != nil {
				return err
			}
		} else if err := s.paymentService.ProcessPayment(tx, order.Payment.ID, cardInfo); err != nil {
			if markErr := s.checkoutRepo.MarkOrderPaymentFailedTx(tx, order.ID); markErr != nil {
				return markErr
			}
			s.broadcastOrderUpdate(order.ID, order.UserID, "payment_failed", map[string]interface{}{
				"title":    "Payment Failed",
				"message":  fmt.Sprintf("Payment failed: %s", err.Error()),
				"order_id": order.ID,
				"status":   "payment_failed",
			})
			return err
		}

		if err := s.checkoutRepo.UpdateOrderStatusTx(tx, order.ID, constants.OrderStatusPaid); err != nil {
			return err
		}

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

	return s.processShipment(ctx, orderID)
}

func (s *Service) broadcastOrderUpdate(orderID, userID uint, eventType string, data map[string]interface{}) {
	msg := websocket.Message{
		Type:      eventType,
		UserID:    userID,
		Data:      data,
		Timestamp: time.Now(),
	}
	roomID := fmt.Sprintf("order_%d", orderID)
	s.hub.BroadcastToRoom(roomID, msg)

	go func() {
		_ = s.notifier.CreateNotification(
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
func (s *Service) processShipment(ctx context.Context, orderID uint) error {
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
func (s *Service) getCarrier(shippingProviderID *uint) string {
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
func (s *Service) reserveCartStock(ctx context.Context, tx *gorm.DB, orderID uint, cartItems []models.CartItem) ([]appinventory.DeltaResult, error) {
	return ReserveCartStock(ctx, tx, s.checkoutRepo, orderID, cartItems, func(ctx context.Context, tx *gorm.DB, orderID, productID uint, qty int) (StockDelta, error) {
		if s.inventoryService == nil {
			return StockDelta{}, nil
		}
		return s.inventoryService.DecrementForSale(ctx, tx, orderID, productID, qty)
	})
}

// applyCoupon validates a coupon code (if provided) and returns the discount amount and coupon ID.
func (s *Service) applyCoupon(ctx context.Context, tx *gorm.DB, userID uint, code string, subtotal float64) (float64, *uint, error) {
	// No coupon code → no discount
	if code == "" {
		return 0, nil, nil
	}

	// Validate the coupon (ideally pass tx for atomicity)
	coupon, discount, err := s.couponService.ValidateCoupon(ctx, code, userID, subtotal)
	if err != nil {
		return 0, nil, err
	}

	return discount, &coupon.ID, nil
}

// createPayment uses the PaymentService to insert a pending payment record (within tx).
func (s *Service) createPayment(tx *gorm.DB, orderID, userID uint, amount float64, method, currency string) error {
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
func (s *Service) createShipment(tx *gorm.DB, orderID, userID uint, carrier string, req dto.CheckoutRequest) error {
	shipReq := appshipment.CreateShipmentRequest{
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
func (s *Service) enqueueFulfillmentJob(ctx context.Context, orderID uint, req dto.CheckoutRequest) {
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
func (s *Service) sendOrderCreatedNotification(userID uint, order *models.Order) {
	go func() {
		_ = s.notifier.CreateNotification(
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

// CancelOrder cancels an order belonging to userID, restores stock, and refunds
// wallet payments. Stripe orders are cancelled without an automatic refund (requires
// manual processing via the Stripe dashboard).
func (s *Service) CancelOrder(ctx context.Context, orderID, userID uint) error {
	ord, err := s.checkoutRepo.FindOrderForCancel(ctx, orderID, userID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return utils.ErrNotFound("order not found")
		}
		return utils.ErrInternal(err)
	}

	if err := s.orderDomain.CanCancel(order.Order{
		ID:     ord.ID,
		UserID: ord.UserID,
		Status: ord.Status,
	}); err != nil {
		return utils.ErrBadRequest(err.Error())
	}

	var restores []appinventory.DeltaResult
	err = s.checkoutRepo.Transaction(ctx, func(tx *gorm.DB) error {
		for _, item := range ord.Items {
			if s.inventoryService != nil {
				change, err := s.inventoryService.RestoreForOrderCancel(ctx, tx, ord.ID, item.ProductID, item.Quantity)
				if err != nil {
					return err
				}
				restores = append(restores, change)
				continue
			}
			if err := s.checkoutRepo.RestoreProductStockTx(tx, item.ProductID, item.Quantity); err != nil {
				return utils.ErrInternal(err)
			}
		}

		if ord.Payment != nil && order.ShouldRefundWalletOnCancel(ord.Payment.Method, ord.Payment.Status) {
			if err := s.walletService.Refund(ctx, ord.UserID, ord.TotalAmount, ord.ID); err != nil {
				return err
			}
			if err := s.checkoutRepo.UpdatePaymentRefundedTx(tx, ord.Payment.ID); err != nil {
				return utils.ErrInternal(err)
			}
		}

		// Cancel any pending/processing shipment.
		return s.checkoutRepo.CancelPendingShipmentsTx(tx, ord.ID)
	})
	if err != nil {
		return err
	}

	for _, change := range restores {
		if change.QuantityBefore != change.QuantityAfter && s.inventoryService != nil {
			s.inventoryService.RunStockSideEffects(ctx, change.Product, change.QuantityBefore, change.QuantityAfter)
		}
	}

	if err := appworkflow.ApplyEvent(ctx, s.engine, workflow.TransitionRequest{
		WorkflowKey: constants.WorkflowEntityOrder,
		EntityID:    ord.ID,
		Event:       "cancel",
		ActorID:     &userID,
		ActorRole:   constants.RoleUser,
	}); err != nil {
		// Fallback when transition rules reject the move (e.g. stale workflow state).
		s.setOrderStateActor(ctx, ord.ID, "cancelled", "cancel", &userID)
	}

	return nil
}

func (s *Service) ensureInvoice(ctx context.Context, orderID uint) {
	if s.invoiceCommands == nil {
		return
	}
	if _, err := s.invoiceCommands.CreateFromPaidOrder(ctx, orderID); err != nil {
		utils.Log.WithError(err).WithField("order_id", orderID).Warn("failed to create invoice for paid order")
	}
}

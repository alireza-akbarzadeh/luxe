package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/tasks"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/alireza-akbarzadeh/luxe/internal/websocket"
	"gorm.io/gorm"
)

type CheckoutServiceInterface interface {
	Checkout(userID uint, req dto.CheckoutRequest) (*models.Order, error)
}

type checkoutService struct {
	db                  *gorm.DB
	notificationService NotificationServiceInterface
	couponService       CouponServiceInterface
	paymentService      PaymentServiceInterface
	shipmentService     ShipmentServiceInterface
	workerPool          *tasks.WorkerPool
	hub                 *websocket.Hub
	salesFeed           *SalesFeedService
}

func NewCheckoutService(
	db *gorm.DB,
	notificationService NotificationServiceInterface,
	couponService CouponServiceInterface,
	paymentService PaymentServiceInterface,
	shipmentService ShipmentServiceInterface,
	workerPool *tasks.WorkerPool,
	hub *websocket.Hub,
	salesFeed *SalesFeedService,
) CheckoutServiceInterface {
	return &checkoutService{
		db:                  db,
		notificationService: notificationService,
		couponService:       couponService,
		paymentService:      paymentService,
		shipmentService:     shipmentService,
		workerPool:          workerPool,
		hub:                 hub,
		salesFeed:           salesFeed,
	}
}

// Checkout converts the user's active cart into an order.
func (s *checkoutService) Checkout(userID uint, req dto.CheckoutRequest) (*models.Order, error) {
	cart, err := s.getActiveCart(userID)
	if err != nil {
		return nil, err
	}
	address, err := s.resolveAddress(userID, req)
	if err != nil {
		return nil, err
	}
	carrier := s.getCarrier(req.ShippingProviderID)

	var order *models.Order
	err = s.db.Transaction(func(tx *gorm.DB) error {
		subtotal, err := s.reserveCartStock(tx, cart.Items)
		if err != nil {
			return err
		}
		discount, couponID, err := s.applyCoupon(tx, userID, req.CouponCode, subtotal)
		if err != nil {
			return err
		}
		totalAmount := subtotal - discount
		if totalAmount < 0 {
			totalAmount = 0
		}

		order, err = s.createOrderRecord(tx, userID, totalAmount, address.ID)
		if err != nil {
			return err
		}

		if err := s.createOrderItems(tx, order.ID, cart.Items); err != nil {
			return err
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

		return s.markCartConverted(tx, cart.ID)
	})
	if err != nil {
		return nil, err
	}

	s.db.Preload("Items.Product").Preload("User").Preload("Payment").First(order, order.ID)
	s.enqueueFulfillmentJob(order.ID, req)
	s.sendOrderCreatedNotification(userID, order)

	return order, nil
}

// ProcessOrder is the background job handler that orchestrates payment and shipment.
func (s *checkoutService) ProcessOrder(orderID uint, cardInfo dto.CardInfo) error {
	// Use a transaction for the payment + order status update
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Load order with payment
		var order models.Order
		if err := tx.Preload("Payment").First(&order, orderID).Error; err != nil {
			return fmt.Errorf("order not found: %w", err)
		}
		if order.Payment == nil {
			return fmt.Errorf("payment record missing for order %d", orderID)
		}

		// 2. Process payment (mock gateway inside)
		if err := s.paymentService.ProcessPayment(tx, order.Payment.ID, cardInfo); err != nil {
			// Payment failed – update order status & cancel shipment
			tx.Model(&order).Update("status", "payment_failed")
			tx.Model(&models.Shipment{}).Where("order_id = ?", order.ID).Update("status", "cancelled")

			// Broadcast failure
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

	// Transaction committed – now handle shipment (outside transaction for performance)
	return s.processShipment(orderID)
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
func (s *checkoutService) processShipment(orderID uint) error {
	// 1. Mark shipment as processing
	var shipment models.Shipment
	if err := s.db.Where("order_id = ?", orderID).First(&shipment).Error; err != nil {
		return fmt.Errorf("shipment not found: %w", err)
	}

	oldStatus := shipment.Status
	shipment.Status = "processing"
	if err := s.db.Save(&shipment).Error; err != nil {
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
	s.db.Model(&shipment).Updates(map[string]interface{}{
		"status":          constants.ShipmentStatusShipped,
		"tracking_number": trackingNumber,
		"shipped_at":      now,
	})

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

func (s *checkoutService) getActiveCart(userID uint) (*models.Cart, error) {
	var cart models.Cart
	err := s.db.Where("user_id = ? AND status = ?", userID, "active").
		Preload("Items.Product").
		First(&cart).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrBadRequest("cart is empty")
		}
		return nil, utils.ErrInternal(err)
	}
	if len(cart.Items) == 0 {
		return nil, utils.ErrBadRequest("cart is empty")
	}
	return &cart, nil
}

// markCartConverted sets the cart status to "converted" inside a transaction.
func (s *checkoutService) markCartConverted(tx *gorm.DB, cartID uint) error {
	return tx.Model(&models.Cart{}).Where("id = ?", cartID).
		Update("status", "converted").Error
}

// resolveAddress finds or creates an address record for the user.
func (s *checkoutService) resolveAddress(userID uint, req dto.CheckoutRequest) (*models.Address, error) {
	address := dto.MapAddress(userID, req)
	err := s.db.Where("user_id = ? AND address_line1 = ? AND postal_code = ?",
		userID, req.AddressLine1, req.Zip).
		FirstOrCreate(&address, address).Error
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	return &address, nil
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
	return DefaultShippingProvider
}

// reserveCartStock locks product rows, validates stock, and decrements inventory.
func (s *checkoutService) reserveCartStock(tx *gorm.DB, cartItems []models.CartItem) (float64, error) {
	var subtotal float64
	for _, item := range cartItems {
		var product models.Product
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&product, item.ProductID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return 0, utils.ErrBadRequest("product not found")
			}
			return 0, utils.ErrInternal(err)
		}
		if !isProductStockAvailable(product, item.Quantity) {
			return 0, utils.ErrBadRequest(
				fmt.Sprintf("insufficient stock for product: %s", product.Name))
		}
		if shouldDecrementProductStock(product) {
			product.Stock -= item.Quantity
			if err := tx.Save(&product).Error; err != nil {
				return 0, utils.ErrInternal(err)
			}
		}
		subtotal += item.Price * float64(item.Quantity)
	}
	return subtotal, nil
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

// createOrderRecord inserts the order row inside the transaction.
func (s *checkoutService) createOrderRecord(tx *gorm.DB, userID uint, totalAmount float64, addressID uint) (*models.Order, error) {
	order := &models.Order{
		UserID:            userID,
		OrderNumber:       generateOrderNumber(userID),
		Status:            constants.OrderStatusPending,
		TotalAmount:       totalAmount,
		Currency:          "USD",
		ShippingAddressID: &addressID,
		BillingAddressID:  &addressID,
	}
	if err := tx.Create(order).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}
	return order, nil
}

// createOrderItems copies cart items into order items and decrements stock.
func (s *checkoutService) createOrderItems(tx *gorm.DB, orderID uint, cartItems []models.CartItem) error {
	for _, item := range cartItems {
		oi := &models.OrderItem{
			OrderID:   orderID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
		if err := tx.Create(oi).Error; err != nil {
			return utils.ErrInternal(err)
		}
	}
	return nil
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

// enqueueFulfillmentJob submits the ProcessOrder job to the worker pool.
func (s *checkoutService) enqueueFulfillmentJob(orderID uint, req dto.CheckoutRequest) {
	cardInfo := dto.CardInfo{
		CardNumber:  req.CardNumber,
		ExpiryMonth: req.ExpiryMonth,
		ExpiryYear:  req.ExpiryYear,
		CVV:         req.CVV,
	}
	job := tasks.Job{
		ID:      fmt.Sprintf("fulfill_%d", orderID),
		Payload: orderID,
		Handler: func(payload interface{}) error {
			id, ok := payload.(uint)
			if !ok {
				return fmt.Errorf("invalid payload type")
			}
			return s.ProcessOrder(id, cardInfo)
		},
	}
	s.workerPool.Enqueue(job)
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

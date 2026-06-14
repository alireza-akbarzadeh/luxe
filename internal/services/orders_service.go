package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/alireza-akbarzadeh/luxe/internal/websocket"
	"gorm.io/gorm"
)

type OrderServiceInterface interface {
	GetUserOrders(userID uint, filters dto.OrderListFilters) ([]models.Order, int64, error)
	GetOrderByID(orderID uint, userID uint) (*models.Order, error)
	GetAllOrders(filters AdminOrderFilters, limit, offset int) ([]models.Order, int64, error)
	UpdateOverdueOrders() error
	UpdateOrderStatus(orderID uint, status string) error
}

type orderService struct {
	db                  *gorm.DB
	notificationService NotificationServiceInterface
	hub                 *websocket.Hub
	salesFeed           *SalesFeedService
}

func NewOrderService(
	db *gorm.DB,
	notificationService NotificationServiceInterface,
	hub *websocket.Hub,
	salesFeed *SalesFeedService,
) OrderServiceInterface {
	return &orderService{
		db:                  db,
		notificationService: notificationService,
		hub:                 hub,
		salesFeed:           salesFeed,
	}
}

const (
	DefaultShippingProvider = "standard"
)

// GetUserOrders returns all orders for a user (paginated).
func (s *orderService) GetUserOrders(userID uint, filters dto.OrderListFilters) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	// Set defaults
	if filters.Limit == 0 {
		filters.Limit = 20
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}

	query := s.db.Model(&models.Order{}).Where("user_id = ?", userID)

	// Apply filters
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.FromDate != nil {
		query = query.Where("created_at >= ?", filters.FromDate)
	}
	if filters.ToDate != nil {
		query = query.Where("created_at <= ?", filters.ToDate)
	}
	if filters.MinAmount != nil {
		query = query.Where("total_amount >= ?", *filters.MinAmount)
	}
	if filters.MaxAmount != nil {
		query = query.Where("total_amount <= ?", *filters.MaxAmount)
	}

	// Count total matching records (efficient)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, utils.ErrInternal(err)
	}

	// Pagination with ordering – uses indexes
	if err := query.Limit(filters.Limit).Offset(filters.Offset).
		Preload("Items.Product").
		Preload("Payment").Preload("Shipment").
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		return nil, 0, utils.ErrInternal(err)
	}

	return orders, total, nil
}

// UpdateOrderStatus updates an order's status and sends real-time notification
func (s *orderService) UpdateOrderStatus(orderID uint, status string) error {
	var order models.Order
	if err := s.db.Preload("User").First(&order, orderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrNotFound("order not found")
		}
		return utils.ErrInternal(err)
	}

	oldStatus := order.Status
	order.Status = status

	if err := s.db.Save(&order).Error; err != nil {
		return utils.ErrInternal(err)
	}

	if s.hub != nil {
		roomID := fmt.Sprintf("order_%d", orderID)
		message := websocket.Message{
			Type:   "order_status_update",
			RoomID: roomID,
			Data: map[string]interface{}{
				"order_id":     order.ID,
				"order_number": order.OrderNumber,
				"old_status":   oldStatus,
				"new_status":   status,
				"updated_at":   order.UpdatedAt,
			},
			Timestamp: time.Now(),
		}
		s.hub.BroadcastToRoom(roomID, message)
	}

	if s.salesFeed != nil {
		eventType := "status_change"
		title := fmt.Sprintf("Order %s updated", order.OrderNumber)
		subtitle := fmt.Sprintf("Status changed to %s", status)
		amount := order.TotalAmount

		switch status {
		case constants.OrderStatusPending:
			eventType = "new_order"
			title = fmt.Sprintf("New order %s", order.OrderNumber)
			subtitle = fmt.Sprintf("Status: %s · $%.2f", status, amount)
		case constants.OrderStatusCancelled:
			eventType = "cancellation"
			title = "Order cancelled"
			subtitle = fmt.Sprintf("%s was cancelled", order.OrderNumber)
		case constants.OrderStatusShipped:
			eventType = "shipment"
			title = "Order shipped"
			subtitle = fmt.Sprintf("%s is on its way", order.OrderNumber)
		}

		s.salesFeed.PublishOrderEvent(eventType, title, subtitle, amount)
	}

	// Keep existing notification (persistent)
	go func() {
		title, msgText := s.getOrderStatusNotificationMessage(status, order.OrderNumber)
		_ = s.notificationService.CreateNotification(
			order.UserID,
			"order_status_update",
			title,
			msgText,
			map[string]interface{}{
				"order_id":     order.ID,
				"order_number": order.OrderNumber,
				"old_status":   oldStatus,
				"new_status":   status,
				"updated_at":   order.UpdatedAt,
			},
		)
	}()

	return nil
}

// getOrderStatusNotificationMessage returns appropriate title and message for order status
func (s *orderService) getOrderStatusNotificationMessage(status, orderNumber string) (string, string) {
	switch status {
	case constants.OrderStatusPaid:
		return "Payment Confirmed", fmt.Sprintf("Payment for order #%s has been confirmed.", orderNumber)
	case constants.OrderStatusShipped:
		return "Order Shipped", fmt.Sprintf("Your order #%s has been shipped and is on its way!", orderNumber)
	case constants.OrderStatusDelivered:
		return "Order Delivered", fmt.Sprintf("Your order #%s has been delivered successfully.", orderNumber)
	case constants.OrderStatusCancelled:
		return "Order Cancelled", fmt.Sprintf("Your order #%s has been cancelled.", orderNumber)
	case constants.OrderStatusRefunded:
		return "Order Refunded", fmt.Sprintf("Your order #%s has been refunded.", orderNumber)
	default:
		return "Order Update", fmt.Sprintf("Your order #%s status has been updated to %s.", orderNumber, status)
	}
}

// GetOrderByID returns a single order by ID, verifying ownership.
func (s *orderService) GetOrderByID(orderID uint, userID uint) (*models.Order, error) {
	var order models.Order
	err := s.db.Where("id = ? AND user_id = ?", orderID, userID).
		Preload("Items.Product").
		Preload("Payment").
		Preload("Shipment").
		First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("order not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return &order, nil
}

// AdminOrderFilters adds user_id filter
type AdminOrderFilters struct {
	dto.OrderFilters
	UserID *uint `json:"user_id,omitempty"`
}

// GetAllOrders returns all orders (admin only) with advanced filters and pagination.
func (s *orderService) GetAllOrders(filters AdminOrderFilters, limit, offset int) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	query := s.db.Model(&models.Order{})

	// Apply filters
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.FromDate != nil {
		query = query.Where("created_at >= ?", filters.FromDate)
	}
	if filters.ToDate != nil {
		query = query.Where("created_at <= ?", filters.ToDate)
	}
	if filters.MinAmount != nil {
		query = query.Where("total_amount >= ?", *filters.MinAmount)
	}
	if filters.MaxAmount != nil {
		query = query.Where("total_amount <= ?", *filters.MaxAmount)
	}
	if filters.UserID != nil {
		query = query.Where("user_id = ?", *filters.UserID)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, utils.ErrInternal(err)
	}

	// Paginated results with preload
	if err := query.Limit(limit).Offset(offset).
		Preload("Items.Product").
		Preload("User").
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		return nil, 0, utils.ErrInternal(err)
	}

	return orders, total, nil
}

// UpdateOverdueOrders marks orders as 'delayed' if they have been 'paid' for more than 7 days.
func (s *orderService) UpdateOverdueOrders() error {
	cutoff := time.Now().Add(-7 * 24 * time.Hour)

	// Find paid orders older than cutoff that are not yet delivered or canceled
	var orders []models.Order
	err := s.db.Where("status = ? AND updated_at < ?", constants.OrderStatusPaid, cutoff).
		Not("status IN (?)", []string{constants.OrderStatusDelivered, constants.OrderStatusCancelled, constants.OrderStatusRefunded}).
		Find(&orders).Error
	if err != nil {
		return utils.ErrInternal(err)
	}

	if len(orders) == 0 {
		utils.Log.Info("No overdue orders found")
		return nil
	}

	// Mark them as 'delayed'
	for _, order := range orders {
		oldStatus := order.Status
		order.Status = "delayed"
		if err := s.db.Save(&order).Error; err != nil {
			utils.Log.WithError(err).Errorf("Failed to update order %d to delayed", order.ID)
		} else {
			utils.Log.Infof("Order %d marked as delayed", order.ID)

			// Send real-time notification for delayed order
			go func(order models.Order) {
				_ = s.notificationService.CreateNotification(
					order.UserID,
					"order_delayed",
					"Order Delayed",
					fmt.Sprintf("Your order #%s is experiencing a delay. We apologize for the inconvenience.", order.OrderNumber),
					map[string]interface{}{
						"order_id":     order.ID,
						"order_number": order.OrderNumber,
						"old_status":   oldStatus,
						"new_status":   "delayed",
						"updated_at":   order.UpdatedAt,
					},
				)
			}(order)
		}
	}
	return nil
}

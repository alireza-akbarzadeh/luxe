package services

import (
	"context"
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
	GetUserOrders(ctx context.Context, userID uint, filters dto.OrderListFilters) ([]models.Order, int64, error)
	GetOrderByID(ctx context.Context, orderID uint, userID uint) (*models.Order, error)
	GetAllOrders(ctx context.Context, filters AdminOrderFilters, limit, offset int) ([]models.Order, int64, error)
	UpdateOverdueOrders(ctx context.Context) error
	UpdateOrderStatus(ctx context.Context, orderID uint, status string) error
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

type orderListQuery struct {
	UserID      *uint
	Status      string
	FromDate    *time.Time
	ToDate      *time.Time
	MinAmount   *float64
	MaxAmount   *float64
	Limit       int
	Offset      int
	PreloadUser bool
}

func orderFiltersFromDTO(userID uint, filters dto.OrderListFilters) orderListQuery {
	return orderListQuery{
		UserID:    &userID,
		Status:    filters.Status,
		FromDate:  filters.FromDate,
		ToDate:    filters.ToDate,
		MinAmount: filters.MinAmount,
		MaxAmount: filters.MaxAmount,
		Limit:     filters.Limit,
		Offset:    filters.Offset,
	}
}

func (s *orderService) GetUserOrders(ctx context.Context, userID uint, filters dto.OrderListFilters) ([]models.Order, int64, error) {
	if filters.Limit == 0 {
		filters.Limit = 20
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}

	q := orderFiltersFromDTO(userID, filters)
	orders, total, err := s.listOrders(ctx, q)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	return orders, total, nil
}

func (s *orderService) UpdateOrderStatus(ctx context.Context, orderID uint, status string) error {
	order, err := s.findOrderByID(ctx, orderID, true)
	if err != nil {
		if isRecordNotFound(err) {
			return utils.ErrNotFound("order not found")
		}
		return utils.ErrInternal(err)
	}

	oldStatus := order.Status
	order.Status = status

	if err := s.db.WithContext(ctx).Save(order).Error; err != nil {
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

func (s *orderService) GetOrderByID(ctx context.Context, orderID uint, userID uint) (*models.Order, error) {
	order, err := s.findOrderByIDAndUserID(ctx, orderID, userID)
	if err != nil {
		if isRecordNotFound(err) {
			return nil, utils.ErrNotFound("order not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return order, nil
}

type AdminOrderFilters struct {
	dto.OrderFilters
	UserID *uint `json:"user_id,omitempty"`
}

func (s *orderService) GetAllOrders(ctx context.Context, filters AdminOrderFilters, limit, offset int) ([]models.Order, int64, error) {
	q := orderListQuery{
		UserID:      filters.UserID,
		Status:      filters.Status,
		FromDate:    filters.FromDate,
		ToDate:      filters.ToDate,
		MinAmount:   filters.MinAmount,
		MaxAmount:   filters.MaxAmount,
		Limit:       limit,
		Offset:      offset,
		PreloadUser: true,
	}

	orders, total, err := s.listOrders(ctx, q)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	return orders, total, nil
}

func (s *orderService) UpdateOverdueOrders(ctx context.Context) error {
	cutoff := time.Now().Add(-7 * 24 * time.Hour)

	var orders []models.Order
	err := s.db.WithContext(ctx).
		Where("status = ? AND updated_at < ?", constants.OrderStatusPaid, cutoff).
		Not("status IN (?)", []string{constants.OrderStatusDelivered, constants.OrderStatusCancelled, constants.OrderStatusRefunded}).
		Find(&orders).Error
	if err != nil {
		return utils.ErrInternal(err)
	}

	if len(orders) == 0 {
		utils.Log.Info("No overdue orders found")
		return nil
	}

	for _, order := range orders {
		oldStatus := order.Status
		order.Status = constants.OrderStatusDelayed
		if err := s.db.WithContext(ctx).Save(&order).Error; err != nil {
			utils.Log.WithError(err).Errorf("Failed to update order %d to delayed", order.ID)
			continue
		}
		utils.Log.Infof("Order %d marked as delayed", order.ID)

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
					"new_status":   constants.OrderStatusDelayed,
					"updated_at":   order.UpdatedAt,
				},
			)
		}(order)
	}
	return nil
}

func (s *orderService) applyOrderListFilters(query *gorm.DB, q orderListQuery) *gorm.DB {
	if q.UserID != nil {
		query = query.Where("user_id = ?", *q.UserID)
	}
	if q.Status != "" {
		query = query.Where("status = ?", q.Status)
	}
	if q.FromDate != nil {
		query = query.Where("created_at >= ?", q.FromDate)
	}
	if q.ToDate != nil {
		query = query.Where("created_at <= ?", q.ToDate)
	}
	if q.MinAmount != nil {
		query = query.Where("total_amount >= ?", *q.MinAmount)
	}
	if q.MaxAmount != nil {
		query = query.Where("total_amount <= ?", *q.MaxAmount)
	}
	return query
}

func (s *orderService) listOrders(ctx context.Context, q orderListQuery) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	query := s.applyOrderListFilters(s.db.WithContext(ctx).Model(&models.Order{}), q)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	listQuery := s.applyOrderListFilters(s.db.WithContext(ctx).Model(&models.Order{}), q)
	listQuery = listQuery.Limit(q.Limit).Offset(q.Offset).
		Preload("Items.Product").
		Preload("Payment").
		Preload("Shipment").
		Order("created_at DESC")

	if q.PreloadUser {
		listQuery = listQuery.Preload("User")
	}

	if err := listQuery.Find(&orders).Error; err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

func (s *orderService) findOrderByIDAndUserID(ctx context.Context, orderID, userID uint) (*models.Order, error) {
	var order models.Order
	err := s.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", orderID, userID).
		Preload("Items.Product").
		Preload("Payment").
		Preload("Shipment").
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (s *orderService) findOrderByID(ctx context.Context, orderID uint, preloadUser bool) (*models.Order, error) {
	q := s.db.WithContext(ctx)
	if preloadUser {
		q = q.Preload("User")
	}
	var order models.Order
	if err := q.First(&order, orderID).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

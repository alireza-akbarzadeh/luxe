package services

import (
	"context"
	"fmt"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/repositories"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/alireza-akbarzadeh/luxe/internal/websocket"
)

type OrderServiceInterface interface {
	GetUserOrders(userID uint, filters dto.OrderListFilters) ([]models.Order, int64, error)
	GetOrderByID(orderID uint, userID uint) (*models.Order, error)
	GetAllOrders(filters AdminOrderFilters, limit, offset int) ([]models.Order, int64, error)
	UpdateOverdueOrders() error
	UpdateOrderStatus(orderID uint, status string) error
}

type orderService struct {
	orders              repositories.OrderRepository
	notificationService NotificationServiceInterface
	hub                 *websocket.Hub
	salesFeed           *SalesFeedService
}

func NewOrderService(
	orders repositories.OrderRepository,
	notificationService NotificationServiceInterface,
	hub *websocket.Hub,
	salesFeed *SalesFeedService,
) OrderServiceInterface {
	return &orderService{
		orders:              orders,
		notificationService: notificationService,
		hub:                 hub,
		salesFeed:           salesFeed,
	}
}

const (
	DefaultShippingProvider = "standard"
)

func (s *orderService) GetUserOrders(userID uint, filters dto.OrderListFilters) ([]models.Order, int64, error) {
	if filters.Limit == 0 {
		filters.Limit = 20
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}

	q := repositories.OrderFiltersFromDTO(userID, filters)
	orders, total, err := s.orders.List(context.Background(), q)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	return orders, total, nil
}

func (s *orderService) UpdateOrderStatus(orderID uint, status string) error {
	ctx := context.Background()
	order, err := s.orders.FindByID(ctx, orderID, true)
	if err != nil {
		if repositories.IsRecordNotFound(err) {
			return utils.ErrNotFound("order not found")
		}
		return utils.ErrInternal(err)
	}

	oldStatus := order.Status
	order.Status = status

	if err := s.orders.Save(ctx, order); err != nil {
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

func (s *orderService) GetOrderByID(orderID uint, userID uint) (*models.Order, error) {
	order, err := s.orders.FindByIDAndUserID(context.Background(), orderID, userID)
	if err != nil {
		if repositories.IsRecordNotFound(err) {
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

func (s *orderService) GetAllOrders(filters AdminOrderFilters, limit, offset int) ([]models.Order, int64, error) {
	q := repositories.OrderListQuery{
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

	orders, total, err := s.orders.List(context.Background(), q)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	return orders, total, nil
}

func (s *orderService) UpdateOverdueOrders() error {
	ctx := context.Background()
	cutoff := time.Now().Add(-7 * 24 * time.Hour)

	orders, err := s.orders.FindOverduePaid(ctx, cutoff)
	if err != nil {
		return utils.ErrInternal(err)
	}

	if len(orders) == 0 {
		utils.Log.Info("No overdue orders found")
		return nil
	}

	for _, order := range orders {
		oldStatus := order.Status
		order.Status = "delayed"
		if err := s.orders.Save(ctx, &order); err != nil {
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
					"new_status":   "delayed",
					"updated_at":   order.UpdatedAt,
				},
			)
		}(order)
	}
	return nil
}

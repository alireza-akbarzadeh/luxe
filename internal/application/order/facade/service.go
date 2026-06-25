package facade

import (
	"context"
	"fmt"
	"time"

	apporder "github.com/alireza-akbarzadeh/luxe/internal/application/order"
	appworkflow "github.com/alireza-akbarzadeh/luxe/internal/application/workflow"
	appsalesfeed "github.com/alireza-akbarzadeh/luxe/internal/application/salesfeed"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/asynq"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/alireza-akbarzadeh/luxe/internal/websocket"
	"gorm.io/gorm"
)

// Notifier sends in-app notifications for order status changes.
type Notifier interface {
	CreateNotification(userID uint, notificationType, title, message string, data interface{}) error
}

// Service orchestrates order queries, status updates, and workflow transitions.
type Service struct {
	notifier  Notifier
	hub       *websocket.Hub
	salesFeed *appsalesfeed.Service
	jobQueue  asynq.JobQueue
	engine    *workflow.Engine
	queries   *apporder.Queries
	commands  *apporder.Commands
}

// NewService wires order queries and commands.
func NewService(
	db *gorm.DB,
	notifier Notifier,
	hub *websocket.Hub,
	salesFeed *appsalesfeed.Service,
	jobQueue asynq.JobQueue,
	engine *workflow.Engine,
) *Service {
	repo := postgres.NewOrderRepository(db)
	return &Service{
		notifier:  notifier,
		hub:       hub,
		salesFeed: salesFeed,
		jobQueue:  jobQueue,
		engine:    engine,
		queries:   apporder.NewQueries(repo),
		commands:  apporder.NewCommands(repo),
	}
}

func (s *Service) applyOrderState(ctx context.Context, order *models.Order, status string, actorID *uint) error {
	actorRole := constants.RoleAdmin
	if appworkflow.ApplyOrderWorkflow(ctx, s.engine, order.ID, status, actorRole, actorID) {
		order.Status = status
		return nil
	}
	if s.engine != nil {
		utils.Log.WithField("order_id", order.ID).WithField("status", status).
			Warn("workflow update failed; writing status directly")
	}
	order.Status = status
	return s.commands.UpdateStatus(ctx, order.ID, status)
}

func (s *Service) GetUserOrders(ctx context.Context, userID uint, filters dto.OrderListFilters) ([]models.Order, int64, error) {
	return s.queries.ListUserOrders(ctx, userID, filters)
}

func (s *Service) UpdateOrderStatus(ctx context.Context, orderID uint, status string, actorID *uint) error {
	order, err := s.queries.FindByID(ctx, orderID, true)
	if err != nil {
		return err
	}

	oldStatus := order.Status

	if err := s.applyOrderState(ctx, order, status, actorID); err != nil {
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
		_ = s.notifier.CreateNotification(
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

	s.enqueueOrderStatusEmail(ctx, order, status)

	return nil
}

func (s *Service) enqueueOrderStatusEmail(ctx context.Context, order *models.Order, status string) {
	emailStatuses := map[string]bool{
		constants.OrderStatusShipped:   true,
		constants.OrderStatusDelivered: true,
		constants.OrderStatusCancelled: true,
	}
	if !emailStatuses[status] || s.jobQueue == nil {
		return
	}

	email := order.User.Email
	if email == "" {
		return
	}

	subject, body := s.getOrderStatusNotificationMessage(status, order.OrderNumber)
	if err := s.jobQueue.EnqueueSendEmail(ctx, email, subject, body); err != nil {
		utils.Log.WithError(err).WithField("order_id", order.ID).Warn("failed to enqueue order status email")
	}
}

func (s *Service) getOrderStatusNotificationMessage(status, orderNumber string) (string, string) {
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

func (s *Service) GetOrderByID(ctx context.Context, orderID uint, userID uint) (*models.Order, error) {
	return s.queries.GetByIDForUser(ctx, orderID, userID)
}

func (s *Service) GetAllOrders(ctx context.Context, filters apporder.AdminOrderFilters, limit, offset int) ([]models.Order, int64, error) {
	return s.queries.ListAdmin(ctx, apporder.ListFilter{
		UserID:      filters.UserID,
		Status:      filters.Status,
		Search:      filters.Search,
		FromDate:    filters.FromDate,
		ToDate:      filters.ToDate,
		MinAmount:   filters.MinAmount,
		MaxAmount:   filters.MaxAmount,
		Limit:       limit,
		Offset:      offset,
		PreloadUser: true,
	})
}

func (s *Service) BulkUpdateOrderStatus(ctx context.Context, orderIDs []uint, status string, actorID *uint) (int64, error) {
	if len(orderIDs) == 0 {
		return 0, utils.ErrBadRequest("no order IDs provided")
	}
	validStatuses := map[string]bool{
		constants.OrderStatusPaid:      true,
		constants.OrderStatusShipped:   true,
		constants.OrderStatusDelivered: true,
		constants.OrderStatusCancelled: true,
	}
	if !validStatuses[status] {
		return 0, utils.ErrBadRequest("invalid bulk status; allowed: paid, shipped, delivered, cancelled")
	}

	orderStatusToStateCode := map[string]string{
		constants.OrderStatusPending:   "pending_payment",
		constants.OrderStatusPaid:      "paid",
		"processing":                   "processing",
		constants.OrderStatusShipped:   "shipped",
		constants.OrderStatusDelivered: "delivered",
		"completed":                    "completed",
		constants.OrderStatusCancelled: "cancelled",
		constants.OrderStatusRefunded:  "refunded",
	}
	code := orderStatusToStateCode[status]
	var updated int64
	for _, id := range orderIDs {
		if appworkflow.ApplyOrderWorkflow(ctx, s.engine, id, status, constants.RoleAdmin, actorID) {
			updated++
			continue
		}
		if code != "" {
			appworkflow.SyncState(ctx, s.engine, constants.WorkflowEntityOrder, id, code, "admin_bulk_set", actorID)
		}
		n, err := s.commands.UpdateStatusByIDs(ctx, []uint{id}, status)
		if err == nil {
			updated += n
		}
	}
	return updated, nil
}

func (s *Service) UpdateOverdueOrders(ctx context.Context) error {
	orders, err := s.commands.FindOverduePaid(ctx)
	if err != nil {
		return utils.ErrInternal(err)
	}

	if len(orders) == 0 {
		utils.Log.Info("No overdue orders found")
		return nil
	}

	for _, order := range orders {
		oldStatus := order.Status
		if err := s.commands.SaveDelayed(ctx, &order); err != nil {
			utils.Log.WithError(err).Errorf("Failed to update order %d to delayed", order.ID)
			continue
		}
		utils.Log.Infof("Order %d marked as delayed", order.ID)

		go func(order models.Order) {
			_ = s.notifier.CreateNotification(
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

func (s *Service) GetOrderAdmin(ctx context.Context, orderID uint) (*models.Order, error) {
	return s.queries.GetAdmin(ctx, orderID)
}

func (s *Service) AvailableTransitions(ctx context.Context, orderID uint) (*models.WorkflowState, []models.WorkflowTransition, error) {
	if s.engine == nil {
		return nil, nil, utils.ErrInternal(fmt.Errorf("workflow engine not configured"))
	}
	if _, err := s.GetOrderAdmin(ctx, orderID); err != nil {
		return nil, nil, err
	}
	return s.engine.AvailableTransitions(ctx, constants.WorkflowEntityOrder, orderID)
}

func (s *Service) PerformTransition(
	ctx context.Context,
	orderID uint,
	event, note, actorRole string,
	actorID *uint,
) (*workflow.TransitionResult, error) {
	if s.engine == nil {
		return nil, utils.ErrInternal(fmt.Errorf("workflow engine not configured"))
	}
	if _, err := s.GetOrderAdmin(ctx, orderID); err != nil {
		return nil, err
	}
	return s.engine.Transition(ctx, workflow.TransitionRequest{
		WorkflowKey: constants.WorkflowEntityOrder,
		EntityID:    orderID,
		Event:       event,
		ActorID:     actorID,
		ActorRole:   actorRole,
		Note:        note,
	})
}

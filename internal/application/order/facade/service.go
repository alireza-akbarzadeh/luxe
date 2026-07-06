package facade

import (
	"context"
	"fmt"
	"time"

	apporder "github.com/alireza-akbarzadeh/luxe/internal/application/order"
	appworkflow "github.com/alireza-akbarzadeh/luxe/internal/application/workflow"
	appsalesfeed "github.com/alireza-akbarzadeh/luxe/internal/application/salesfeed"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	domainorder "github.com/alireza-akbarzadeh/luxe/internal/domain/order"
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

// VendorStoreNotifier pushes store-scoped realtime events for vendor dashboards.
type VendorStoreNotifier interface {
	NotifyVendorStoresForOrder(
		ctx context.Context,
		orderID uint,
		wsEventType, notifType, title, message string,
		data map[string]interface{},
	)
}

// Service orchestrates order queries, status updates, and workflow transitions.
type Service struct {
	notifier      Notifier
	vendorNotify  VendorStoreNotifier
	hub           *websocket.Hub
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
	vendorNotify VendorStoreNotifier,
	hub *websocket.Hub,
	salesFeed *appsalesfeed.Service,
	jobQueue asynq.JobQueue,
	engine *workflow.Engine,
) *Service {
	repo := postgres.NewOrderRepository(db)
	return &Service{
		notifier:     notifier,
		vendorNotify: vendorNotify,
		hub:          hub,
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
		data := map[string]interface{}{
			"order_id":     order.ID,
			"order_number": order.OrderNumber,
			"old_status":   oldStatus,
			"new_status":   status,
			"updated_at":   order.UpdatedAt,
		}
		_ = s.notifier.CreateNotification(
			order.UserID,
			"order_status_update",
			title,
			msgText,
			data,
		)
		if s.vendorNotify != nil {
			s.vendorNotify.NotifyVendorStoresForOrder(
				ctx,
				order.ID,
				websocket.EventVendorOrderUpdate,
				"vendor_order_update",
				title,
				msgText,
				data,
			)
		}
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
	if !domainorder.IsValidBulkAdminStatus(status) {
		return 0, utils.ErrBadRequest("invalid bulk status; allowed: paid, shipped, delivered, cancelled")
	}

	code := domainorder.WorkflowStateCode(status)
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
	trackingNumber string,
) (*workflow.TransitionResult, error) {
	if s.engine == nil {
		return nil, utils.ErrInternal(fmt.Errorf("workflow engine not configured"))
	}
	if _, err := s.GetOrderAdmin(ctx, orderID); err != nil {
		return nil, err
	}
	return s.transitionOrder(ctx, orderID, event, note, actorRole, actorID, trackingNumber)
}

func (s *Service) VendorAvailableTransitions(ctx context.Context, storeID, orderID uint) (*models.WorkflowState, []models.WorkflowTransition, error) {
	if s.engine == nil {
		return nil, nil, utils.ErrInternal(fmt.Errorf("workflow engine not configured"))
	}
	if _, err := s.GetVendorStoreOrder(ctx, storeID, orderID); err != nil {
		return nil, nil, err
	}
	return s.engine.AvailableTransitions(ctx, constants.WorkflowEntityOrder, orderID)
}

func (s *Service) VendorPerformTransition(
	ctx context.Context,
	storeID, orderID uint,
	event, note, actorRole string,
	actorID *uint,
	trackingNumber string,
) (*workflow.TransitionResult, error) {
	if s.engine == nil {
		return nil, utils.ErrInternal(fmt.Errorf("workflow engine not configured"))
	}
	if _, err := s.GetVendorStoreOrder(ctx, storeID, orderID); err != nil {
		return nil, err
	}
	return s.transitionOrder(ctx, orderID, event, note, actorRole, actorID, trackingNumber)
}

func (s *Service) transitionOrder(
	ctx context.Context,
	orderID uint,
	event, note, actorRole string,
	actorID *uint,
	trackingNumber string,
) (*workflow.TransitionResult, error) {
	orderBefore, err := s.queries.FindByID(ctx, orderID, true)
	if err != nil {
		return nil, err
	}
	oldStatus := orderBefore.Status

	result, err := s.engine.Transition(ctx, workflow.TransitionRequest{
		WorkflowKey: constants.WorkflowEntityOrder,
		EntityID:    orderID,
		Event:       event,
		ActorID:     actorID,
		ActorRole:   actorRole,
		Note:        note,
	})
	if err != nil {
		return nil, err
	}

	if event == "ship" {
		if err := s.commands.MarkShipmentShipped(ctx, orderID, trackingNumber); err != nil {
			utils.Log.WithError(err).WithField("order_id", orderID).Warn("failed to update shipment after ship transition")
		}
	}

	orderAfter, reloadErr := s.GetOrderAdmin(ctx, orderID)
	if reloadErr != nil {
		return result, nil
	}

	s.publishOrderTransitionEffects(ctx, orderAfter, oldStatus, orderAfter.Status, event, trackingNumber)
	return result, nil
}

func (s *Service) publishOrderTransitionEffects(
	ctx context.Context,
	orderAfter *models.Order,
	oldStatus, newStatus, event, trackingNumber string,
) {
	if oldStatus == newStatus {
		return
	}

	if s.hub != nil && orderAfter != nil {
		roomID := fmt.Sprintf("order_%d", orderAfter.ID)
		s.hub.BroadcastToRoom(roomID, websocket.Message{
			Type:   "order_status_update",
			RoomID: roomID,
			Data: map[string]interface{}{
				"order_id":     orderAfter.ID,
				"order_number": orderAfter.OrderNumber,
				"old_status":   oldStatus,
				"new_status":   newStatus,
				"updated_at":   orderAfter.UpdatedAt,
			},
			Timestamp: time.Now(),
		})

		if event == "ship" || newStatus == constants.OrderStatusShipped {
			shipmentData := map[string]interface{}{
				"title":           "Package Shipped",
				"message":         fmt.Sprintf("Your order #%s has been shipped.", orderAfter.OrderNumber),
				"order_id":        orderAfter.ID,
				"order_number":    orderAfter.OrderNumber,
				"status":          constants.ShipmentStatusShipped,
				"tracking_number": trackingNumber,
			}
			if orderAfter.Shipment != nil {
				shipmentData["carrier"] = orderAfter.Shipment.Carrier
				if trackingNumber == "" && orderAfter.Shipment.TrackingNumber != "" {
					shipmentData["tracking_number"] = orderAfter.Shipment.TrackingNumber
				}
			}
			s.hub.BroadcastToRoom(roomID, websocket.Message{
				Type:      "shipment_shipped",
				RoomID:    roomID,
				Data:      shipmentData,
				Timestamp: time.Now(),
			})
		}
	}

	if s.salesFeed != nil && orderAfter != nil {
		eventType := "status_change"
		title := fmt.Sprintf("Order %s updated", orderAfter.OrderNumber)
		subtitle := fmt.Sprintf("Status changed to %s", newStatus)
		if newStatus == constants.OrderStatusShipped {
			eventType = "shipment"
			title = "Order shipped"
			subtitle = fmt.Sprintf("%s is on its way", orderAfter.OrderNumber)
		}
		s.salesFeed.PublishOrderEvent(eventType, title, subtitle, orderAfter.TotalAmount)
	}

	if orderAfter == nil || s.notifier == nil {
		return
	}

	go func() {
		title, msgText := s.getOrderStatusNotificationMessage(newStatus, orderAfter.OrderNumber)
		data := map[string]interface{}{
			"order_id":     orderAfter.ID,
			"order_number": orderAfter.OrderNumber,
			"old_status":   oldStatus,
			"new_status":   newStatus,
			"updated_at":   orderAfter.UpdatedAt,
		}
		_ = s.notifier.CreateNotification(
			orderAfter.UserID,
			"order_status_update",
			title,
			msgText,
			data,
		)
		if s.vendorNotify != nil {
			wsEvent := websocket.EventVendorOrderUpdate
			notifType := "vendor_order_update"
			if newStatus == constants.OrderStatusShipped {
				wsEvent = websocket.EventVendorOrderShipment
				notifType = "vendor_order_shipment"
			}
			s.vendorNotify.NotifyVendorStoresForOrder(
				ctx,
				orderAfter.ID,
				wsEvent,
				notifType,
				title,
				msgText,
				data,
			)
		}
	}()

	s.enqueueOrderStatusEmail(ctx, orderAfter, newStatus)
}

func (s *Service) ListVendorStoreOrders(
	ctx context.Context,
	storeID uint,
	filters apporder.AdminOrderFilters,
	limit, offset int,
) ([]models.Order, int64, error) {
	return s.queries.ListVendorStore(ctx, storeID, apporder.ListFilter{
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

func (s *Service) GetVendorStoreOrderStats(ctx context.Context, storeID uint) (apporder.VendorOrderStats, error) {
	return s.queries.GetVendorStoreStats(ctx, storeID)
}

func (s *Service) GetVendorStoreOrder(ctx context.Context, storeID, orderID uint) (*models.Order, error) {
	return s.queries.GetVendorStoreOrder(ctx, storeID, orderID)
}

func (s *Service) GetVendorStoreSalesSummary(ctx context.Context, storeID uint, from, to time.Time) (apporder.VendorSalesSummary, error) {
	return s.queries.GetVendorStoreSalesSummary(ctx, storeID, from, to)
}

func (s *Service) ListVendorTopProducts(ctx context.Context, storeID uint, from, to time.Time, limit int) ([]apporder.VendorTopProduct, error) {
	return s.queries.ListVendorTopProducts(ctx, storeID, from, to, limit)
}

func (s *Service) ListVendorDailySales(ctx context.Context, storeID uint, from, to time.Time) ([]apporder.VendorDailySales, error) {
	return s.queries.ListVendorDailySales(ctx, storeID, from, to)
}

func (s *Service) ListVendorStoreCustomers(ctx context.Context, storeID uint, from, to time.Time, limit int) ([]apporder.VendorStoreCustomer, error) {
	return s.queries.ListVendorStoreCustomers(ctx, storeID, from, to, limit)
}

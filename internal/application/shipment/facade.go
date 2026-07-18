package shipment

import (
	"context"
	"fmt"
	"time"

	appworkflow "github.com/alireza-akbarzadeh/luxe/internal/application/workflow"
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

// Notifier sends in-app notifications for shipment updates.
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

// Service orchestrates shipment creation, status updates, and provider management.
type Service struct {
	commands     *Commands
	queries      *Queries
	workerPool   asynq.JobQueue
	notifier     Notifier
	vendorNotify VendorStoreNotifier
	wsHub        *websocket.Hub
	engine       *workflow.Engine
}

// NewService wires shipment commands and queries.
func NewService(
	db *gorm.DB,
	workerPool asynq.JobQueue,
	notifier Notifier,
	vendorNotify VendorStoreNotifier,
	wsHub *websocket.Hub,
	engine *workflow.Engine,
) *Service {
	repo := postgres.NewShipmentRepository(db)
	return &Service{
		commands:     NewCommands(repo),
		queries:      NewQueries(repo),
		workerPool:   workerPool,
		notifier:     notifier,
		vendorNotify: vendorNotify,
		wsHub:        wsHub,
		engine:       engine,
	}
}

func (s *Service) setShipmentState(ctx context.Context, shipmentID uint, status string) {
	appworkflow.ApplyShipmentWorkflow(ctx, s.engine, shipmentID, status)
}

func (s *Service) broadcastShipmentUpdate(
	orderID, userID uint,
	eventType string,
	data map[string]interface{},
) {
	msg := websocket.Message{
		Type:      eventType,
		UserID:    userID,
		Data:      data,
		Timestamp: time.Now(),
	}
	roomID := fmt.Sprintf("order_%d", orderID)
	s.wsHub.BroadcastToRoom(roomID, msg)

	title, _ := data["title"].(string)
	message, _ := data["message"].(string)
	go func() {
		_ = s.notifier.CreateNotification(userID, eventType, title, message, data)
	}()
	if s.vendorNotify != nil {
		s.vendorNotify.NotifyVendorStoresForOrder(
			context.Background(),
			orderID,
			websocket.EventVendorOrderShipment,
			"vendor_order_shipment",
			title,
			message,
			data,
		)
	}
}

func (s *Service) toCreateInput(req CreateShipmentRequest) CreateInput {
	return CreateInput{
		OrderID:        req.OrderID,
		Carrier:        req.Carrier,
		TrackingNumber: req.TrackingNumber,
		AddressLine1:   req.AddressLine1,
		AddressLine2:   req.AddressLine2,
		City:           req.City,
		State:          req.State,
		PostalCode:     req.PostalCode,
		Country:        req.Country,
		ProviderID:     req.ProviderID,
		ShippingPrice:  req.ShippingPrice,
	}
}

func (s *Service) CreateShipment(req CreateShipmentRequest) (*models.Shipment, error) {
	ctx := context.Background()
	shipment, err := s.commands.Create(ctx, s.toCreateInput(req))
	if err != nil {
		return nil, err
	}

	if err := s.workerPool.EnqueueProcessShipment(ctx, shipment.ID); err != nil {
		utils.Log.WithError(err).WithField("shipment_id", shipment.ID).Error("failed to enqueue shipment processing job")
	}

	s.broadcastShipmentUpdate(shipment.OrderID, shipment.UserID, "shipment_created", map[string]interface{}{
		"title":           "Shipment Created",
		"message":         "Your order shipment has been created and is being prepared for delivery.",
		"shipment_id":     shipment.ID,
		"order_id":        shipment.OrderID,
		"carrier":         shipment.Carrier,
		"tracking_number": shipment.TrackingNumber,
		"status":          shipment.Status,
	})

	return shipment, nil
}

func (s *Service) ProcessShipmentBackground(ctx context.Context, shipmentID uint) error {
	time.Sleep(2 * time.Second)

	shipment, oldStatus, err := s.commands.MarkProcessing(ctx, shipmentID)
	if err != nil {
		return err
	}

	s.setShipmentState(ctx, shipmentID, "processing")
	utils.Log.Infof("Shipment %d processed in background", shipmentID)

	s.broadcastShipmentUpdate(shipment.OrderID, shipment.UserID, "shipment_status_update", map[string]interface{}{
		"title":       "Shipment Processing",
		"message":     "Your shipment is now being processed and prepared for delivery.",
		"shipment_id": shipment.ID,
		"order_id":    shipment.OrderID,
		"carrier":     shipment.Carrier,
		"old_status":  oldStatus,
		"new_status":  "processing",
	})

	return nil
}

func (s *Service) CreateShipmentRecord(tx *gorm.DB, req CreateShipmentRequest) (*models.Shipment, error) {
	return s.commands.CreateInTx(tx, s.toCreateInput(req))
}

func (s *Service) GetShipmentByID(id uint) (*models.Shipment, error) {
	return s.queries.GetByID(context.Background(), id)
}

func (s *Service) ListAdmin(ctx context.Context, filters dto.AdminShipmentListFilters) ([]models.Shipment, int64, error) {
	return s.queries.ListAdmin(ctx, filters)
}

func (s *Service) GetShipmentsByOrderID(orderID uint) ([]models.Shipment, error) {
	return s.queries.GetByOrderID(context.Background(), orderID)
}

func (s *Service) UpdateShipmentStatus(id uint, status string) error {
	ctx := context.Background()
	shipment, oldStatus, err := s.commands.UpdateStatus(ctx, id, status)
	if err != nil {
		return err
	}

	s.setShipmentState(ctx, id, status)

	title, message := s.getShipmentStatusNotificationMessage(status, shipment.TrackingNumber)
	s.broadcastShipmentUpdate(shipment.OrderID, shipment.UserID, "shipment_status_update", map[string]interface{}{
		"title":           title,
		"message":         message,
		"shipment_id":     shipment.ID,
		"order_id":        shipment.OrderID,
		"carrier":         shipment.Carrier,
		"tracking_number": shipment.TrackingNumber,
		"old_status":      oldStatus,
		"new_status":      status,
	})

	return nil
}

func (s *Service) AvailableTransitions(ctx context.Context, shipmentID uint) (*models.WorkflowState, []models.WorkflowTransition, error) {
	if s.engine == nil {
		return nil, nil, utils.ErrInternal(fmt.Errorf("workflow engine not configured"))
	}
	if _, err := s.GetShipmentByID(shipmentID); err != nil {
		return nil, nil, err
	}
	return s.engine.AvailableTransitions(ctx, constants.WorkflowEntityShipment, shipmentID)
}

func (s *Service) PerformTransition(
	ctx context.Context,
	shipmentID uint,
	event, note, actorRole string,
	actorID *uint,
) (*workflow.TransitionResult, error) {
	if s.engine == nil {
		return nil, utils.ErrInternal(fmt.Errorf("workflow engine not configured"))
	}
	if _, err := s.GetShipmentByID(shipmentID); err != nil {
		return nil, err
	}
	return s.engine.Transition(ctx, workflow.TransitionRequest{
		WorkflowKey: constants.WorkflowEntityShipment,
		EntityID:    shipmentID,
		Event:       event,
		ActorID:     actorID,
		ActorRole:   actorRole,
		Note:        note,
	})
}

func (s *Service) getShipmentStatusNotificationMessage(status, trackingNumber string) (string, string) {
	switch status {
	case constants.ShipmentStatusShipped:
		if trackingNumber != "" {
			return "Package Shipped", fmt.Sprintf("Your package has been shipped! Track it with: %s", trackingNumber)
		}
		return "Package Shipped", "Your package has been shipped and is on its way!"
	case constants.ShipmentStatusDelivered:
		return "Package Delivered", "Your package has been delivered successfully!"
	default:
		return "Shipment Update", fmt.Sprintf("Your shipment status has been updated to: %s", status)
	}
}

func (s *Service) GetShippingProviders() ([]models.ShippingProviders, error) {
	return s.queries.ListActiveProviders(context.Background())
}

func (s *Service) ListShippingProvidersAdmin(ctx context.Context) ([]models.ShippingProviders, error) {
	return s.queries.ListProvidersAdmin(ctx)
}

func (s *Service) GetShippingProviderByID(providerID uint) (*models.ShippingProviders, error) {
	return s.queries.GetProviderByID(context.Background(), providerID)
}

func (s *Service) DeleteShippingProvider(providerID uint) error {
	return s.commands.DeleteProvider(context.Background(), providerID)
}

func (s *Service) CreateShippingProvider(req dto.CreateShippingProviderRequest) (*models.ShippingProviders, error) {
	return s.commands.CreateProvider(context.Background(), CreateProviderInput{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		IsActive:    req.IsActive,
	})
}

func (s *Service) UpdateShippingProvider(providerID uint, req dto.UpdateShippingProviderRequest) (*models.ShippingProviders, error) {
	return s.commands.UpdateProvider(context.Background(), providerID, UpdateProviderInput{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		IsActive:    req.IsActive,
	})
}

func (s *Service) SimulateDeliveries() error {
	ctx := context.Background()
	shipments, err := s.commands.SimulateDeliveries(ctx)
	if err != nil {
		return err
	}

	for i := range shipments {
		shipment := shipments[i]
		oldStatus, err := s.commands.MarkDelivered(ctx, &shipment)
		if err != nil {
			utils.Log.WithError(err).Errorf("Failed to mark shipment %d as delivered", shipment.ID)
			continue
		}
		s.setShipmentState(ctx, shipment.ID, constants.ShipmentStatusDelivered)

		s.broadcastShipmentUpdate(shipment.OrderID, shipment.UserID, "shipment_delivered", map[string]interface{}{
			"title":           "Package Delivered",
			"message":         "Your package has been delivered successfully!",
			"shipment_id":     shipment.ID,
			"order_id":        shipment.OrderID,
			"tracking_number": shipment.TrackingNumber,
			"carrier":         shipment.Carrier,
			"old_status":      oldStatus,
			"new_status":      shipment.Status,
			"delivered_at":    shipment.DeliveredAt,
		})
	}

	return nil
}

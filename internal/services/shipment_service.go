package services

import (
	"context"
	"fmt"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	appshipment "github.com/alireza-akbarzadeh/luxe/internal/application/shipment"
	appworkflow "github.com/alireza-akbarzadeh/luxe/internal/application/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/asynq"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/alireza-akbarzadeh/luxe/internal/websocket"
	"gorm.io/gorm"
)

type CreateShipmentRequest struct {
	OrderID        uint    `json:"order_id" validate:"required,gt=0"`
	Carrier        string  `json:"carrier" validate:"required"`
	TrackingNumber string  `json:"tracking_number,omitempty"`
	AddressLine1   string  `json:"address_line1" validate:"required"`
	AddressLine2   string  `json:"address_line2,omitempty"`
	City           string  `json:"city" validate:"required"`
	State          string  `json:"state,omitempty"`
	PostalCode     string  `json:"postal_code" validate:"required"`
	Country        string  `json:"country" validate:"required"`
	ProviderID     *uint   `json:"provider_id,omitempty"`
	ShippingPrice  float64 `json:"shipping_price" validate:"gte=0"`
}

type ShipmentServiceInterface interface {
	CreateShipment(req CreateShipmentRequest) (*models.Shipment, error)
	GetShipmentByID(id uint) (*models.Shipment, error)
	GetShipmentsByOrderID(orderID uint) ([]models.Shipment, error)
	ListAdmin(ctx context.Context, filters dto.AdminShipmentListFilters) ([]models.Shipment, int64, error)
	UpdateShipmentStatus(id uint, status string) error
	AvailableTransitions(ctx context.Context, shipmentID uint) (*models.WorkflowState, []models.WorkflowTransition, error)
	PerformTransition(ctx context.Context, shipmentID uint, event, note, actorRole string, actorID *uint) (*workflow.TransitionResult, error)
	CreateShipmentRecord(tx *gorm.DB, req CreateShipmentRequest) (*models.Shipment, error)
	SimulateDeliveries() error

	DeleteShippingProvider(providerID uint) error
	GetShippingProviderByID(providerID uint) (*models.ShippingProviders, error)
	GetShippingProviders() ([]models.ShippingProviders, error)
	ListShippingProvidersAdmin(ctx context.Context) ([]models.ShippingProviders, error)
	CreateShippingProvider(req dto.CreateShippingProviderRequest) (*models.ShippingProviders, error)
	UpdateShippingProvider(providerID uint, req dto.UpdateShippingProviderRequest) (*models.ShippingProviders, error)
	ProcessShipmentBackground(ctx context.Context, shipmentID uint) error
}

type shipmentService struct {
	commands            *appshipment.Commands
	queries             *appshipment.Queries
	workerPool          asynq.JobQueue
	notificationService NotificationServiceInterface
	wsHub               *websocket.Hub
	engine              *workflow.Engine
}

func NewShipmentService(
	db *gorm.DB,
	workerPool asynq.JobQueue,
	notificationService NotificationServiceInterface,
	wsHub *websocket.Hub,
	engine *workflow.Engine,
) ShipmentServiceInterface {
	repo := postgres.NewShipmentRepository(db)
	return &shipmentService{
		commands:            appshipment.NewCommands(repo),
		queries:             appshipment.NewQueries(repo),
		workerPool:          workerPool,
		notificationService: notificationService,
		wsHub:               wsHub,
		engine:              engine,
	}
}

func (s *shipmentService) setShipmentState(ctx context.Context, shipmentID uint, status string) {
	appworkflow.ApplyShipmentWorkflow(ctx, s.engine, shipmentID, status)
}

func (s *shipmentService) broadcastShipmentUpdate(
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
		_ = s.notificationService.CreateNotification(userID, eventType, title, message, data)
	}()
}

func (s *shipmentService) toCreateInput(req CreateShipmentRequest) appshipment.CreateInput {
	return appshipment.CreateInput{
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

func (s *shipmentService) CreateShipment(req CreateShipmentRequest) (*models.Shipment, error) {
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

func (s *shipmentService) ProcessShipmentBackground(ctx context.Context, shipmentID uint) error {
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

func (s *shipmentService) CreateShipmentRecord(tx *gorm.DB, req CreateShipmentRequest) (*models.Shipment, error) {
	return s.commands.CreateInTx(tx, s.toCreateInput(req))
}

func (s *shipmentService) GetShipmentByID(id uint) (*models.Shipment, error) {
	return s.queries.GetByID(context.Background(), id)
}

func (s *shipmentService) ListAdmin(ctx context.Context, filters dto.AdminShipmentListFilters) ([]models.Shipment, int64, error) {
	return s.queries.ListAdmin(ctx, filters)
}

func (s *shipmentService) GetShipmentsByOrderID(orderID uint) ([]models.Shipment, error) {
	return s.queries.GetByOrderID(context.Background(), orderID)
}

func (s *shipmentService) UpdateShipmentStatus(id uint, status string) error {
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

func (s *shipmentService) AvailableTransitions(ctx context.Context, shipmentID uint) (*models.WorkflowState, []models.WorkflowTransition, error) {
	if s.engine == nil {
		return nil, nil, utils.ErrInternal(fmt.Errorf("workflow engine not configured"))
	}
	if _, err := s.GetShipmentByID(shipmentID); err != nil {
		return nil, nil, err
	}
	return s.engine.AvailableTransitions(ctx, constants.WorkflowEntityShipment, shipmentID)
}

func (s *shipmentService) PerformTransition(
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

func (s *shipmentService) getShipmentStatusNotificationMessage(status, trackingNumber string) (string, string) {
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

func (s *shipmentService) GetShippingProviders() ([]models.ShippingProviders, error) {
	return s.queries.ListActiveProviders(context.Background())
}

func (s *shipmentService) ListShippingProvidersAdmin(ctx context.Context) ([]models.ShippingProviders, error) {
	return s.queries.ListProvidersAdmin(ctx)
}

func (s *shipmentService) GetShippingProviderByID(providerID uint) (*models.ShippingProviders, error) {
	return s.queries.GetProviderByID(context.Background(), providerID)
}

func (s *shipmentService) DeleteShippingProvider(providerID uint) error {
	return s.commands.DeleteProvider(context.Background(), providerID)
}

func (s *shipmentService) CreateShippingProvider(req dto.CreateShippingProviderRequest) (*models.ShippingProviders, error) {
	return s.commands.CreateProvider(context.Background(), appshipment.CreateProviderInput{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		IsActive:    req.IsActive,
	})
}

func (s *shipmentService) UpdateShippingProvider(providerID uint, req dto.UpdateShippingProviderRequest) (*models.ShippingProviders, error) {
	return s.commands.UpdateProvider(context.Background(), providerID, appshipment.UpdateProviderInput{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		IsActive:    req.IsActive,
	})
}

func (s *shipmentService) SimulateDeliveries() error {
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

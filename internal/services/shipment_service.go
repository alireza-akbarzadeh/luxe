package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/services/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/tasks"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/alireza-akbarzadeh/luxe/internal/websocket"
	"gorm.io/gorm"
)

// shipmentService manages shipment records and carrier simulation.

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
	UpdateShipmentStatus(id uint, status string) error
	CreateShipmentRecord(tx *gorm.DB, req CreateShipmentRequest) (*models.Shipment, error)
	SimulateDeliveries() error

	DeleteShippingProvider(providerID uint) error
	GetShippingProviderByID(providerID uint) (*models.ShippingProviders, error)
	GetShippingProviders() ([]models.ShippingProviders, error)
	CreateShippingProvider(req dto.CreateShippingProviderRequest) (*models.ShippingProviders, error)
	UpdateShippingProvider(providerID uint, req dto.UpdateShippingProviderRequest) (*models.ShippingProviders, error)
	ProcessShipmentBackground(ctx context.Context, shipmentID uint) error
}

type shipmentService struct {
	db                  *gorm.DB
	workerPool          tasks.JobQueue
	notificationService NotificationServiceInterface
	wsHub               *websocket.Hub
	engine              *workflow.Engine
}

func NewShipmentService(
	db *gorm.DB,
	workerPool tasks.JobQueue,
	notificationService NotificationServiceInterface,
	wsHub *websocket.Hub,
	engine *workflow.Engine,
) ShipmentServiceInterface {
	return &shipmentService{
		db:                  db,
		workerPool:          workerPool,
		notificationService: notificationService,
		wsHub:               wsHub,
		engine:              engine,
	}
}

// setShipmentState syncs a shipment status into the workflow engine (best-effort).
func (s *shipmentService) setShipmentState(ctx context.Context, shipmentID uint, status string) {
	applyShipmentWorkflow(ctx, s.engine, shipmentID, status)
}

// ─── WebSocket + Notification helper ─────────────────────────────────────

// broadcastShipmentUpdate pushes a real‑time update to both the user's
// personal notification channel and the order‑specific WebSocket room.
func (s *shipmentService) broadcastShipmentUpdate(
	orderID, userID uint,
	eventType string,
	data map[string]interface{},
) {
	// 1. Send to the order room (frontend order‑tracking page listens here)
	msg := websocket.Message{
		Type:      eventType,
		UserID:    userID,
		Data:      data,
		Timestamp: time.Now(),
	}
	roomID := fmt.Sprintf("order_%d", orderID)
	s.wsHub.BroadcastToRoom(roomID, msg)

	// 2. Also store a persistent notification (delivers via personal WS channel too)
	title, _ := data["title"].(string)
	message, _ := data["message"].(string)
	go func() {
		_ = s.notificationService.CreateNotification(
			userID,
			eventType,
			title,
			message,
			data,
		)
	}()
}

// ─── CreateShipment (standalone, with its own background job) ────────────

func (s *shipmentService) CreateShipment(req CreateShipmentRequest) (*models.Shipment, error) {
	var order models.Order
	if err := s.db.First(&order, req.OrderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound(constants.ErrOrderNotFound)
		}
		return nil, utils.ErrInternal(err)
	}

	shipment := &models.Shipment{
		OrderID:        req.OrderID,
		UserID:         order.UserID,
		Carrier:        req.Carrier,
		TrackingNumber: req.TrackingNumber,
		Status:         constants.ShipmentStatusPending,
		AddressLine1:   req.AddressLine1,
		AddressLine2:   req.AddressLine2,
		City:           req.City,
		State:          req.State,
		PostalCode:     req.PostalCode,
		Country:        req.Country,
		ProviderID:     req.ProviderID,
		ShippingPrice:  req.ShippingPrice,
	}

	if err := s.db.Create(shipment).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	// Enqueue background job
	if err := s.workerPool.EnqueueProcessShipment(context.Background(), shipment.ID); err != nil {
		utils.Log.WithError(err).WithField("shipment_id", shipment.ID).Error("failed to enqueue shipment processing job")
	}

	// Broadcast creation event
	s.broadcastShipmentUpdate(order.ID, order.UserID, "shipment_created", map[string]interface{}{
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

// ProcessShipmentBackground runs async shipment processing (carrier simulation).
func (s *shipmentService) ProcessShipmentBackground(ctx context.Context, shipmentID uint) error {
	return s.processShipment(ctx, shipmentID)
}

// processShipment is the background job handler (standalone flow only).
func (s *shipmentService) processShipment(ctx context.Context, shipmentID uint) error {
	time.Sleep(2 * time.Second) // simulate carrier API

	var shipment models.Shipment
	if err := s.db.WithContext(ctx).First(&shipment, shipmentID).Error; err != nil {
		return err
	}

	oldStatus := shipment.Status

	if err := s.db.WithContext(ctx).Model(&models.Shipment{}).Where("id = ?", shipmentID).
		Update("status", "processing").Error; err != nil {
		return err
	}
	s.setShipmentState(ctx, shipmentID, "processing")

	utils.Log.Infof("Shipment %d processed in background", shipmentID)

	// Broadcast status change
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

// ─── CreateShipmentRecord (used inside checkout transaction) ─────────────

func (s *shipmentService) CreateShipmentRecord(tx *gorm.DB, req CreateShipmentRequest) (*models.Shipment, error) {
	var order models.Order
	if err := tx.First(&order, req.OrderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound(constants.ErrOrderNotFound)
		}
		return nil, utils.ErrInternal(err)
	}

	shipment := &models.Shipment{
		OrderID:        req.OrderID,
		UserID:         order.UserID,
		Carrier:        req.Carrier,
		TrackingNumber: req.TrackingNumber,
		Status:         constants.ShipmentStatusPending,
		AddressLine1:   req.AddressLine1,
		AddressLine2:   req.AddressLine2,
		City:           req.City,
		State:          req.State,
		PostalCode:     req.PostalCode,
		Country:        req.Country,
		ProviderID:     req.ProviderID,
		ShippingPrice:  req.ShippingPrice,
	}

	if err := tx.Create(shipment).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}
	return shipment, nil
}

// ─── Read helpers ────────────────────────────────────────────────────────

func (s *shipmentService) GetShipmentByID(id uint) (*models.Shipment, error) {
	var shipment models.Shipment
	if err := s.db.First(&shipment, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound(constants.ErrShipmentNotFound)
		}
		return nil, utils.ErrInternal(err)
	}
	return &shipment, nil
}

func (s *shipmentService) GetShipmentsByOrderID(orderID uint) ([]models.Shipment, error) {
	var shipments []models.Shipment
	if err := s.db.Where("order_id = ?", orderID).Find(&shipments).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}
	return shipments, nil
}

// ─── UpdateShipmentStatus (admin manual update) ──────────────────────────

func (s *shipmentService) UpdateShipmentStatus(id uint, status string) error {
	var shipment models.Shipment
	if err := s.db.First(&shipment, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrNotFound(constants.ErrShipmentNotFound)
		}
		return utils.ErrInternal(err)
	}

	oldStatus := shipment.Status

	result := s.db.Model(&models.Shipment{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return utils.ErrInternal(result.Error)
	}
	if result.RowsAffected == 0 {
		return utils.ErrNotFound(constants.ErrShipmentNotFound)
	}

	s.setShipmentState(context.Background(), id, status)

	// Broadcast status change
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

// ─── Shipping Providers CRUD ─────────────────────────────────────────────

func (s *shipmentService) GetShippingProviders() ([]models.ShippingProviders, error) {
	var providers []models.ShippingProviders
	err := s.db.Where("is_active = ?", true).Order("price ASC").Find(&providers).Error
	return providers, err
}

func (s *shipmentService) GetShippingProviderByID(providerID uint) (*models.ShippingProviders, error) {
	var provider models.ShippingProviders
	err := s.db.First(&provider, providerID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("shipping provider not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return &provider, nil
}

func (s *shipmentService) DeleteShippingProvider(providerID uint) error {
	result := s.db.Delete(models.ShippingProviders{}, providerID)
	if result.Error != nil {
		return utils.ErrInternal(result.Error)
	}
	if result.RowsAffected == 0 {
		return utils.ErrNotFound("shipping provider not found")
	}
	return nil
}

func (s *shipmentService) CreateShippingProvider(req dto.CreateShippingProviderRequest) (*models.ShippingProviders, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	provider := models.ShippingProviders{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		IsActive:    isActive,
	}
	if err := s.db.Create(&provider).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}
	return &provider, nil
}

func (s *shipmentService) UpdateShippingProvider(providerID uint, req dto.UpdateShippingProviderRequest) (*models.ShippingProviders, error) {
	var provider models.ShippingProviders
	if err := s.db.First(&provider, providerID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("shipping provider not found")
		}
		return nil, utils.ErrInternal(err)
	}
	if req.Name != nil {
		provider.Name = *req.Name
	}
	if req.Description != nil {
		provider.Description = *req.Description
	}
	if req.Price != nil {
		provider.Price = *req.Price
	}
	if req.IsActive != nil {
		provider.IsActive = *req.IsActive
	}
	if err := s.db.Save(&provider).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}
	return &provider, nil
}

// ─── Cron: Simulate deliveries ───────────────────────────────────────────

// SimulateDeliveries marks shipped shipments as delivered after 24 hours.
// Call this from a cron job (e.g., every 5 minutes).
func (s *shipmentService) SimulateDeliveries() error {

	cutoff := time.Now().Add(-5 * time.Minute)
	var shipments []models.Shipment
	if err := s.db.Where("status = ? AND shipped_at <= ?", constants.ShipmentStatusShipped, cutoff).
		Find(&shipments).Error; err != nil {
		return utils.ErrInternal(err)
	}

	for _, shipment := range shipments {
		oldStatus := shipment.Status
		shipment.Status = constants.ShipmentStatusDelivered

		// Correct assignment: DeliveredAt expects *time.Time
		now := time.Now()
		shipment.DeliveredAt = &now

		if err := s.db.Save(&shipment).Error; err != nil {
			utils.Log.WithError(err).Errorf("Failed to mark shipment %d as delivered", shipment.ID)
			continue
		}
		s.setShipmentState(context.Background(), shipment.ID, constants.ShipmentStatusDelivered)

		// Broadcast delivery event to the order room
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

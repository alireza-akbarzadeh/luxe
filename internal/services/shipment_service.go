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
	"gorm.io/gorm"
)

type CreateShipmentRequest struct {
	OrderID        uint   `json:"order_id" validate:"required,gt=0"`
	Carrier        string `json:"carrier" validate:"required"`
	TrackingNumber string `json:"tracking_number,omitempty"`
	AddressLine1   string `json:"address_line1" validate:"required"`
	AddressLine2   string `json:"address_line2,omitempty"`
	City           string `json:"city" validate:"required"`
	State          string `json:"state,omitempty"`
	PostalCode     string `json:"postal_code" validate:"required"`
	Country        string `json:"country" validate:"required"`
}

type ShipmentServiceInterface interface {
	CreateShipment(req CreateShipmentRequest) (*models.Shipment, error)
	GetShipmentByID(id uint) (*models.Shipment, error)
	GetShipmentsByOrderID(orderID uint) ([]models.Shipment, error)
	UpdateShipmentStatus(id uint, status string) error

	DeleteShippingProvider(providerId uint) error
	GetShippingProviderByID(providerId uint) (*models.ShippingProviders, error)
	GetShippingProviders() ([]models.ShippingProviders, error)
	CreateShippingProvider(req dto.CreateShippingProviderRequest) (*models.ShippingProviders, error)
	UpdateShippingProvider(providerId uint, req dto.UpdateShippingProviderRequest) (*models.ShippingProviders, error)
}

type shipmentService struct {
	db                  *gorm.DB
	workerPool          *tasks.WorkerPool
	notificationService NotificationServiceInterface
}

func NewShipmentService(db *gorm.DB, workerPool *tasks.WorkerPool, notificationService NotificationServiceInterface) ShipmentServiceInterface {
	return &shipmentService{
		db:                  db,
		workerPool:          workerPool,
		notificationService: notificationService,
	}
}

// CreateShipment creates a shipment record and enqueues a background job.
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
	}

	if err := s.db.Create(shipment).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	// Enqueue background job (e.g., call carrier API, generate label)
	job := tasks.Job{
		ID:      fmt.Sprintf("shipment_%d", shipment.ID),
		Payload: shipment.ID,
		Handler: s.processShipment,
	}
	s.workerPool.Enqueue(job)

	// Send real-time notification for shipment creation
	go func() {
		_ = s.notificationService.CreateNotification(
			shipment.UserID,
			"shipment_created",
			"Shipment Created",
			fmt.Sprintf("Your order shipment has been created and is being prepared for delivery."),
			map[string]interface{}{
				"shipment_id":     shipment.ID,
				"order_id":        shipment.OrderID,
				"carrier":         shipment.Carrier,
				"tracking_number": shipment.TrackingNumber,
				"status":          shipment.Status,
			},
		)
	}()

	return shipment, nil
}

// processShipment is the job handler called by the worker pool.
func (s *shipmentService) processShipment(payload interface{}) error {
	shipmentID, ok := payload.(uint)
	if !ok {
		return fmt.Errorf("invalid payload type")
	}

	// Simulate external carrier API call (label generation, tracking update)
	time.Sleep(2 * time.Second)

	// Get shipment before update
	var shipment models.Shipment
	if err := s.db.First(&shipment, shipmentID).Error; err != nil {
		return err
	}

	oldStatus := shipment.Status

	// Update status to 'processing' (or 'shipped' after successful API call)
	if err := s.db.Model(&models.Shipment{}).Where("id = ?", shipmentID).
		Update("status", "processing").Error; err != nil {
		return err
	}

	utils.Log.Infof("Shipment %d processed in background", shipmentID)

	// Send real-time notification for status change
	go func() {
		_ = s.notificationService.CreateNotification(
			shipment.UserID,
			"shipment_status_update",
			"Shipment Processing",
			fmt.Sprintf("Your shipment is now being processed and prepared for delivery."),
			map[string]interface{}{
				"shipment_id":     shipment.ID,
				"order_id":        shipment.OrderID,
				"carrier":         shipment.Carrier,
				"tracking_number": shipment.TrackingNumber,
				"old_status":      oldStatus,
				"new_status":      "processing",
			},
		)
	}()

	return nil
}

// GetShipmentByID retrieves a shipment by its ID.
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

// GetShipmentsByOrderID returns all shipments associated with an order.
func (s *shipmentService) GetShipmentsByOrderID(orderID uint) ([]models.Shipment, error) {
	var shipments []models.Shipment
	if err := s.db.Where("order_id = ?", orderID).Find(&shipments).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}
	return shipments, nil
}

// UpdateShipmentStatus manually updates a shipment's status (admin only).
func (s *shipmentService) UpdateShipmentStatus(id uint, status string) error {
	var shipment models.Shipment
	if err := s.db.First(&shipment, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrNotFound(constants.ErrShipmentNotFound)
		}
		return utils.ErrInternal(err)
	}

	oldStatus := shipment.Status
	shipment.Status = status

	result := s.db.Model(&models.Shipment{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return utils.ErrInternal(result.Error)
	}
	if result.RowsAffected == 0 {
		return utils.ErrNotFound(constants.ErrShipmentNotFound)
	}

	// Send real-time notification for status change
	go func() {
		title, message := s.getShipmentStatusNotificationMessage(status, shipment.TrackingNumber)
		_ = s.notificationService.CreateNotification(
			shipment.UserID,
			"shipment_status_update",
			title,
			message,
			map[string]interface{}{
				"shipment_id":     shipment.ID,
				"order_id":        shipment.OrderID,
				"carrier":         shipment.Carrier,
				"tracking_number": shipment.TrackingNumber,
				"old_status":      oldStatus,
				"new_status":      status,
			},
		)
	}()

	return nil
}

// getShipmentStatusNotificationMessage returns appropriate title and message for shipment status
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

// GetShippingProviders getting the all the services
func (s *shipmentService) GetShippingProviders() ([]models.ShippingProviders, error) {
	var providers []models.ShippingProviders
	err := s.db.Where("is_active = ?", true).Order("price ASC").Find(&providers).Error
	return providers, err
}

// GetShippingProviderByID find the provider with the given id
func (s *shipmentService) GetShippingProviderByID(providerId uint) (*models.ShippingProviders, error) {
	var provider models.ShippingProviders
	err := s.db.First(&provider, providerId).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("shipping provider not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return &provider, nil
}

// DeleteShippingProvider remove  provider with the given id
func (s *shipmentService) DeleteShippingProvider(providerId uint) error {
	result := s.db.Delete(models.ShippingProviders{}, providerId)
	if result.Error != nil {
		return utils.ErrInternal(result.Error)
	}
	if result.RowsAffected == 0 {
		return utils.ErrNotFound("product not found")
	}
	return nil
}

// CreateShippingProvider creates a new shipping provider
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
	err := s.db.Create(&provider).Error
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	return &provider, nil
}

// UpdateShippingProvider updates an existing shipping provider by ID
func (s *shipmentService) UpdateShippingProvider(providerId uint, req dto.UpdateShippingProviderRequest) (*models.ShippingProviders, error) {

	var provider models.ShippingProviders
	err := s.db.First(&provider, providerId).Error
	if err != nil {
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
	err = s.db.Save(&provider).Error
	if err != nil {
	}
	return nil, utils.ErrInternal(err)
	return &provider, nil
}

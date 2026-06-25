package shipment

import (
	"context"
	"errors"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

// CreateInput is the application-layer shipment creation payload.
type CreateInput struct {
	OrderID        uint
	Carrier        string
	TrackingNumber string
	AddressLine1   string
	AddressLine2   string
	City           string
	State          string
	PostalCode     string
	Country        string
	ProviderID     *uint
	ShippingPrice  float64
}

// Commands orchestrates shipment write use cases.
type Commands struct {
	repo *postgres.ShipmentRepository
}

// NewCommands creates shipment command use cases.
func NewCommands(repo *postgres.ShipmentRepository) *Commands {
	return &Commands{repo: repo}
}

func (c *Commands) buildShipment(order *models.Order, in CreateInput) *models.Shipment {
	return &models.Shipment{
		OrderID:        in.OrderID,
		UserID:         order.UserID,
		Carrier:        in.Carrier,
		TrackingNumber: in.TrackingNumber,
		Status:         constants.ShipmentStatusPending,
		AddressLine1:   in.AddressLine1,
		AddressLine2:   in.AddressLine2,
		City:           in.City,
		State:          in.State,
		PostalCode:     in.PostalCode,
		Country:        in.Country,
		ProviderID:     in.ProviderID,
		ShippingPrice:  in.ShippingPrice,
	}
}

func (c *Commands) loadOrder(ctx context.Context, orderID uint) (*models.Order, error) {
	order, err := c.repo.FindOrderByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound(constants.ErrOrderNotFound)
		}
		return nil, utils.ErrInternal(err)
	}
	return order, nil
}

// Create inserts a standalone shipment record.
func (c *Commands) Create(ctx context.Context, in CreateInput) (*models.Shipment, error) {
	order, err := c.loadOrder(ctx, in.OrderID)
	if err != nil {
		return nil, err
	}
	shipment := c.buildShipment(order, in)
	if err := c.repo.Create(ctx, shipment); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return shipment, nil
}

// CreateInTx inserts a shipment inside checkout transaction.
func (c *Commands) CreateInTx(tx *gorm.DB, in CreateInput) (*models.Shipment, error) {
	order, err := c.repo.FindOrderByIDTx(tx, in.OrderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound(constants.ErrOrderNotFound)
		}
		return nil, utils.ErrInternal(err)
	}
	shipment := c.buildShipment(order, in)
	if err := c.repo.CreateInTx(tx, shipment); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return shipment, nil
}

// MarkProcessing sets shipment status to processing.
func (c *Commands) MarkProcessing(ctx context.Context, shipmentID uint) (*models.Shipment, string, error) {
	shipment, err := c.repo.FindByIDPlain(ctx, shipmentID)
	if err != nil {
		return nil, "", err
	}
	oldStatus := shipment.Status
	_, err = c.repo.UpdateStatus(ctx, shipmentID, "processing")
	if err != nil {
		return nil, "", err
	}
	shipment.Status = "processing"
	return shipment, oldStatus, nil
}

// UpdateStatus updates shipment status by id.
func (c *Commands) UpdateStatus(ctx context.Context, id uint, status string) (*models.Shipment, string, error) {
	shipment, err := c.repo.FindByIDPlain(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", utils.ErrNotFound(constants.ErrShipmentNotFound)
		}
		return nil, "", utils.ErrInternal(err)
	}
	oldStatus := shipment.Status
	rows, err := c.repo.UpdateStatus(ctx, id, status)
	if err != nil {
		return nil, "", utils.ErrInternal(err)
	}
	if rows == 0 {
		return nil, "", utils.ErrNotFound(constants.ErrShipmentNotFound)
	}
	shipment.Status = status
	return shipment, oldStatus, nil
}

// MarkDelivered marks a shipment as delivered with timestamp.
func (c *Commands) MarkDelivered(ctx context.Context, shipment *models.Shipment) (string, error) {
	oldStatus := shipment.Status
	shipment.Status = constants.ShipmentStatusDelivered
	now := time.Now()
	shipment.DeliveredAt = &now
	if err := c.repo.Save(ctx, shipment); err != nil {
		return "", err
	}
	return oldStatus, nil
}

// SimulateDeliveries loads shipped shipments ready for delivery simulation.
func (c *Commands) SimulateDeliveries(ctx context.Context) ([]models.Shipment, error) {
	cutoff := time.Now().Add(-5 * time.Minute)
	shipments, err := c.repo.FindShippedBefore(ctx, cutoff)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	return shipments, nil
}

// SaveDelivered persists a delivered shipment.
func (c *Commands) SaveDelivered(ctx context.Context, shipment *models.Shipment) error {
	if err := c.repo.Save(ctx, shipment); err != nil {
		return utils.ErrInternal(err)
	}
	return nil
}

// CreateProviderInput is the shipping provider creation payload.
type CreateProviderInput struct {
	Name        string
	Description string
	Price       float64
	IsActive    *bool
}

// CreateProvider inserts a shipping provider.
func (c *Commands) CreateProvider(ctx context.Context, in CreateProviderInput) (*models.ShippingProviders, error) {
	isActive := true
	if in.IsActive != nil {
		isActive = *in.IsActive
	}
	provider := &models.ShippingProviders{
		Name:        in.Name,
		Description: in.Description,
		Price:       in.Price,
		IsActive:    isActive,
	}
	if err := c.repo.CreateProvider(ctx, provider); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return provider, nil
}

// UpdateProviderInput is the shipping provider update payload.
type UpdateProviderInput struct {
	Name        *string
	Description *string
	Price       *float64
	IsActive    *bool
}

// UpdateProvider updates a shipping provider.
func (c *Commands) UpdateProvider(ctx context.Context, providerID uint, in UpdateProviderInput) (*models.ShippingProviders, error) {
	provider, err := c.repo.FindProviderByID(ctx, providerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("shipping provider not found")
		}
		return nil, utils.ErrInternal(err)
	}
	if in.Name != nil {
		provider.Name = *in.Name
	}
	if in.Description != nil {
		provider.Description = *in.Description
	}
	if in.Price != nil {
		provider.Price = *in.Price
	}
	if in.IsActive != nil {
		provider.IsActive = *in.IsActive
	}
	if err := c.repo.SaveProvider(ctx, provider); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return provider, nil
}

// DeleteProvider removes a shipping provider.
func (c *Commands) DeleteProvider(ctx context.Context, providerID uint) error {
	rows, err := c.repo.DeleteProvider(ctx, providerID)
	if err != nil {
		return utils.ErrInternal(err)
	}
	if rows == 0 {
		return utils.ErrNotFound("shipping provider not found")
	}
	return nil
}

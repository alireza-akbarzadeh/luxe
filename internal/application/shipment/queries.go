package shipment

import (
	"context"
	"errors"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

// Queries orchestrates shipment read use cases.
type Queries struct {
	repo *postgres.ShipmentRepository
}

// NewQueries creates shipment query use cases.
func NewQueries(repo *postgres.ShipmentRepository) *Queries {
	return &Queries{repo: repo}
}

// GetByID loads a shipment with relations.
func (q *Queries) GetByID(ctx context.Context, id uint) (*models.Shipment, error) {
	shipment, err := q.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound(constants.ErrShipmentNotFound)
		}
		return nil, utils.ErrInternal(err)
	}
	return shipment, nil
}

// ListAdmin returns paginated shipments for admin.
func (q *Queries) ListAdmin(ctx context.Context, filters dto.AdminShipmentListFilters) ([]models.Shipment, int64, error) {
	limit, offset := filters.Limit, filters.Offset
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	total, err := q.repo.CountAdmin(ctx, filters)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}

	shipments, err := q.repo.ListAdmin(ctx, filters, limit, offset)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	return shipments, total, nil
}

// GetByOrderID returns shipments for an order.
func (q *Queries) GetByOrderID(ctx context.Context, orderID uint) ([]models.Shipment, error) {
	shipments, err := q.repo.FindByOrderID(ctx, orderID)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	return shipments, nil
}

// ListActiveProviders returns active shipping providers.
func (q *Queries) ListActiveProviders(ctx context.Context) ([]models.ShippingProviders, error) {
	return q.repo.ListActiveProviders(ctx)
}

// ListProvidersAdmin returns all shipping providers for admin.
func (q *Queries) ListProvidersAdmin(ctx context.Context) ([]models.ShippingProviders, error) {
	providers, err := q.repo.ListAllProviders(ctx)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	return providers, nil
}

// GetProviderByID loads a shipping provider.
func (q *Queries) GetProviderByID(ctx context.Context, providerID uint) (*models.ShippingProviders, error) {
	provider, err := q.repo.FindProviderByID(ctx, providerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("shipping provider not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return provider, nil
}

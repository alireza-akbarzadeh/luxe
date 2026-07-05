package postgres

import (
	"context"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// ShipmentRepository persists shipments and shipping providers with GORM.
type ShipmentRepository struct {
	db *gorm.DB
}

// NewShipmentRepository creates a GORM-backed shipment repository.
func NewShipmentRepository(db *gorm.DB) *ShipmentRepository {
	return &ShipmentRepository{db: db}
}

// DB returns the underlying database handle for transaction-scoped writes.
func (r *ShipmentRepository) DB() *gorm.DB {
	return r.db
}

// FindOrderByID loads an order by primary key.
func (r *ShipmentRepository) FindOrderByID(ctx context.Context, orderID uint) (*models.Order, error) {
	var order models.Order
	if err := r.db.WithContext(ctx).First(&order, orderID).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// FindOrderByIDTx loads an order inside an existing transaction.
func (r *ShipmentRepository) FindOrderByIDTx(tx *gorm.DB, orderID uint) (*models.Order, error) {
	var order models.Order
	if err := tx.First(&order, orderID).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// Create inserts a shipment record.
func (r *ShipmentRepository) Create(ctx context.Context, shipment *models.Shipment) error {
	return r.db.WithContext(ctx).Create(shipment).Error
}

// CreateInTx inserts a shipment inside an existing transaction (checkout).
func (r *ShipmentRepository) CreateInTx(tx *gorm.DB, shipment *models.Shipment) error {
	return tx.Create(shipment).Error
}

// FindByID loads a shipment with relations.
func (r *ShipmentRepository) FindByID(ctx context.Context, id uint) (*models.Shipment, error) {
	var shipment models.Shipment
	err := r.db.WithContext(ctx).
		Preload("Order").
		Preload("Provider").
		Preload("WorkflowState").
		First(&shipment, id).Error
	if err != nil {
		return nil, err
	}
	return &shipment, nil
}

// FindByIDPlain loads a shipment without preloads.
func (r *ShipmentRepository) FindByIDPlain(ctx context.Context, id uint) (*models.Shipment, error) {
	var shipment models.Shipment
	if err := r.db.WithContext(ctx).First(&shipment, id).Error; err != nil {
		return nil, err
	}
	return &shipment, nil
}

func (r *ShipmentRepository) applyAdminFilters(q *gorm.DB, filters dto.AdminShipmentListFilters) *gorm.DB {
	if filters.Status != "" {
		q = q.Where("shipments.status = ?", filters.Status)
	}
	if filters.Carrier != "" {
		q = q.Where("shipments.carrier ILIKE ?", filters.Carrier)
	}
	if filters.OrderID != nil {
		q = q.Where("shipments.order_id = ?", *filters.OrderID)
	}
	if filters.Search != "" {
		term := "%" + filters.Search + "%"
		q = q.Joins("LEFT JOIN orders ON orders.id = shipments.order_id").
			Where(
				"shipments.tracking_number ILIKE ? OR orders.order_number ILIKE ? OR shipments.carrier ILIKE ?",
				term, term, term,
			)
	}
	return q
}

// CountAdmin counts shipments matching admin filters.
func (r *ShipmentRepository) CountAdmin(ctx context.Context, filters dto.AdminShipmentListFilters) (int64, error) {
	q := r.applyAdminFilters(r.db.WithContext(ctx).Model(&models.Shipment{}), filters)
	var total int64
	err := q.Count(&total).Error
	return total, err
}

// ListAdmin returns paginated shipments for admin.
func (r *ShipmentRepository) ListAdmin(ctx context.Context, filters dto.AdminShipmentListFilters, limit, offset int) ([]models.Shipment, error) {
	q := r.applyAdminFilters(r.db.WithContext(ctx).Model(&models.Shipment{}), filters)
	var shipments []models.Shipment
	err := q.
		Preload("Order").
		Preload("User").
		Preload("WorkflowState").
		Order("shipments.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&shipments).Error
	return shipments, err
}

// FindByOrderID returns shipments for an order.
func (r *ShipmentRepository) FindByOrderID(ctx context.Context, orderID uint) ([]models.Shipment, error) {
	var shipments []models.Shipment
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Find(&shipments).Error
	return shipments, err
}

// UpdateStatus sets shipment status by id.
func (r *ShipmentRepository) UpdateStatus(ctx context.Context, id uint, status string) (int64, error) {
	result := r.db.WithContext(ctx).Model(&models.Shipment{}).Where("id = ?", id).Update("status", status)
	return result.RowsAffected, result.Error
}

// FindShippedBefore returns shipped shipments older than cutoff.
func (r *ShipmentRepository) FindShippedBefore(ctx context.Context, cutoff time.Time) ([]models.Shipment, error) {
	var shipments []models.Shipment
	err := r.db.WithContext(ctx).
		Where("status = ? AND shipped_at <= ?", constants.ShipmentStatusShipped, cutoff).
		Find(&shipments).Error
	return shipments, err
}

// Save persists shipment changes.
func (r *ShipmentRepository) Save(ctx context.Context, shipment *models.Shipment) error {
	return r.db.WithContext(ctx).Save(shipment).Error
}

// ListActiveProviders returns active shipping providers ordered by price.
func (r *ShipmentRepository) ListActiveProviders(ctx context.Context) ([]models.ShippingProviders, error) {
	var providers []models.ShippingProviders
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Order("price ASC").Find(&providers).Error
	return providers, err
}

// ListAllProviders returns all shipping providers for admin.
func (r *ShipmentRepository) ListAllProviders(ctx context.Context) ([]models.ShippingProviders, error) {
	var providers []models.ShippingProviders
	err := r.db.WithContext(ctx).Order("name ASC, id ASC").Find(&providers).Error
	return providers, err
}

// FindProviderByID loads a shipping provider.
func (r *ShipmentRepository) FindProviderByID(ctx context.Context, id uint) (*models.ShippingProviders, error) {
	var provider models.ShippingProviders
	if err := r.db.WithContext(ctx).First(&provider, id).Error; err != nil {
		return nil, err
	}
	return &provider, nil
}

// CreateProvider inserts a shipping provider.
func (r *ShipmentRepository) CreateProvider(ctx context.Context, provider *models.ShippingProviders) error {
	return r.db.WithContext(ctx).Create(provider).Error
}

// SaveProvider persists provider changes.
func (r *ShipmentRepository) SaveProvider(ctx context.Context, provider *models.ShippingProviders) error {
	return r.db.WithContext(ctx).Save(provider).Error
}

// DeleteProvider removes a shipping provider by id.
func (r *ShipmentRepository) DeleteProvider(ctx context.Context, id uint) (int64, error) {
	result := r.db.WithContext(ctx).Delete(&models.ShippingProviders{}, id)
	return result.RowsAffected, result.Error
}

// StoreDeliveryStats aggregates delivered shipment timing for a store's catalog.
func (r *ShipmentRepository) StoreDeliveryStats(ctx context.Context, storeID uint) (deliveredCount int64, avgDays float64, err error) {
	if storeID == 0 {
		return 0, 0, nil
	}

	type row struct {
		Count   int64
		AvgDays *float64
	}
	var stats row

	err = r.db.WithContext(ctx).Model(&models.Shipment{}).
		Select(`COUNT(DISTINCT shipments.id) AS count,
			AVG(EXTRACT(EPOCH FROM (shipments.delivered_at - shipments.shipped_at)) / 86400) AS avg_days`).
		Joins("JOIN orders ON orders.id = shipments.order_id AND orders.deleted_at IS NULL").
		Joins("JOIN order_items ON order_items.order_id = orders.id AND order_items.deleted_at IS NULL").
		Joins("JOIN products ON products.id = order_items.product_id AND products.deleted_at IS NULL").
		Where(
			"products.store_id = ? AND shipments.shipped_at IS NOT NULL AND shipments.delivered_at IS NOT NULL AND shipments.deleted_at IS NULL",
			storeID,
		).
		Scan(&stats).Error
	if err != nil {
		return 0, 0, err
	}

	if stats.AvgDays != nil {
		avgDays = *stats.AvgDays
	}

	return stats.Count, avgDays, nil
}

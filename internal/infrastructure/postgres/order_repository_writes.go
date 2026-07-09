package postgres

import (
	"context"
	"strings"
	"time"

	apporder "github.com/alireza-akbarzadeh/luxe/internal/application/order"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// List implements apporder.Reader.
func (r *OrderRepository) List(ctx context.Context, filter apporder.ListFilter) ([]models.Order, int64, error) {
	q := OrderListQuery{
		UserID:         filter.UserID,
		StoreID:        filter.StoreID,
		Status:         filter.Status,
		PaymentStatus:  filter.PaymentStatus,
		ShipmentStatus: filter.ShipmentStatus,
		Tag:            filter.Tag,
		Search:         filter.Search,
		FromDate:       filter.FromDate,
		ToDate:         filter.ToDate,
		MinAmount:      filter.MinAmount,
		MaxAmount:      filter.MaxAmount,
		Limit:          filter.Limit,
		Offset:         filter.Offset,
		PreloadUser:    filter.PreloadUser,
	}
	return r.list(ctx, q)
}

func (r *OrderRepository) list(ctx context.Context, q OrderListQuery) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	countQuery := r.applyListFilters(r.db.WithContext(ctx).Model(&models.Order{}), q)
	if q.StoreID != nil {
		if err := countQuery.Select("COUNT(DISTINCT orders.id)").Scan(&total).Error; err != nil {
			return nil, 0, err
		}
	} else if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	listQuery := r.applyListFilters(r.db.WithContext(ctx).Model(&models.Order{}), q)
	if q.StoreID != nil {
		listQuery = listQuery.Distinct("orders.id")
	}
	listQuery = listQuery.Limit(q.Limit).Offset(q.Offset).
		Preload("Items.Product").
		Preload("Payment").
		Preload("Shipment").
		Preload("WorkflowState").
		Preload("Tags", func(db *gorm.DB) *gorm.DB {
			return db.Order("tag ASC")
		}).
		Order("orders.created_at DESC")

	if q.PreloadUser {
		listQuery = listQuery.Preload("User")
	}

	if err := listQuery.Find(&orders).Error; err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

// UpdateStatus sets order.status for a single order.
func (r *OrderRepository) UpdateStatus(ctx context.Context, orderID uint, status string) error {
	return r.db.WithContext(ctx).Model(&models.Order{}).Where("id = ?", orderID).Update("status", status).Error
}

// UpdateStatusByIDs sets order.status for multiple orders.
func (r *OrderRepository) UpdateStatusByIDs(ctx context.Context, orderIDs []uint, status string) (int64, error) {
	if len(orderIDs) == 0 {
		return 0, nil
	}
	res := r.db.WithContext(ctx).Model(&models.Order{}).Where("id IN ?", orderIDs).Update("status", status)
	return res.RowsAffected, res.Error
}

// FindOverduePaid returns paid orders past cutoff excluding terminal statuses.
func (r *OrderRepository) FindOverduePaid(ctx context.Context, cutoff time.Time, excludedStatuses []string) ([]models.Order, error) {
	var orders []models.Order
	err := r.db.WithContext(ctx).
		Where("status = ? AND updated_at < ?", constants.OrderStatusPaid, cutoff).
		Not("status IN ?", excludedStatuses).
		Find(&orders).Error
	return orders, err
}

// Save persists order field changes.
func (r *OrderRepository) Save(ctx context.Context, order *models.Order) error {
	return r.db.WithContext(ctx).Save(order).Error
}

// UpdateNotes sets the notes field on an order.
func (r *OrderRepository) UpdateNotes(ctx context.Context, orderID uint, notes string) error {
	return r.db.WithContext(ctx).Model(&models.Order{}).Where("id = ?", orderID).Update("notes", notes).Error
}

// ReplaceTags replaces all tags on an order.
func (r *OrderRepository) ReplaceTags(ctx context.Context, orderID uint, tags []string) error {
	normalized := normalizeOrderTags(tags)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("order_id = ?", orderID).Delete(&models.OrderTag{}).Error; err != nil {
			return err
		}
		if len(normalized) == 0 {
			return nil
		}
		rows := make([]models.OrderTag, len(normalized))
		for i, tag := range normalized {
			rows[i] = models.OrderTag{OrderID: orderID, Tag: tag}
		}
		return tx.Create(&rows).Error
	})
}

func normalizeOrderTags(tags []string) []string {
	seen := make(map[string]struct{}, len(tags))
	out := make([]string, 0, len(tags))
	for _, raw := range tags {
		tag := strings.TrimSpace(strings.ToLower(raw))
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	return out
}

// UpdateShipmentByOrderID updates shipment columns for the given order.
func (r *OrderRepository) UpdateShipmentByOrderID(ctx context.Context, orderID uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&models.Shipment{}).Where("order_id = ?", orderID).Updates(updates).Error
}

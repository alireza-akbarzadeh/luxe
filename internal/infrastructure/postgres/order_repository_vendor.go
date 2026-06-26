package postgres

import (
	"context"

	apporder "github.com/alireza-akbarzadeh/luxe/internal/application/order"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

type storeStatusCountRow struct {
	Status string
	Count  int64
}

// CountByStoreStatus returns total orders and counts grouped by status for a store.
func (r *OrderRepository) CountByStoreStatus(ctx context.Context, storeID uint) (apporder.VendorOrderStats, error) {
	stats := apporder.VendorOrderStats{ByStatus: map[string]int64{}}

	base := r.applyStoreScope(r.db.WithContext(ctx).Model(&models.Order{}), storeID)

	if err := base.Select("COUNT(DISTINCT orders.id)").Scan(&stats.Total).Error; err != nil {
		return stats, err
	}

	var rows []storeStatusCountRow
	err := r.applyStoreScope(r.db.WithContext(ctx).Model(&models.Order{}), storeID).
		Select("orders.status, COUNT(DISTINCT orders.id) AS count").
		Group("orders.status").
		Scan(&rows).Error
	if err != nil {
		return stats, err
	}

	for _, row := range rows {
		stats.ByStatus[row.Status] = row.Count
	}
	return stats, nil
}

// ListStoreIDsForOrder returns distinct store ids linked to an order via line items.
func (r *OrderRepository) ListStoreIDsForOrder(ctx context.Context, orderID uint) ([]uint, error) {
	var storeIDs []uint
	err := r.db.WithContext(ctx).
		Model(&models.OrderItem{}).
		Joins("JOIN products ON products.id = order_items.product_id AND products.deleted_at IS NULL").
		Where("order_items.order_id = ? AND order_items.deleted_at IS NULL AND products.store_id IS NOT NULL", orderID).
		Distinct("products.store_id").
		Pluck("products.store_id", &storeIDs).Error
	return storeIDs, err
}

// OrderBelongsToStore checks whether an order contains line items from the store.
func (r *OrderRepository) OrderBelongsToStore(ctx context.Context, orderID, storeID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Order{}).
		Joins("JOIN order_items ON order_items.order_id = orders.id AND order_items.deleted_at IS NULL").
		Joins("JOIN products ON products.id = order_items.product_id AND products.deleted_at IS NULL").
		Where("orders.id = ? AND products.store_id = ?", orderID, storeID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ListStoreIDsForOrder returns distinct store ids linked to an order via line items.
func (r *OrderRepository) ListStoreIDsForOrder(ctx context.Context, orderID uint) ([]uint, error) {
	var storeIDs []uint
	err := r.db.WithContext(ctx).
		Model(&models.OrderItem{}).
		Joins("JOIN products ON products.id = order_items.product_id AND products.deleted_at IS NULL").
		Where("order_items.order_id = ? AND order_items.deleted_at IS NULL AND products.store_id IS NOT NULL", orderID).
		Distinct("products.store_id").
		Pluck("products.store_id", &storeIDs).Error
	return storeIDs, err
}

package postgres

import (
	"context"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

type productStatusCountRow struct {
	Status string
	Count  int64
}

// CountByStoreStatus returns product totals, status breakdown, and low-stock count for a store.
func (r *ProductRepository) CountByStoreStatus(ctx context.Context, storeID uint) (dto.VendorProductStats, error) {
	stats := dto.VendorProductStats{ByStatus: map[string]int64{}}

	base := r.db.WithContext(ctx).Model(&models.Product{}).Where("store_id = ?", storeID)

	if err := base.Count(&stats.Total).Error; err != nil {
		return stats, err
	}

	var rows []productStatusCountRow
	err := r.db.WithContext(ctx).Model(&models.Product{}).
		Select("status, COUNT(*) AS count").
		Where("store_id = ?", storeID).
		Group("status").
		Scan(&rows).Error
	if err != nil {
		return stats, err
	}

	for _, row := range rows {
		stats.ByStatus[row.Status] = row.Count
	}

	err = r.db.WithContext(ctx).Model(&models.Product{}).
		Where("store_id = ? AND status = ? AND track_inventory = ? AND stock <= low_stock_threshold AND stock > 0",
			storeID, constants.ProductStatusActive, true).
		Count(&stats.LowStock).Error
	if err != nil {
		return stats, err
	}

	return stats, nil
}

// ListLowStockByStore returns active tracked products at or below the low-stock threshold.
func (r *ProductRepository) ListLowStockByStore(ctx context.Context, storeID uint, limit int) ([]models.Product, error) {
	if limit <= 0 {
		limit = 5
	}
	var products []models.Product
	err := r.db.WithContext(ctx).
		Where(
			"store_id = ? AND status = ? AND track_inventory = ? AND stock <= low_stock_threshold AND stock > 0",
			storeID, constants.ProductStatusActive, true,
		).
		Order("stock ASC").
		Limit(limit).
		Find(&products).Error
	return products, err
}

type ProductDemandRow struct {
	ProductID         uint
	Name              string
	Price             float64
	CompareAtPrice    *float64
	Cost              *float64
	Stock             int
	LowStockThreshold int
	UnitsSold         int64
	Revenue           float64
}

// ListDemandSignalsByStore joins active products with period sales for vendor AI forecasting and pricing.
func (r *ProductRepository) ListDemandSignalsByStore(
	ctx context.Context,
	storeID uint,
	from, to time.Time,
	trackInventoryOnly bool,
	limit int,
) ([]ProductDemandRow, error) {
	if limit <= 0 {
		limit = 50
	}

	revenueStatuses := []string{
		constants.OrderStatusPaid,
		constants.OrderStatusShipped,
		constants.OrderStatusDelivered,
	}

	query := r.db.WithContext(ctx).Table("products").
		Select(`
			products.id AS product_id,
			products.name,
			products.price,
			products.compare_at_price,
			products.cost,
			products.stock,
			products.low_stock_threshold,
			COALESCE(SUM(order_items.quantity), 0) AS units_sold,
			COALESCE(SUM(order_items.quantity * order_items.price), 0) AS revenue`).
		Joins(`LEFT JOIN order_items ON order_items.product_id = products.id AND order_items.deleted_at IS NULL`).
		Joins(`LEFT JOIN orders ON orders.id = order_items.order_id AND orders.deleted_at IS NULL AND orders.status IN ? AND orders.created_at >= ? AND orders.created_at < ?`,
			revenueStatuses, from, to).
		Where("products.store_id = ? AND products.status = ? AND products.deleted_at IS NULL",
			storeID, constants.ProductStatusActive)

	if trackInventoryOnly {
		query = query.Where("products.track_inventory = ?", true)
	}

	var rows []ProductDemandRow
	err := query.
		Group("products.id").
		Order("units_sold DESC, products.stock ASC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

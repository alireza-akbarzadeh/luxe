package postgres

import (
	"context"
	"time"

	apporder "github.com/alireza-akbarzadeh/luxe/internal/application/order"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

var vendorRevenueStatuses = []string{
	constants.OrderStatusPaid,
	constants.OrderStatusShipped,
	constants.OrderStatusDelivered,
}

type storeStatusCountRow struct {
	Status string
	Count  int64
}

type vendorTopProductRow struct {
	ProductID uint
	Name      string
	Revenue   float64
	Units     int64
}

type vendorDailySalesRow struct {
	Day     time.Time
	Revenue float64
	Orders  int64
}

func (r *OrderRepository) vendorStoreItemsQuery(ctx context.Context, storeID uint) *gorm.DB {
	return r.db.WithContext(ctx).
		Table("order_items").
		Joins("JOIN orders ON orders.id = order_items.order_id AND orders.deleted_at IS NULL").
		Joins("JOIN products ON products.id = order_items.product_id AND products.deleted_at IS NULL").
		Where("products.store_id = ?", storeID).
		Where("order_items.deleted_at IS NULL")
}

// GetVendorStoreSalesSummary returns revenue, orders, and units for a store in a date range.
func (r *OrderRepository) GetVendorStoreSalesSummary(
	ctx context.Context,
	storeID uint,
	from, to time.Time,
) (apporder.VendorSalesSummary, error) {
	var summary apporder.VendorSalesSummary

	type row struct {
		Revenue    float64
		UnitsSold  int64
		OrderCount int64
	}
	var result row

	err := r.vendorStoreItemsQuery(ctx, storeID).
		Where("orders.created_at >= ? AND orders.created_at < ?", from, to).
		Where("orders.status IN ?", vendorRevenueStatuses).
		Select(`
			COALESCE(SUM(order_items.quantity * order_items.price), 0) AS revenue,
			COALESCE(SUM(order_items.quantity), 0) AS units_sold,
			COUNT(DISTINCT orders.id) AS order_count`).
		Scan(&result).Error
	if err != nil {
		return summary, err
	}

	summary.Revenue = result.Revenue
	summary.UnitsSold = result.UnitsSold
	summary.OrderCount = result.OrderCount
	if summary.OrderCount > 0 {
		summary.AvgOrderValue = summary.Revenue / float64(summary.OrderCount)
	}

	return summary, nil
}

// ListVendorTopProducts returns top-selling products for a store in a date range.
func (r *OrderRepository) ListVendorTopProducts(
	ctx context.Context,
	storeID uint,
	from, to time.Time,
	limit int,
) ([]apporder.VendorTopProduct, error) {
	if limit <= 0 {
		limit = 5
	}

	var rows []vendorTopProductRow
	err := r.vendorStoreItemsQuery(ctx, storeID).
		Where("orders.created_at >= ? AND orders.created_at < ?", from, to).
		Where("orders.status IN ?", vendorRevenueStatuses).
		Select(`
			products.id AS product_id,
			products.name AS name,
			COALESCE(SUM(order_items.quantity * order_items.price), 0) AS revenue,
			COALESCE(SUM(order_items.quantity), 0) AS units`).
		Group("products.id, products.name").
		Order("revenue DESC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	products := make([]apporder.VendorTopProduct, len(rows))
	for i, row := range rows {
		products[i] = apporder.VendorTopProduct{
			ProductID: row.ProductID,
			Name:      row.Name,
			Revenue:   row.Revenue,
			Units:     row.Units,
		}
	}
	return products, nil
}

// ListVendorDailySales returns per-day revenue and order counts for charting.
func (r *OrderRepository) ListVendorDailySales(
	ctx context.Context,
	storeID uint,
	from, to time.Time,
) ([]apporder.VendorDailySales, error) {
	var rows []vendorDailySalesRow
	err := r.vendorStoreItemsQuery(ctx, storeID).
		Where("orders.created_at >= ? AND orders.created_at < ?", from, to).
		Where("orders.status IN ?", vendorRevenueStatuses).
		Select(`
			DATE(orders.created_at) AS day,
			COALESCE(SUM(order_items.quantity * order_items.price), 0) AS revenue,
			COUNT(DISTINCT orders.id) AS orders`).
		Group("DATE(orders.created_at)").
		Order("day ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	series := make([]apporder.VendorDailySales, len(rows))
	for i, row := range rows {
		series[i] = apporder.VendorDailySales{
			Date:    row.Day.Format("2006-01-02"),
			Revenue: row.Revenue,
			Orders:  row.Orders,
		}
	}
	return series, nil
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

type vendorStoreCustomerRow struct {
	UserID      uint
	FirstName   string
	LastName    string
	Email       string
	OrderCount  int64
	TotalSpend  float64
	LastOrderAt time.Time
}

// ListVendorStoreCustomers returns buyer aggregates for a store in a date range.
func (r *OrderRepository) ListVendorStoreCustomers(
	ctx context.Context,
	storeID uint,
	from, to time.Time,
	limit int,
) ([]apporder.VendorStoreCustomer, error) {
	if limit <= 0 {
		limit = 100
	}

	var rows []vendorStoreCustomerRow
	query := r.applyStoreScope(r.db.WithContext(ctx).Table("orders"), storeID).
		Joins("JOIN users ON users.id = orders.user_id AND users.deleted_at IS NULL").
		Where("orders.status IN ?", vendorRevenueStatuses)

	if !from.IsZero() {
		query = query.Where("orders.created_at >= ?", from)
	}
	if !to.IsZero() {
		query = query.Where("orders.created_at < ?", to)
	}

	err := query.
		Select(`
			users.id AS user_id,
			users.first_name,
			users.last_name,
			users.email,
			COUNT(DISTINCT orders.id) AS order_count,
			COALESCE(SUM(order_items.quantity * order_items.price), 0) AS total_spend,
			MAX(orders.created_at) AS last_order_at`).
		Group("users.id, users.first_name, users.last_name, users.email").
		Order("total_spend DESC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	customers := make([]apporder.VendorStoreCustomer, len(rows))
	for i, row := range rows {
		customers[i] = apporder.VendorStoreCustomer{
			UserID:      row.UserID,
			FirstName:   row.FirstName,
			LastName:    row.LastName,
			Email:       row.Email,
			OrderCount:  row.OrderCount,
			TotalSpend:  row.TotalSpend,
			LastOrderAt: row.LastOrderAt,
		}
	}
	return customers, nil
}

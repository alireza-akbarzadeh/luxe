package postgres

import (
	"context"
	"strings"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// AdminRepository provides admin dashboard and user management queries.
type AdminRepository struct {
	db *gorm.DB
}

// NewAdminRepository creates a GORM-backed admin repository.
func NewAdminRepository(db *gorm.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

var revenueOrderStatuses = []string{
	constants.OrderStatusPaid,
	constants.OrderStatusShipped,
	constants.OrderStatusDelivered,
}

// CountUsers returns total user count.
func (r *AdminRepository) CountUsers(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).Count(&count).Error
	return count, err
}

// CountActiveUsers returns active user count.
func (r *AdminRepository) CountActiveUsers(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).Where("is_active = ?", true).Count(&count).Error
	return count, err
}

// CountAdminUsers returns admin user count.
func (r *AdminRepository) CountAdminUsers(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).Where("role = ?", constants.RoleAdmin).Count(&count).Error
	return count, err
}

// CountOrders returns total order count.
func (r *AdminRepository) CountOrders(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Order{}).Count(&count).Error
	return count, err
}

// CountActiveProducts returns active product count.
func (r *AdminRepository) CountActiveProducts(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Product{}).Where("status = ?", constants.ProductStatusActive).Count(&count).Error
	return count, err
}

// SumRevenue returns total revenue for paid/shipped/delivered orders.
func (r *AdminRepository) SumRevenue(ctx context.Context) (float64, error) {
	var total float64
	err := r.db.WithContext(ctx).Model(&models.Order{}).
		Where("status IN ?", revenueOrderStatuses).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&total).Error
	return total, err
}

// CountPendingOrders returns pending order count.
func (r *AdminRepository) CountPendingOrders(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Order{}).Where("status = ?", constants.OrderStatusPending).Count(&count).Error
	return count, err
}

// SumWalletBalance returns total wallet balance.
func (r *AdminRepository) SumWalletBalance(ctx context.Context) (float64, error) {
	var total float64
	err := r.db.WithContext(ctx).Model(&models.Wallet{}).
		Select("COALESCE(SUM(balance), 0)").
		Scan(&total).Error
	return total, err
}

// CountLowStockProducts returns low-stock active product count.
func (r *AdminRepository) CountLowStockProducts(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Product{}).
		Where("stock <= low_stock_threshold AND status = ?", constants.ProductStatusActive).
		Count(&count).Error
	return count, err
}

func (r *AdminRepository) applyUserFilters(db *gorm.DB, filters dto.AdminUserFilters) *gorm.DB {
	if filters.Search != "" {
		term := "%" + strings.ToLower(filters.Search) + "%"
		db = db.Where(
			"LOWER(email) LIKE ? OR LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ? OR LOWER(CONCAT(first_name, ' ', last_name)) LIKE ?",
			term, term, term, term,
		)
	} else if filters.Email != "" {
		db = db.Where("email ILIKE ?", "%"+filters.Email+"%")
	}
	if filters.Role != "" {
		db = db.Where("role = ?", filters.Role)
	}
	if filters.IsActive != nil {
		db = db.Where("is_active = ?", *filters.IsActive)
	}
	return db
}

// CountUsersFiltered counts users matching filters.
func (r *AdminRepository) CountUsersFiltered(ctx context.Context, filters dto.AdminUserFilters) (int64, error) {
	db := r.applyUserFilters(r.db.WithContext(ctx).Model(&models.User{}), filters)
	var total int64
	err := db.Count(&total).Error
	return total, err
}

// ListUsersFiltered returns paginated users matching filters.
func (r *AdminRepository) ListUsersFiltered(ctx context.Context, filters dto.AdminUserFilters) ([]models.User, error) {
	db := r.applyUserFilters(r.db.WithContext(ctx).Model(&models.User{}), filters)
	var users []models.User
	err := db.Limit(filters.Limit).Offset(filters.Offset).
		Order("created_at DESC").Find(&users).Error
	return users, err
}

// UpdateUserRole sets a user's role.
func (r *AdminRepository) UpdateUserRole(ctx context.Context, userID uint, role string) (int64, error) {
	result := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("role", role)
	return result.RowsAffected, result.Error
}

// UpdateUserActive sets a user's active flag.
func (r *AdminRepository) UpdateUserActive(ctx context.Context, userID uint, active bool) (int64, error) {
	result := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("is_active", active)
	return result.RowsAffected, result.Error
}

// ListOrdersForExport returns orders for CSV export.
func (r *AdminRepository) ListOrdersForExport(ctx context.Context, filters dto.AdminOrderExportFilters) ([]models.Order, error) {
	db := r.db.WithContext(ctx).Model(&models.Order{}).
		Preload("User").
		Order("created_at DESC")

	if filters.Status != "" {
		db = db.Where("status = ?", filters.Status)
	}
	if filters.FromDate != "" {
		if t, err := time.Parse(time.DateOnly, filters.FromDate); err == nil {
			db = db.Where("created_at >= ?", t)
		}
	}
	if filters.ToDate != "" {
		if t, err := time.Parse(time.DateOnly, filters.ToDate); err == nil {
			db = db.Where("created_at <= ?", t.Add(24*time.Hour))
		}
	}

	var orders []models.Order
	err := db.Limit(10000).Find(&orders).Error
	return orders, err
}

// SumRevenueSince sums revenue for orders created since the given time.
func (r *AdminRepository) SumRevenueSince(ctx context.Context, since time.Time) (float64, error) {
	var total float64
	err := r.db.WithContext(ctx).Model(&models.Order{}).
		Where("created_at >= ? AND status IN ?", since, revenueOrderStatuses).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&total).Error
	return total, err
}

// SumRevenueBetween sums revenue for orders in a date range.
func (r *AdminRepository) SumRevenueBetween(ctx context.Context, start, end time.Time) (float64, error) {
	var total float64
	err := r.db.WithContext(ctx).Model(&models.Order{}).
		Where("created_at >= ? AND created_at < ? AND status IN ?", start, end, revenueOrderStatuses).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&total).Error
	return total, err
}

// CountOrdersSince counts orders created since the given time.
func (r *AdminRepository) CountOrdersSince(ctx context.Context, since time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Order{}).Where("created_at >= ?", since).Count(&count).Error
	return count, err
}

// CountOrdersBetween counts orders in a date range.
func (r *AdminRepository) CountOrdersBetween(ctx context.Context, start, end time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Order{}).
		Where("created_at >= ? AND created_at < ?", start, end).
		Count(&count).Error
	return count, err
}

// CountPaidOrdersSince counts paid-status orders since the given time.
func (r *AdminRepository) CountPaidOrdersSince(ctx context.Context, since time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Order{}).
		Where("created_at >= ? AND status IN ?", since, revenueOrderStatuses).
		Count(&count).Error
	return count, err
}

// CountPaidOrdersBetween counts paid-status orders in a date range.
func (r *AdminRepository) CountPaidOrdersBetween(ctx context.Context, start, end time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Order{}).
		Where("created_at >= ? AND created_at < ? AND status IN ?", start, end, revenueOrderStatuses).
		Count(&count).Error
	return count, err
}

// CountUsersSince counts users created since the given time.
func (r *AdminRepository) CountUsersSince(ctx context.Context, since time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).Where("created_at >= ?", since).Count(&count).Error
	return count, err
}

// CountUsersBetween counts users created in a date range.
func (r *AdminRepository) CountUsersBetween(ctx context.Context, start, end time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).
		Where("created_at >= ? AND created_at < ?", start, end).
		Count(&count).Error
	return count, err
}

// DashboardSeriesRow holds daily order/revenue aggregates.
type DashboardSeriesRow struct {
	Date    time.Time
	Revenue float64
	Orders  int64
}

// DailyOrderSeries returns daily revenue/order counts since start.
func (r *AdminRepository) DailyOrderSeries(ctx context.Context, start time.Time) ([]DashboardSeriesRow, error) {
	var rows []DashboardSeriesRow
	err := r.db.WithContext(ctx).Model(&models.Order{}).
		Select(`created_at::date AS date,
			COALESCE(SUM(CASE WHEN status IN ? THEN total_amount ELSE 0 END), 0) AS revenue,
			COUNT(*) AS orders`, revenueOrderStatuses).
		Where("created_at >= ?", start).
		Group("created_at::date").
		Order("date ASC").
		Scan(&rows).Error
	return rows, err
}

// RevenueDailyRow holds daily revenue report aggregates.
type RevenueDailyRow struct {
	Date       time.Time
	Revenue    float64
	Orders     int64
	PaidOrders int64
}

// DailyRevenueSeries returns daily revenue report rows since start.
func (r *AdminRepository) DailyRevenueSeries(ctx context.Context, start time.Time) ([]RevenueDailyRow, error) {
	var rows []RevenueDailyRow
	err := r.db.WithContext(ctx).Model(&models.Order{}).
		Select(`created_at::date AS date,
			COALESCE(SUM(CASE WHEN status IN ? THEN total_amount ELSE 0 END), 0) AS revenue,
			COUNT(*) AS orders,
			COUNT(CASE WHEN status IN ? THEN 1 END) AS paid_orders`, revenueOrderStatuses, revenueOrderStatuses).
		Where("created_at >= ?", start).
		Group("created_at::date").
		Order("date ASC").
		Scan(&rows).Error
	return rows, err
}

// OrderStatusCounts returns order counts grouped by status since start.
func (r *AdminRepository) OrderStatusCounts(ctx context.Context, since time.Time) ([]dto.AdminDashboardStatusCount, error) {
	var rows []dto.AdminDashboardStatusCount
	err := r.db.WithContext(ctx).Model(&models.Order{}).
		Select("status, COUNT(*) as count").
		Where("created_at >= ?", since).
		Group("status").
		Order("count DESC").
		Scan(&rows).Error
	return rows, err
}

// RecentOrders returns recent orders with user preload.
func (r *AdminRepository) RecentOrders(ctx context.Context, limit int) ([]models.Order, error) {
	var orders []models.Order
	err := r.db.WithContext(ctx).Preload("User").
		Order("created_at DESC").
		Limit(limit).
		Find(&orders).Error
	return orders, err
}

// TopProductRow holds top product aggregate data.
type TopProductRow struct {
	ID        uint
	Name      string
	SKU       string
	Stock     int
	UnitsSold int64
	Revenue   float64
}

// TopProducts returns top products by revenue since start.
func (r *AdminRepository) TopProducts(ctx context.Context, start time.Time, limit int) ([]TopProductRow, error) {
	var rows []TopProductRow
	err := r.db.WithContext(ctx).Model(&models.OrderItem{}).
		Select(`
			p.id,
			p.name,
			p.sku,
			p.stock,
			COALESCE(SUM(order_items.quantity), 0) AS units_sold,
			COALESCE(SUM(order_items.quantity * order_items.price), 0) AS revenue`).
		Joins("INNER JOIN orders o ON o.id = order_items.order_id AND o.deleted_at IS NULL").
		Joins("INNER JOIN products p ON p.id = order_items.product_id AND p.deleted_at IS NULL").
		Where("o.created_at >= ? AND o.status IN ?", start, revenueOrderStatuses).
		Group("p.id, p.name, p.sku, p.stock").
		Order("revenue DESC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

// LowStockProducts returns low-stock active products.
func (r *AdminRepository) LowStockProducts(ctx context.Context, limit int) ([]models.Product, error) {
	var products []models.Product
	err := r.db.WithContext(ctx).
		Where("stock <= low_stock_threshold AND status = ?", constants.ProductStatusActive).
		Order("stock ASC").
		Limit(limit).
		Find(&products).Error
	return products, err
}

// HourlyOrderRow holds hourly order aggregates.
type HourlyOrderRow struct {
	Hour    time.Time
	Revenue float64
	Orders  int64
}

// HourlyOrderSeries returns hourly revenue/order counts since start.
func (r *AdminRepository) HourlyOrderSeries(ctx context.Context, start time.Time) ([]HourlyOrderRow, error) {
	var rows []HourlyOrderRow
	err := r.db.WithContext(ctx).Model(&models.Order{}).
		Select(`date_trunc('hour', created_at) AS hour,
			COALESCE(SUM(CASE WHEN status IN ? THEN total_amount ELSE 0 END), 0) AS revenue,
			COUNT(*) AS orders`, revenueOrderStatuses).
		Where("created_at >= ?", start).
		Group("date_trunc('hour', created_at)").
		Order("hour ASC").
		Scan(&rows).Error
	return rows, err
}

// OrdersSince returns orders created since start with user preload.
func (r *AdminRepository) OrdersSince(ctx context.Context, start time.Time, limit int) ([]models.Order, error) {
	var orders []models.Order
	err := r.db.WithContext(ctx).Preload("User").
		Where("created_at >= ?", start).
		Order("created_at DESC").
		Limit(limit).
		Find(&orders).Error
	return orders, err
}

// RevenueOrderStatuses returns statuses counted as revenue (shared helper).
func RevenueOrderStatuses() []string {
	return revenueOrderStatuses
}

// RecentAuditLogs returns recent audit log entries with user preload.
func (r *AdminRepository) RecentAuditLogs(ctx context.Context, limit int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	err := r.db.WithContext(ctx).Preload("User").
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// PingDB measures database round-trip latency in milliseconds.
func (r *AdminRepository) PingDB(ctx context.Context) (int64, error) {
	start := time.Now()
	var one int
	err := r.db.WithContext(ctx).Raw("SELECT 1").Scan(&one).Error
	if err != nil {
		return 0, err
	}
	return time.Since(start).Milliseconds(), nil
}

// CountFailedWebhooksSince counts failed webhook events since start.
func (r *AdminRepository) CountFailedWebhooksSince(ctx context.Context, since time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.WebhookEvent{}).
		Where("created_at >= ? AND status = ?", since, "failed").
		Count(&count).Error
	return count, err
}

// CountWebhooksSince counts webhook events since start.
func (r *AdminRepository) CountWebhooksSince(ctx context.Context, since time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.WebhookEvent{}).
		Where("created_at >= ?", since).
		Count(&count).Error
	return count, err
}

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
	if filters.MembershipTier != "" {
		db = db.Where("membership_tier = ?", filters.MembershipTier)
	}
	if filters.CustomerSegment != "" {
		db = db.Where("customer_segment = ?", filters.CustomerSegment)
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

func (r *AdminRepository) applyProductExportFilters(db *gorm.DB, filters dto.AdminProductExportFilters) *gorm.DB {
	if filters.Status != "" {
		db = db.Where("status = ?", filters.Status)
	}
	if filters.Name != "" {
		db = db.Where("LOWER(name) LIKE LOWER(?)", "%"+filters.Name+"%")
	}
	if filters.SKU != "" {
		db = db.Where("sku LIKE ?", "%"+filters.SKU+"%")
	}
	if filters.CategoryID != 0 {
		db = db.Where("category_id = ?", filters.CategoryID)
	}
	if filters.BrandID != 0 {
		db = db.Where("brand_id = ?", filters.BrandID)
	}
	if filters.MinPrice != 0 {
		db = db.Where("price >= ?", filters.MinPrice)
	}
	if filters.MaxPrice != 0 {
		db = db.Where("price <= ?", filters.MaxPrice)
	}
	if filters.IsDigital != nil {
		db = db.Where("is_digital = ?", *filters.IsDigital)
	}
	return db
}

// ListProductsForExport returns products for CSV export (max 10 000 rows).
func (r *AdminRepository) ListProductsForExport(ctx context.Context, filters dto.AdminProductExportFilters) ([]models.Product, error) {
	db := r.applyProductExportFilters(
		r.db.WithContext(ctx).Model(&models.Product{}).
			Preload("Category").
			Preload("Brand").
			Order("id DESC"),
		filters,
	)

	var products []models.Product
	err := db.Limit(10000).Find(&products).Error
	return products, err
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

// UserOrderStats holds purchase metrics for a customer.
type UserOrderStats struct {
	OrderCount int64
	TotalSpent float64
}

// FindUserByID loads a user by primary key.
func (r *AdminRepository) FindUserByID(ctx context.Context, userID uint) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, userID).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserOrderStats returns order count and revenue for a user.
func (r *AdminRepository) GetUserOrderStats(ctx context.Context, userID uint) (UserOrderStats, error) {
	var stats UserOrderStats
	err := r.db.WithContext(ctx).Model(&models.Order{}).
		Where("user_id = ?", userID).
		Count(&stats.OrderCount).Error
	if err != nil {
		return stats, err
	}
	err = r.db.WithContext(ctx).Model(&models.Order{}).
		Where("user_id = ? AND status IN ?", userID, revenueOrderStatuses).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&stats.TotalSpent).Error
	return stats, err
}

// GetUserOrderStatsBatch returns purchase metrics keyed by user ID.
func (r *AdminRepository) GetUserOrderStatsBatch(ctx context.Context, userIDs []uint) (map[uint]UserOrderStats, error) {
	result := make(map[uint]UserOrderStats, len(userIDs))
	if len(userIDs) == 0 {
		return result, nil
	}

	type countRow struct {
		UserID uint
		Count  int64
	}
	var counts []countRow
	if err := r.db.WithContext(ctx).Model(&models.Order{}).
		Select("user_id, COUNT(*) AS count").
		Where("user_id IN ?", userIDs).
		Group("user_id").
		Scan(&counts).Error; err != nil {
		return nil, err
	}
	for _, row := range counts {
		entry := result[row.UserID]
		entry.OrderCount = row.Count
		result[row.UserID] = entry
	}

	type spendRow struct {
		UserID uint
		Total  float64
	}
	var spends []spendRow
	if err := r.db.WithContext(ctx).Model(&models.Order{}).
		Select("user_id, COALESCE(SUM(total_amount), 0) AS total").
		Where("user_id IN ? AND status IN ?", userIDs, revenueOrderStatuses).
		Group("user_id").
		Scan(&spends).Error; err != nil {
		return nil, err
	}
	for _, row := range spends {
		entry := result[row.UserID]
		entry.TotalSpent = row.Total
		result[row.UserID] = entry
	}

	return result, nil
}

// CountUserAddresses returns saved address count for a user.
func (r *AdminRepository) CountUserAddresses(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Address{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}

// ListUserAddresses returns all addresses for a user.
func (r *AdminRepository) ListUserAddresses(ctx context.Context, userID uint) ([]models.Address, error) {
	var addresses []models.Address
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("is_default DESC, created_at DESC").
		Find(&addresses).Error
	return addresses, err
}

// UpdateUserAdminNotes sets admin CRM notes on a user.
func (r *AdminRepository) UpdateUserAdminNotes(ctx context.Context, userID uint, notes string) (int64, error) {
	result := r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", userID).
		Update("admin_notes", notes)
	return result.RowsAffected, result.Error
}

// UpdateUserCustomerSegment assigns a CRM segment to a user.
func (r *AdminRepository) UpdateUserCustomerSegment(ctx context.Context, userID uint, segment string) (int64, error) {
	result := r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", userID).
		Update("customer_segment", segment)
	return result.RowsAffected, result.Error
}

// CountCustomers returns users with the default customer role.
func (r *AdminRepository) CountCustomers(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).
		Where("role = ?", constants.RoleUser).
		Count(&count).Error
	return count, err
}

// CountPlusMembers returns active Luxe Plus members.
func (r *AdminRepository) CountPlusMembers(ctx context.Context) (int64, error) {
	var count int64
	now := time.Now()
	err := r.db.WithContext(ctx).Model(&models.User{}).
		Where("membership_tier = ? AND (plus_expires_at IS NULL OR plus_expires_at > ?)", constants.MembershipTierPlus, now).
		Count(&count).Error
	return count, err
}

// CountNewCustomersSince returns customers created on or after the given time.
func (r *AdminRepository) CountNewCustomersSince(ctx context.Context, since time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).
		Where("role = ? AND created_at >= ?", constants.RoleUser, since).
		Count(&count).Error
	return count, err
}

// CountCustomersBySegment returns customers with the given CRM segment.
func (r *AdminRepository) CountCustomersBySegment(ctx context.Context, segment string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).
		Where("role = ? AND customer_segment = ?", constants.RoleUser, segment).
		Count(&count).Error
	return count, err
}

// ProfitTotals holds revenue, cost, and profit aggregates.
type ProfitTotals struct {
	Revenue float64
	Cost    float64
	Profit  float64
}

func (r *AdminRepository) profitTotalsBetween(ctx context.Context, start, end *time.Time) (ProfitTotals, error) {
	q := r.db.WithContext(ctx).Model(&models.OrderItem{}).
		Select(`
			COALESCE(SUM(order_items.quantity * order_items.price), 0) AS revenue,
			COALESCE(SUM(order_items.quantity * COALESCE(products.cost, 0)), 0) AS cost`).
		Joins("INNER JOIN orders o ON o.id = order_items.order_id AND o.deleted_at IS NULL").
		Joins("INNER JOIN products ON products.id = order_items.product_id AND products.deleted_at IS NULL").
		Where("o.status IN ?", revenueOrderStatuses)
	if start != nil {
		q = q.Where("o.created_at >= ?", *start)
	}
	if end != nil {
		q = q.Where("o.created_at < ?", *end)
	}
	var row struct {
		Revenue float64
		Cost    float64
	}
	if err := q.Scan(&row).Error; err != nil {
		return ProfitTotals{}, err
	}
	return ProfitTotals{
		Revenue: row.Revenue,
		Cost:    row.Cost,
		Profit:  row.Revenue - row.Cost,
	}, nil
}

// SumProfitSince returns profit (revenue minus product cost) since start.
func (r *AdminRepository) SumProfitSince(ctx context.Context, since time.Time) (ProfitTotals, error) {
	return r.profitTotalsBetween(ctx, &since, nil)
}

// SumProfitBetween returns profit between start (inclusive) and end (exclusive).
func (r *AdminRepository) SumProfitBetween(ctx context.Context, start, end time.Time) (ProfitTotals, error) {
	return r.profitTotalsBetween(ctx, &start, &end)
}

// ProfitDailyRow holds per-day profit aggregates.
type ProfitDailyRow struct {
	Date    time.Time
	Revenue float64
	Cost    float64
	Profit  float64
}

// DailyProfitSeries returns daily revenue, cost, and profit since start.
func (r *AdminRepository) DailyProfitSeries(ctx context.Context, start time.Time) ([]ProfitDailyRow, error) {
	var rows []struct {
		Date    time.Time
		Revenue float64
		Cost    float64
	}
	err := r.db.WithContext(ctx).Model(&models.OrderItem{}).
		Select(`
			date_trunc('day', o.created_at) AS date,
			COALESCE(SUM(order_items.quantity * order_items.price), 0) AS revenue,
			COALESCE(SUM(order_items.quantity * COALESCE(products.cost, 0)), 0) AS cost`).
		Joins("INNER JOIN orders o ON o.id = order_items.order_id AND o.deleted_at IS NULL").
		Joins("INNER JOIN products ON products.id = order_items.product_id AND products.deleted_at IS NULL").
		Where("o.created_at >= ? AND o.status IN ?", start, revenueOrderStatuses).
		Group("date_trunc('day', o.created_at)").
		Order("date ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]ProfitDailyRow, len(rows))
	for i, row := range rows {
		out[i] = ProfitDailyRow{
			Date:    row.Date,
			Revenue: row.Revenue,
			Cost:    row.Cost,
			Profit:  row.Revenue - row.Cost,
		}
	}
	return out, nil
}

// CustomerSegmentRow holds segment counts for analytics.
type CustomerSegmentRow struct {
	Segment string
	Count   int64
}

// CustomerSegmentCounts returns customer counts grouped by CRM segment.
func (r *AdminRepository) CustomerSegmentCounts(ctx context.Context) ([]CustomerSegmentRow, error) {
	var rows []CustomerSegmentRow
	err := r.db.WithContext(ctx).Model(&models.User{}).
		Select(`COALESCE(NULLIF(customer_segment, ''), 'unassigned') AS segment, COUNT(*) AS count`).
		Where("role = ?", constants.RoleUser).
		Group("COALESCE(NULLIF(customer_segment, ''), 'unassigned')").
		Order("count DESC").
		Scan(&rows).Error
	return rows, err
}

// OrderFunnelRow holds funnel stage counts.
type OrderFunnelRow struct {
	Status string
	Count  int64
}

// OrderFunnelCounts returns order counts by status since start.
func (r *AdminRepository) OrderFunnelCounts(ctx context.Context, since time.Time) ([]OrderFunnelRow, error) {
	var rows []OrderFunnelRow
	err := r.db.WithContext(ctx).Model(&models.Order{}).
		Select("status, COUNT(*) AS count").
		Where("created_at >= ?", since).
		Group("status").
		Scan(&rows).Error
	return rows, err
}

// CohortRow holds monthly cohort metrics.
type CohortRow struct {
	Cohort         string
	Customers      int64
	RepeatCustomers int64
	Revenue        float64
}

// MonthlyCohorts returns cohort metrics for customers whose first order was since start.
func (r *AdminRepository) MonthlyCohorts(ctx context.Context, since time.Time, limit int) ([]CohortRow, error) {
	var rows []CohortRow
	err := r.db.WithContext(ctx).Raw(`
		WITH first_orders AS (
			SELECT user_id, MIN(created_at) AS first_order_at
			FROM orders
			WHERE deleted_at IS NULL AND user_id IS NOT NULL
			GROUP BY user_id
		),
		cohort_base AS (
			SELECT
				fo.user_id,
				to_char(date_trunc('month', fo.first_order_at), 'YYYY-MM') AS cohort
			FROM first_orders fo
			WHERE fo.first_order_at >= ?
		),
		repeat_flags AS (
			SELECT
				cb.cohort,
				cb.user_id,
				EXISTS (
					SELECT 1 FROM orders o2
					WHERE o2.user_id = cb.user_id
						AND o2.deleted_at IS NULL
						AND o2.id != (
							SELECT o3.id FROM orders o3
							WHERE o3.user_id = cb.user_id AND o3.deleted_at IS NULL
							ORDER BY o3.created_at ASC LIMIT 1
						)
				) AS is_repeat
			FROM cohort_base cb
		),
		cohort_revenue AS (
			SELECT
				cb.cohort,
				COALESCE(SUM(o.total_amount), 0) AS revenue
			FROM cohort_base cb
			INNER JOIN orders o ON o.user_id = cb.user_id AND o.deleted_at IS NULL
			WHERE o.status IN ?
			GROUP BY cb.cohort
		)
		SELECT
			rf.cohort,
			COUNT(DISTINCT rf.user_id) AS customers,
			COUNT(DISTINCT CASE WHEN rf.is_repeat THEN rf.user_id END) AS repeat_customers,
			COALESCE(cr.revenue, 0) AS revenue
		FROM repeat_flags rf
		LEFT JOIN cohort_revenue cr ON cr.cohort = rf.cohort
		GROUP BY rf.cohort, cr.revenue
		ORDER BY rf.cohort DESC
		LIMIT ?
	`, since, revenueOrderStatuses, limit).Scan(&rows).Error
	return rows, err
}


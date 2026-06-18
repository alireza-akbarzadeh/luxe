package dto

import "time"

// AdminStatsResponse holds platform-wide aggregated metrics for admin dashboards.
type AdminStatsResponse struct {
	TotalUsers          int64   `json:"total_users"`
	ActiveUsers         int64   `json:"active_users"`
	AdminUsers          int64   `json:"admin_users"`
	TotalOrders         int64   `json:"total_orders"`
	TotalActiveProducts int64   `json:"total_active_products"`
	TotalRevenue        float64 `json:"total_revenue"`
	PendingOrders       int64   `json:"pending_orders"`
	TotalWalletBalance  float64 `json:"total_wallet_balance"`
	LowStockProducts    int64   `json:"low_stock_products"`
}

// AdminUserFilters are query params for the admin user listing endpoint.
type AdminUserFilters struct {
	Search   string `form:"search"`
	Email    string `form:"email"`
	Role     string `form:"role"`
	IsActive *bool  `form:"is_active"`
	Limit    int    `form:"limit"`
	Offset   int    `form:"offset"`
}

// AdminUserResponse is the safe public projection of a User for admin views.
type AdminUserResponse struct {
	ID              uint       `json:"id"`
	Email           string     `json:"email"`
	FirstName       string     `json:"first_name"`
	LastName        string     `json:"last_name"`
	Role            string     `json:"role"`
	IsActive        bool       `json:"is_active"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	LastLoginAt     *time.Time `json:"last_login_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// UpdateUserRoleRequest is the body for PATCH /admin/users/:id/role.
type UpdateUserRoleRequest struct {
	Role string `json:"role" validate:"required,min=1,max=80"`
}

// ToggleUserActiveRequest is the body for PATCH /admin/users/:id/active.
type ToggleUserActiveRequest struct {
	IsActive bool `json:"is_active"`
}

// BulkUpdateOrderStatusRequest is the body for POST /admin/orders/bulk-status.
type BulkUpdateOrderStatusRequest struct {
	OrderIDs []uint `json:"order_ids" validate:"required,min=1,max=500,dive,gt=0"`
	Status   string `json:"status"    validate:"required,oneof=paid shipped delivered cancelled"`
}

// AdminOrderExportFilters are query params for the CSV export endpoint.
type AdminOrderExportFilters struct {
	Status   string `form:"status"`
	FromDate string `form:"from_date"`
	ToDate   string `form:"to_date"`
}

// AdminDashboardFilters are query params for GET /admin/dashboard/overview.
type AdminDashboardFilters struct {
	Period string `form:"period" validate:"omitempty,oneof=7d 30d 90d"`
}

// AdminDashboardKPI is a metric with period-over-period comparison.
type AdminDashboardKPI struct {
	Value         float64 `json:"value"`
	PreviousValue float64 `json:"previous_value"`
	ChangePercent float64 `json:"change_percent"`
}

// AdminDashboardKPIs groups headline KPI cards for the admin home dashboard.
type AdminDashboardKPIs struct {
	Revenue       AdminDashboardKPI `json:"revenue"`
	Orders        AdminDashboardKPI `json:"orders"`
	AvgOrderValue AdminDashboardKPI `json:"avg_order_value"`
	NewCustomers  AdminDashboardKPI `json:"new_customers"`
}

// AdminDashboardSeriesPoint is one day in the revenue/orders time series.
type AdminDashboardSeriesPoint struct {
	Date    string  `json:"date"`
	Revenue float64 `json:"revenue"`
	Orders  int64   `json:"orders"`
}

// AdminDashboardStatusCount is orders grouped by status for the selected period.
type AdminDashboardStatusCount struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

// AdminDashboardRecentOrder is a lightweight row for the recent orders table.
type AdminDashboardRecentOrder struct {
	ID           uint      `json:"id"`
	OrderNumber  string    `json:"order_number"`
	Status       string    `json:"status"`
	TotalAmount  float64   `json:"total_amount"`
	Currency     string    `json:"currency"`
	CustomerName string    `json:"customer_name"`
	CustomerEmail string   `json:"customer_email"`
	CreatedAt    time.Time `json:"created_at"`
}

// AdminDashboardTopProduct is a best-seller row for the dashboard table.
type AdminDashboardTopProduct struct {
	ID        uint    `json:"id"`
	Name      string  `json:"name"`
	SKU       string  `json:"sku"`
	UnitsSold int64   `json:"units_sold"`
	Revenue   float64 `json:"revenue"`
	Stock     int     `json:"stock"`
}

// AdminDashboardLowStockProduct is an inventory alert row.
type AdminDashboardLowStockProduct struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	SKU       string `json:"sku"`
	Stock     int    `json:"stock"`
	Threshold int    `json:"threshold"`
}

// AdminSalesFeedEvent is one row in the live sales activity feed.
type AdminSalesFeedEvent struct {
	ID        string  `json:"id"`
	Type      string  `json:"type"`
	Title     string  `json:"title"`
	Subtitle  string  `json:"subtitle"`
	Amount    float64 `json:"amount,omitempty"`
	Timestamp int64   `json:"timestamp"`
}

// AdminSalesFeedSnapshotResponse seeds the live sales feed page before WebSocket updates.
type AdminSalesFeedSnapshotResponse struct {
	TotalOrdersToday  int64                       `json:"total_orders_today"`
	TotalRevenueToday float64                     `json:"total_revenue_today"`
	StatusCounts      []AdminDashboardStatusCount `json:"status_counts"`
	RecentEvents      []AdminSalesFeedEvent       `json:"recent_events"`
	RevenueSeries     []AdminDashboardSeriesPoint `json:"revenue_series"`
	GeneratedAt       time.Time                   `json:"generated_at"`
}

// AdminDashboardOverviewResponse powers the admin commerce dashboard.
type AdminDashboardOverviewResponse struct {
	Period           string                          `json:"period"`
	GeneratedAt      time.Time                       `json:"generated_at"`
	KPIs             AdminDashboardKPIs              `json:"kpis"`
	RevenueSeries    []AdminDashboardSeriesPoint     `json:"revenue_series"`
	OrdersByStatus   []AdminDashboardStatusCount     `json:"orders_by_status"`
	RecentOrders     []AdminDashboardRecentOrder     `json:"recent_orders"`
	TopProducts      []AdminDashboardTopProduct      `json:"top_products"`
	LowStockProducts []AdminDashboardLowStockProduct `json:"low_stock_products"`
	Platform         AdminStatsResponse              `json:"platform"`
}

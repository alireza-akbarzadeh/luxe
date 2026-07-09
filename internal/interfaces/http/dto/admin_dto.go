package dto

import (
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

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
	Search          string `form:"search"`
	Email           string `form:"email"`
	Role            string `form:"role"`
	IsActive        *bool  `form:"is_active"`
	MembershipTier  string `form:"membership_tier"`
	CustomerSegment string `form:"customer_segment"`
	Limit           int    `form:"limit"`
	Offset          int    `form:"offset"`
}

// AdminUserResponse is the safe public projection of a User for admin views.
type AdminUserResponse struct {
	ID              uint       `json:"id"`
	Email           string     `json:"email"`
	FirstName       string     `json:"first_name"`
	LastName        string     `json:"last_name"`
	Role            string     `json:"role"`
	IsActive        bool       `json:"is_active"`
	Phone           string     `json:"phone,omitempty"`
	AvatarURL       string     `json:"avatar_url,omitempty"`
	MembershipTier  string     `json:"membership_tier,omitempty"`
	IsPlusActive    bool       `json:"is_plus_active"`
	CustomerSegment string     `json:"customer_segment,omitempty"`
	OrderCount      int64      `json:"order_count"`
	TotalSpent      float64    `json:"total_spent"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	LastLoginAt     *time.Time `json:"last_login_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// AdminCustomerDetailResponse is the full customer profile for admin detail views.
type AdminCustomerDetailResponse struct {
	AdminUserResponse
	PlusSubscribedAt *time.Time `json:"plus_subscribed_at,omitempty"`
	PlusExpiresAt    *time.Time `json:"plus_expires_at,omitempty"`
	AdminNotes       string     `json:"admin_notes,omitempty"`
	AddressCount     int64      `json:"address_count"`
}

// AdminCustomerStats holds aggregate CRM metrics for the customers dashboard.
type AdminCustomerStats struct {
	TotalCustomers int64 `json:"total_customers"`
	PlusMembers    int64 `json:"plus_members"`
	NewThisMonth   int64 `json:"new_this_month"`
	VipCustomers   int64 `json:"vip_customers"`
}

// UpdateCustomerNotesRequest updates admin-only notes on a customer profile.
type UpdateCustomerNotesRequest struct {
	AdminNotes string `json:"admin_notes" validate:"max=2048"`
}

// UpdateCustomerSegmentRequest assigns a CRM segment to a customer.
type UpdateCustomerSegmentRequest struct {
	CustomerSegment string `json:"customer_segment" validate:"omitempty,oneof=vip loyal new at_risk"`
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

// AdminProductExportFilters are query params for GET /admin/products/export.
type AdminProductExportFilters struct {
	Status     string  `form:"status"`
	Name       string  `form:"name"`
	SKU        string  `form:"sku"`
	CategoryID uint    `form:"category_id"`
	BrandID    uint    `form:"brand_id"`
	MinPrice   float64 `form:"min_price"`
	MaxPrice   float64 `form:"max_price"`
	IsDigital  *bool   `form:"is_digital"`
	Format     string  `form:"format" validate:"omitempty,oneof=csv"`
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

// AdminDashboardActivity is one row in the dashboard recent-activity feed.
type AdminDashboardActivity struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Actor       string    `json:"actor,omitempty"`
	Href        string    `json:"href,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// AdminDashboardInsight is a rule-based business insight for the admin home dashboard.
type AdminDashboardInsight struct {
	ID          string `json:"id"`
	Severity    string `json:"severity"`
	Title       string `json:"title"`
	Body        string `json:"body"`
	ActionLabel string `json:"action_label,omitempty"`
	ActionHref  string `json:"action_href,omitempty"`
}

// AdminDashboardHealth summarizes platform operational health.
type AdminDashboardHealth struct {
	Status       string  `json:"status"`
	APILatencyMs int64   `json:"api_latency_ms"`
	ErrorRate    float64 `json:"error_rate"`
	QueueDepth   int64   `json:"queue_depth"`
	Message      string  `json:"message,omitempty"`
}

// AdminDashboardOverviewResponse powers the admin commerce dashboard.
type AdminDashboardOverviewResponse struct {
	Period           string                          `json:"period"`
	GeneratedAt      time.Time                       `json:"generated_at"`
	KPIs             AdminDashboardKPIs              `json:"kpis"`
	Conversion       AdminDashboardKPI               `json:"conversion"`
	KpiSparklines    map[string][]float64            `json:"kpi_sparklines"`
	RevenueSeries    []AdminDashboardSeriesPoint     `json:"revenue_series"`
	OrdersByStatus   []AdminDashboardStatusCount     `json:"orders_by_status"`
	RecentActivity   []AdminDashboardActivity        `json:"recent_activity"`
	AiInsights       []AdminDashboardInsight         `json:"ai_insights"`
	PlatformHealth   AdminDashboardHealth            `json:"platform_health"`
	RecentOrders     []AdminDashboardRecentOrder     `json:"recent_orders"`
	TopProducts      []AdminDashboardTopProduct      `json:"top_products"`
	LowStockProducts []AdminDashboardLowStockProduct `json:"low_stock_products"`
	Platform         AdminStatsResponse              `json:"platform"`
}

// AdminDashboardExportFilters are query params for GET /admin/dashboard/export.
type AdminDashboardExportFilters struct {
	Period string `form:"period" validate:"omitempty,oneof=7d 30d 90d"`
	Format string `form:"format" validate:"omitempty,oneof=csv"`
}

// AdminRevenueReportFilters are query params for GET /admin/reports/revenue.
type AdminRevenueReportFilters struct {
	Period string `form:"period" validate:"omitempty,oneof=7d 30d 90d"`
}

// AdminRevenueDailyRow is one day in the revenue report breakdown.
type AdminRevenueDailyRow struct {
	Date          string  `json:"date"`
	Revenue       float64 `json:"revenue"`
	Orders        int64   `json:"orders"`
	PaidOrders    int64   `json:"paid_orders"`
	AvgOrderValue float64 `json:"avg_order_value"`
}

// AdminRevenueReportSummary groups headline metrics for the revenue report.
type AdminRevenueReportSummary struct {
	Revenue         AdminDashboardKPI `json:"revenue"`
	Orders          AdminDashboardKPI `json:"orders"`
	AvgOrderValue   AdminDashboardKPI `json:"avg_order_value"`
	AvgDailyRevenue AdminDashboardKPI `json:"avg_daily_revenue"`
}

// AdminRevenueReportResponse powers the admin daily revenue report page.
type AdminRevenueReportResponse struct {
	Period      string                    `json:"period"`
	GeneratedAt time.Time                 `json:"generated_at"`
	Summary     AdminRevenueReportSummary `json:"summary"`
	Daily       []AdminRevenueDailyRow    `json:"daily"`
}

// AdminSalesAnalyticsFilters are query params for GET /admin/analytics/sales.
type AdminSalesAnalyticsFilters struct {
	Period string `form:"period" validate:"omitempty,oneof=7d 30d 90d"`
}

// AdminSalesAnalyticsKPIs groups headline metrics for the sales analytics dashboard.
type AdminSalesAnalyticsKPIs struct {
	Revenue       AdminDashboardKPI `json:"revenue"`
	Profit        AdminDashboardKPI `json:"profit"`
	Orders        AdminDashboardKPI `json:"orders"`
	Customers     AdminDashboardKPI `json:"customers"`
	AvgOrderValue AdminDashboardKPI `json:"avg_order_value"`
}

// AdminSalesProfitPoint is one day in the profit time series.
type AdminSalesProfitPoint struct {
	Date    string  `json:"date"`
	Revenue float64 `json:"revenue"`
	Cost    float64 `json:"cost"`
	Profit  float64 `json:"profit"`
}

// AdminSalesSegmentCount is customers grouped by CRM segment.
type AdminSalesSegmentCount struct {
	Segment string `json:"segment"`
	Count   int64  `json:"count"`
}

// AdminSalesFunnelStep is one stage in the order lifecycle funnel.
type AdminSalesFunnelStep struct {
	Step  string  `json:"step"`
	Count int64   `json:"count"`
	Rate  float64 `json:"rate"`
}

// AdminSalesCohortRow is monthly cohort retention and revenue.
type AdminSalesCohortRow struct {
	Cohort      string  `json:"cohort"`
	Customers   int64   `json:"customers"`
	RepeatRate  float64 `json:"repeat_rate"`
	Revenue     float64 `json:"revenue"`
}

// AdminSalesAnalyticsResponse powers the interactive sales analytics dashboard.
type AdminSalesAnalyticsResponse struct {
	Period           string                      `json:"period"`
	GeneratedAt      time.Time                   `json:"generated_at"`
	KPIs             AdminSalesAnalyticsKPIs     `json:"kpis"`
	RevenueSeries    []AdminDashboardSeriesPoint `json:"revenue_series"`
	ProfitSeries     []AdminSalesProfitPoint     `json:"profit_series"`
	OrdersByStatus   []AdminDashboardStatusCount `json:"orders_by_status"`
	TopProducts      []AdminDashboardTopProduct  `json:"top_products"`
	CustomerSegments []AdminSalesSegmentCount    `json:"customer_segments"`
	OrderFunnel      []AdminSalesFunnelStep      `json:"order_funnel"`
	Cohorts          []AdminSalesCohortRow       `json:"cohorts"`
}

// ToAdminUserResponse maps a user model to the admin list/detail projection.
func ToAdminUserResponse(user *models.User, stats UserOrderStats) AdminUserResponse {
	if user == nil {
		return AdminUserResponse{}
	}
	public := ToUserResponse(user)
	return AdminUserResponse{
		ID:              user.ID,
		Email:           user.Email,
		FirstName:       user.FirstName,
		LastName:        user.LastName,
		Role:            user.Role,
		IsActive:        user.IsActive,
		Phone:           user.Phone,
		AvatarURL:       user.AvatarURL,
		MembershipTier:  public.MembershipTier,
		IsPlusActive:    public.IsPlusActive,
		CustomerSegment: user.CustomerSegment,
		OrderCount:      stats.OrderCount,
		TotalSpent:      stats.TotalSpent,
		EmailVerifiedAt: user.EmailVerifiedAt,
		LastLoginAt:     user.LastLoginAt,
		CreatedAt:       user.CreatedAt,
	}
}

// UserOrderStats holds purchase metrics attached to admin user responses.
type UserOrderStats struct {
	OrderCount int64
	TotalSpent float64
}


package dto

import "time"

// AdminStatsResponse holds platform-wide aggregated metrics for admin dashboards.
type AdminStatsResponse struct {
	TotalUsers          int64   `json:"total_users"`
	TotalOrders         int64   `json:"total_orders"`
	TotalActiveProducts int64   `json:"total_active_products"`
	TotalRevenue        float64 `json:"total_revenue"`
	PendingOrders       int64   `json:"pending_orders"`
	TotalWalletBalance  float64 `json:"total_wallet_balance"`
	LowStockProducts    int64   `json:"low_stock_products"`
}

// AdminUserFilters are query params for the admin user listing endpoint.
type AdminUserFilters struct {
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
	Role string `json:"role" validate:"required,oneof=admin user"`
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

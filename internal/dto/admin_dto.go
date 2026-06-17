package dto

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

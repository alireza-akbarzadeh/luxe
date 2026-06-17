package services

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
)

var revenueOrderStatuses = []string{
	constants.OrderStatusPaid,
	constants.OrderStatusShipped,
	constants.OrderStatusDelivered,
}

func dashboardPeriodDays(period string) int {
	switch period {
	case "7d":
		return 7
	case "90d":
		return 90
	default:
		return 30
	}
}

func dashboardKPI(current, previous float64) dto.AdminDashboardKPI {
	change := 0.0
	if previous > 0 {
		change = ((current - previous) / previous) * 100
	} else if current > 0 {
		change = 100
	}
	return dto.AdminDashboardKPI{
		Value:         current,
		PreviousValue: previous,
		ChangePercent: math.Round(change*10) / 10,
	}
}

func (s *adminService) GetDashboardOverview(ctx context.Context, filters dto.AdminDashboardFilters) (*dto.AdminDashboardOverviewResponse, error) {
	days := dashboardPeriodDays(filters.Period)
	periodLabel := fmt.Sprintf("%dd", days)
	if filters.Period != "" {
		periodLabel = filters.Period
	}

	now := time.Now().UTC()
	currentStart := now.AddDate(0, 0, -days)
	previousStart := currentStart.AddDate(0, 0, -days)

	db := s.db.WithContext(ctx)

	platform, err := s.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	var currentRevenue float64
	if err := db.Model(&models.Order{}).
		Where("created_at >= ? AND status IN ?", currentStart, revenueOrderStatuses).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&currentRevenue).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	var previousRevenue float64
	if err := db.Model(&models.Order{}).
		Where("created_at >= ? AND created_at < ? AND status IN ?", previousStart, currentStart, revenueOrderStatuses).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&previousRevenue).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	var currentOrders int64
	if err := db.Model(&models.Order{}).
		Where("created_at >= ?", currentStart).
		Count(&currentOrders).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	var previousOrders int64
	if err := db.Model(&models.Order{}).
		Where("created_at >= ? AND created_at < ?", previousStart, currentStart).
		Count(&previousOrders).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	currentAOV := 0.0
	if currentOrders > 0 {
		currentAOV = currentRevenue / float64(currentOrders)
	}
	previousAOV := 0.0
	if previousOrders > 0 {
		previousAOV = previousRevenue / float64(previousOrders)
	}

	var currentCustomers int64
	if err := db.Model(&models.User{}).
		Where("created_at >= ?", currentStart).
		Count(&currentCustomers).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	var previousCustomers int64
	if err := db.Model(&models.User{}).
		Where("created_at >= ? AND created_at < ?", previousStart, currentStart).
		Count(&previousCustomers).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	type seriesRow struct {
		Date    time.Time
		Revenue float64
		Orders  int64
	}
	var seriesRows []seriesRow
	if err := db.Raw(`
		SELECT created_at::date AS date,
		       COALESCE(SUM(CASE WHEN status IN ? THEN total_amount ELSE 0 END), 0) AS revenue,
		       COUNT(*) AS orders
		FROM orders
		WHERE created_at >= ? AND deleted_at IS NULL
		GROUP BY created_at::date
		ORDER BY date ASC`, revenueOrderStatuses, currentStart).Scan(&seriesRows).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	seriesMap := make(map[string]dto.AdminDashboardSeriesPoint, len(seriesRows))
	for _, row := range seriesRows {
		key := row.Date.Format(time.DateOnly)
		seriesMap[key] = dto.AdminDashboardSeriesPoint{
			Date:    key,
			Revenue: row.Revenue,
			Orders:  row.Orders,
		}
	}

	revenueSeries := make([]dto.AdminDashboardSeriesPoint, 0, days)
	for i := days - 1; i >= 0; i-- {
		day := now.AddDate(0, 0, -i).Format(time.DateOnly)
		if point, ok := seriesMap[day]; ok {
			revenueSeries = append(revenueSeries, point)
		} else {
			revenueSeries = append(revenueSeries, dto.AdminDashboardSeriesPoint{Date: day})
		}
	}

	var statusRows []dto.AdminDashboardStatusCount
	if err := db.Model(&models.Order{}).
		Select("status, COUNT(*) as count").
		Where("created_at >= ?", currentStart).
		Group("status").
		Order("count DESC").
		Scan(&statusRows).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	var recentOrders []models.Order
	if err := db.Preload("User").
		Order("created_at DESC").
		Limit(8).
		Find(&recentOrders).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	recent := make([]dto.AdminDashboardRecentOrder, len(recentOrders))
	for i, order := range recentOrders {
		name := fmt.Sprintf("%s %s", order.User.FirstName, order.User.LastName)
		if name == " " {
			name = order.User.Email
		}
		recent[i] = dto.AdminDashboardRecentOrder{
			ID:            order.ID,
			OrderNumber:   order.OrderNumber,
			Status:        order.Status,
			TotalAmount:   order.TotalAmount,
			Currency:      order.Currency,
			CustomerName:  name,
			CustomerEmail: order.User.Email,
			CreatedAt:     order.CreatedAt,
		}
	}

	type topProductRow struct {
		ID        uint
		Name      string
		SKU       string
		Stock     int
		UnitsSold int64
		Revenue   float64
	}
	var topRows []topProductRow
	if err := db.Table("order_items oi").
		Select(`
			p.id,
			p.name,
			p.sku,
			p.stock,
			COALESCE(SUM(oi.quantity), 0) AS units_sold,
			COALESCE(SUM(oi.quantity * oi.price), 0) AS revenue`).
		Joins("INNER JOIN orders o ON o.id = oi.order_id AND o.deleted_at IS NULL").
		Joins("INNER JOIN products p ON p.id = oi.product_id AND p.deleted_at IS NULL").
		Where("o.created_at >= ? AND o.status IN ?", currentStart, revenueOrderStatuses).
		Group("p.id, p.name, p.sku, p.stock").
		Order("revenue DESC").
		Limit(5).
		Scan(&topRows).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	topProducts := make([]dto.AdminDashboardTopProduct, len(topRows))
	for i, row := range topRows {
		topProducts[i] = dto.AdminDashboardTopProduct{
			ID:        row.ID,
			Name:      row.Name,
			SKU:       row.SKU,
			UnitsSold: row.UnitsSold,
			Revenue:   row.Revenue,
			Stock:     row.Stock,
		}
	}

	var lowStock []models.Product
	if err := db.Where("stock <= low_stock_threshold AND status = ?", constants.ProductStatusActive).
		Order("stock ASC").
		Limit(6).
		Find(&lowStock).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	lowStockProducts := make([]dto.AdminDashboardLowStockProduct, len(lowStock))
	for i, product := range lowStock {
		lowStockProducts[i] = dto.AdminDashboardLowStockProduct{
			ID:        product.ID,
			Name:      product.Name,
			SKU:       product.SKU,
			Stock:     product.Stock,
			Threshold: product.LowStockThreshold,
		}
	}

	return &dto.AdminDashboardOverviewResponse{
		Period:      periodLabel,
		GeneratedAt: now,
		KPIs: dto.AdminDashboardKPIs{
			Revenue:       dashboardKPI(currentRevenue, previousRevenue),
			Orders:        dashboardKPI(float64(currentOrders), float64(previousOrders)),
			AvgOrderValue: dashboardKPI(currentAOV, previousAOV),
			NewCustomers:  dashboardKPI(float64(currentCustomers), float64(previousCustomers)),
		},
		RevenueSeries:    revenueSeries,
		OrdersByStatus:   statusRows,
		RecentOrders:     recent,
		TopProducts:      topProducts,
		LowStockProducts: lowStockProducts,
		Platform:         *platform,
	}, nil
}

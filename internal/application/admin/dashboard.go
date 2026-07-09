package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

// GetDashboardOverview builds the admin dashboard overview response.
func (q *Queries) GetDashboardOverview(ctx context.Context, filters dto.AdminDashboardFilters) (*dto.AdminDashboardOverviewResponse, error) {
	days := dashboardPeriodDays(filters.Period)
	periodLabel := fmt.Sprintf("%dd", days)
	if filters.Period != "" {
		periodLabel = filters.Period
	}

	now := time.Now().UTC()
	currentStart := now.AddDate(0, 0, -days)
	previousStart := currentStart.AddDate(0, 0, -days)

	platform, err := q.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	currentRevenue, err := q.repo.SumRevenueSince(ctx, currentStart)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	previousRevenue, err := q.repo.SumRevenueBetween(ctx, previousStart, currentStart)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	currentOrders, err := q.repo.CountOrdersSince(ctx, currentStart)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	previousOrders, err := q.repo.CountOrdersBetween(ctx, previousStart, currentStart)
	if err != nil {
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

	currentCustomers, err := q.repo.CountUsersSince(ctx, currentStart)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	previousCustomers, err := q.repo.CountUsersBetween(ctx, previousStart, currentStart)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	seriesRows, err := q.repo.DailyOrderSeries(ctx, currentStart)
	if err != nil {
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

	statusRows, err := q.repo.OrderStatusCounts(ctx, currentStart)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	recentOrders, err := q.repo.RecentOrders(ctx, 8)
	if err != nil {
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

	topRows, err := q.repo.TopProducts(ctx, currentStart, 5)
	if err != nil {
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

	lowStock, err := q.repo.LowStockProducts(ctx, 6)
	if err != nil {
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

	currentPaid, err := q.repo.CountPaidOrdersSince(ctx, currentStart)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	previousPaid, err := q.repo.CountPaidOrdersBetween(ctx, previousStart, currentStart)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	conversionCurrent := conversionRate(currentPaid, currentOrders)
	conversionPrevious := conversionRate(previousPaid, previousOrders)

	kpis := dto.AdminDashboardKPIs{
		Revenue:       toDashboardKPI(currentRevenue, previousRevenue),
		Orders:        toDashboardKPI(float64(currentOrders), float64(previousOrders)),
		AvgOrderValue: toDashboardKPI(currentAOV, previousAOV),
		NewCustomers:  toDashboardKPI(float64(currentCustomers), float64(previousCustomers)),
	}

	auditLogs, err := q.repo.RecentAuditLogs(ctx, 8)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	activityOrders, err := q.repo.OrdersSince(ctx, currentStart, 8)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	platformHealth, err := buildPlatformHealth(ctx, q, *platform)
	if err != nil {
		return nil, err
	}

	return &dto.AdminDashboardOverviewResponse{
		Period:         periodLabel,
		GeneratedAt:    now,
		KPIs:           kpis,
		Conversion:     toDashboardKPI(conversionCurrent, conversionPrevious),
		KpiSparklines:  buildKpiSparklines(revenueSeries),
		RevenueSeries:  revenueSeries,
		OrdersByStatus: statusRows,
		RecentActivity: buildRecentActivity(activityOrders, auditLogs),
		AiInsights:     buildAiInsights(kpis, toDashboardKPI(conversionCurrent, conversionPrevious), *platform),
		PlatformHealth: platformHealth,
		RecentOrders:   recent,
		TopProducts:    topProducts,
		LowStockProducts: lowStockProducts,
		Platform:       *platform,
	}, nil
}

func toDashboardKPI(current, previous float64) dto.AdminDashboardKPI {
	k := dashboardKPI(current, previous)
	return dto.AdminDashboardKPI{
		Value:         k.Value,
		PreviousValue: k.PreviousValue,
		ChangePercent: k.ChangePercent,
	}
}

package admin

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

// GetSalesAnalytics builds the interactive sales analytics dashboard response.
func (q *Queries) GetSalesAnalytics(ctx context.Context, filters dto.AdminSalesAnalyticsFilters) (*dto.AdminSalesAnalyticsResponse, error) {
	days := dashboardPeriodDays(filters.Period)
	periodLabel := fmt.Sprintf("%dd", days)
	if filters.Period != "" {
		periodLabel = filters.Period
	}

	now := time.Now().UTC()
	currentStart := now.AddDate(0, 0, -days)
	previousStart := currentStart.AddDate(0, 0, -days)

	currentRevenue, err := q.repo.SumRevenueSince(ctx, currentStart)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	previousRevenue, err := q.repo.SumRevenueBetween(ctx, previousStart, currentStart)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	currentProfitTotals, err := q.repo.SumProfitSince(ctx, currentStart)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	previousProfitTotals, err := q.repo.SumProfitBetween(ctx, previousStart, currentStart)
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

	currentCustomers, err := q.repo.CountUsersSince(ctx, currentStart)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	previousCustomers, err := q.repo.CountUsersBetween(ctx, previousStart, currentStart)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	currentPaidOrders, err := q.repo.CountPaidOrdersSince(ctx, currentStart)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	previousPaidOrders, err := q.repo.CountPaidOrdersBetween(ctx, previousStart, currentStart)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	currentAOV := 0.0
	if currentPaidOrders > 0 {
		currentAOV = currentRevenue / float64(currentPaidOrders)
	}
	previousAOV := 0.0
	if previousPaidOrders > 0 {
		previousAOV = previousRevenue / float64(previousPaidOrders)
	}

	revenueRows, err := q.repo.DailyRevenueSeries(ctx, currentStart)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	profitRows, err := q.repo.DailyProfitSeries(ctx, currentStart)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	statusCounts, err := q.repo.OrderStatusCounts(ctx, currentStart)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	topRows, err := q.repo.TopProducts(ctx, currentStart, 8)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	segmentRows, err := q.repo.CustomerSegmentCounts(ctx)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	funnelRows, err := q.repo.OrderFunnelCounts(ctx, currentStart)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	cohortRows, err := q.repo.MonthlyCohorts(ctx, currentStart.AddDate(0, -6, 0), 6)
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

	segments := make([]dto.AdminSalesSegmentCount, len(segmentRows))
	for i, row := range segmentRows {
		segments[i] = dto.AdminSalesSegmentCount{
			Segment: row.Segment,
			Count:   row.Count,
		}
	}

	cohorts := make([]dto.AdminSalesCohortRow, len(cohortRows))
	for i, row := range cohortRows {
		repeatRate := 0.0
		if row.Customers > 0 {
			repeatRate = math.Round((float64(row.RepeatCustomers)/float64(row.Customers))*1000) / 10
		}
		cohorts[i] = dto.AdminSalesCohortRow{
			Cohort:     row.Cohort,
			Customers:  row.Customers,
			RepeatRate: repeatRate,
			Revenue:    row.Revenue,
		}
	}

	return &dto.AdminSalesAnalyticsResponse{
		Period:      periodLabel,
		GeneratedAt: now,
		KPIs: dto.AdminSalesAnalyticsKPIs{
			Revenue:       toDashboardKPI(currentRevenue, previousRevenue),
			Profit:        toDashboardKPI(currentProfitTotals.Profit, previousProfitTotals.Profit),
			Orders:        toDashboardKPI(float64(currentOrders), float64(previousOrders)),
			Customers:     toDashboardKPI(float64(currentCustomers), float64(previousCustomers)),
			AvgOrderValue: toDashboardKPI(currentAOV, previousAOV),
		},
		RevenueSeries:    buildAnalyticsRevenueSeries(now, days, revenueRows),
		ProfitSeries:     buildAnalyticsProfitSeries(now, days, profitRows),
		OrdersByStatus:   statusCounts,
		TopProducts:      topProducts,
		CustomerSegments: segments,
		OrderFunnel:      buildOrderFunnel(funnelRows),
		Cohorts:          cohorts,
	}, nil
}

func buildAnalyticsRevenueSeries(now time.Time, days int, rows []postgres.RevenueDailyRow) []dto.AdminDashboardSeriesPoint {
	rowMap := make(map[string]dto.AdminDashboardSeriesPoint, len(rows))
	for _, row := range rows {
		key := row.Date.Format(time.DateOnly)
		rowMap[key] = dto.AdminDashboardSeriesPoint{
			Date:    key,
			Revenue: row.Revenue,
			Orders:  row.Orders,
		}
	}

	series := make([]dto.AdminDashboardSeriesPoint, 0, days)
	for i := days - 1; i >= 0; i-- {
		day := now.AddDate(0, 0, -i).Format(time.DateOnly)
		if point, ok := rowMap[day]; ok {
			series = append(series, point)
			continue
		}
		series = append(series, dto.AdminDashboardSeriesPoint{Date: day})
	}
	return series
}

func buildAnalyticsProfitSeries(now time.Time, days int, rows []postgres.ProfitDailyRow) []dto.AdminSalesProfitPoint {
	rowMap := make(map[string]postgres.ProfitDailyRow, len(rows))
	for _, row := range rows {
		rowMap[row.Date.Format(time.DateOnly)] = row
	}

	series := make([]dto.AdminSalesProfitPoint, 0, days)
	for i := days - 1; i >= 0; i-- {
		day := now.AddDate(0, 0, -i).Format(time.DateOnly)
		point := rowMap[day]
		series = append(series, dto.AdminSalesProfitPoint{
			Date:    day,
			Revenue: point.Revenue,
			Cost:    point.Cost,
			Profit:  point.Profit,
		})
	}
	return series
}

func buildOrderFunnel(rows []postgres.OrderFunnelRow) []dto.AdminSalesFunnelStep {
	statusMap := make(map[string]int64, len(rows))
	var total int64
	for _, row := range rows {
		statusMap[row.Status] = row.Count
		total += row.Count
	}

	steps := []struct {
		step   string
		status string
	}{
		{step: "Created", status: ""},
		{step: "Paid", status: constants.OrderStatusPaid},
		{step: "Shipped", status: constants.OrderStatusShipped},
		{step: "Delivered", status: constants.OrderStatusDelivered},
	}

	out := make([]dto.AdminSalesFunnelStep, 0, len(steps))
	for _, s := range steps {
		var count int64
		if s.status == "" {
			count = total
		} else {
			count = statusMap[s.status]
		}
		rate := 0.0
		if total > 0 {
			rate = math.Round((float64(count)/float64(total))*1000) / 10
		}
		out = append(out, dto.AdminSalesFunnelStep{
			Step:  s.step,
			Count: count,
			Rate:  rate,
		})
	}
	return out
}

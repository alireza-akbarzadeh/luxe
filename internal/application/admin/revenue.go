package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

// GetRevenueReport builds the admin revenue report response.
func (q *Queries) GetRevenueReport(ctx context.Context, filters dto.AdminRevenueReportFilters) (*dto.AdminRevenueReportResponse, error) {
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

	currentOrders, err := q.repo.CountOrdersSince(ctx, currentStart)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	previousOrders, err := q.repo.CountOrdersBetween(ctx, previousStart, currentStart)
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

	currentAvgDaily := currentRevenue / float64(days)
	previousAvgDaily := previousRevenue / float64(days)

	rows, err := q.repo.DailyRevenueSeries(ctx, currentStart)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	rowMap := make(map[string]struct {
		Revenue    float64
		Orders     int64
		PaidOrders int64
	}, len(rows))
	for _, row := range rows {
		rowMap[row.Date.Format(time.DateOnly)] = struct {
			Revenue    float64
			Orders     int64
			PaidOrders int64
		}{Revenue: row.Revenue, Orders: row.Orders, PaidOrders: row.PaidOrders}
	}

	daily := make([]dto.AdminRevenueDailyRow, 0, days)
	for i := days - 1; i >= 0; i-- {
		dayKey := now.AddDate(0, 0, -i).Format(time.DateOnly)
		if row, ok := rowMap[dayKey]; ok {
			aov := 0.0
			if row.PaidOrders > 0 {
				aov = row.Revenue / float64(row.PaidOrders)
			}
			daily = append(daily, dto.AdminRevenueDailyRow{
				Date:          dayKey,
				Revenue:       row.Revenue,
				Orders:        row.Orders,
				PaidOrders:    row.PaidOrders,
				AvgOrderValue: aov,
			})
			continue
		}
		daily = append(daily, dto.AdminRevenueDailyRow{Date: dayKey})
	}

	return &dto.AdminRevenueReportResponse{
		Period:      periodLabel,
		GeneratedAt: now,
		Summary: dto.AdminRevenueReportSummary{
			Revenue:         toDashboardKPI(currentRevenue, previousRevenue),
			Orders:          toDashboardKPI(float64(currentOrders), float64(previousOrders)),
			AvgOrderValue:   toDashboardKPI(currentAOV, previousAOV),
			AvgDailyRevenue: toDashboardKPI(currentAvgDaily, previousAvgDaily),
		},
		Daily: daily,
	}, nil
}

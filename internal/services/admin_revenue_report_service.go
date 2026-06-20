package services

import (
	"context"
	"fmt"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
)

func (s *adminService) GetRevenueReport(ctx context.Context, filters dto.AdminRevenueReportFilters) (*dto.AdminRevenueReportResponse, error) {
	days := dashboardPeriodDays(filters.Period)
	periodLabel := fmt.Sprintf("%dd", days)
	if filters.Period != "" {
		periodLabel = filters.Period
	}

	now := time.Now().UTC()
	currentStart := now.AddDate(0, 0, -days)
	previousStart := currentStart.AddDate(0, 0, -days)

	db := s.db.WithContext(ctx)

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

	var currentPaidOrders int64
	if err := db.Model(&models.Order{}).
		Where("created_at >= ? AND status IN ?", currentStart, revenueOrderStatuses).
		Count(&currentPaidOrders).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	var previousPaidOrders int64
	if err := db.Model(&models.Order{}).
		Where("created_at >= ? AND created_at < ? AND status IN ?", previousStart, currentStart, revenueOrderStatuses).
		Count(&previousPaidOrders).Error; err != nil {
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

	type dailyRow struct {
		Date       time.Time
		Revenue    float64
		Orders     int64
		PaidOrders int64
	}
	var rows []dailyRow
	if err := db.Raw(`
		SELECT created_at::date AS date,
		       COALESCE(SUM(CASE WHEN status IN ? THEN total_amount ELSE 0 END), 0) AS revenue,
		       COUNT(*) AS orders,
		       COUNT(CASE WHEN status IN ? THEN 1 END) AS paid_orders
		FROM orders
		WHERE created_at >= ? AND deleted_at IS NULL
		GROUP BY created_at::date
		ORDER BY date ASC`, revenueOrderStatuses, revenueOrderStatuses, currentStart).Scan(&rows).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	rowMap := make(map[string]dailyRow, len(rows))
	for _, row := range rows {
		rowMap[row.Date.Format(time.DateOnly)] = row
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
			Revenue:         dashboardKPI(currentRevenue, previousRevenue),
			Orders:          dashboardKPI(float64(currentOrders), float64(previousOrders)),
			AvgOrderValue:   dashboardKPI(currentAOV, previousAOV),
			AvgDailyRevenue: dashboardKPI(currentAvgDaily, previousAvgDaily),
		},
		Daily: daily,
	}, nil
}

package services

import (
	"context"
	"fmt"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
)

func startOfDayUTC(now time.Time) time.Time {
	y, m, d := now.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func salesFeedEventType(status string) string {
	switch status {
	case constants.OrderStatusCancelled:
		return "cancellation"
	case constants.OrderStatusShipped:
		return "shipment"
	case constants.OrderStatusRefunded:
		return "refund"
	case constants.OrderStatusPaid:
		return "payment"
	case constants.OrderStatusPending:
		return "new_order"
	default:
		return "status_change"
	}
}

func salesFeedEventTitle(order models.Order, customerName string) string {
	switch order.Status {
	case constants.OrderStatusCancelled:
		return "Order cancelled"
	case constants.OrderStatusShipped:
		return "Order shipped"
	case constants.OrderStatusRefunded:
		return "Refund issued"
	case constants.OrderStatusPaid:
		return "Payment received"
	case constants.OrderStatusPending:
		return fmt.Sprintf("New order %s", order.OrderNumber)
	default:
		return fmt.Sprintf("Order %s updated", order.OrderNumber)
	}
}

func salesFeedEventSubtitle(order models.Order, customerName string) string {
	switch order.Status {
	case constants.OrderStatusCancelled:
		return fmt.Sprintf("%s was cancelled", order.OrderNumber)
	case constants.OrderStatusShipped:
		return fmt.Sprintf("%s for %s is on its way", order.OrderNumber, customerName)
	case constants.OrderStatusRefunded:
		return fmt.Sprintf("$%.2f refunded · %s", order.TotalAmount, order.OrderNumber)
	case constants.OrderStatusPaid:
		return fmt.Sprintf("%s paid $%.2f for %s", customerName, order.TotalAmount, order.OrderNumber)
	default:
		return fmt.Sprintf("%s · %s · $%.2f", order.OrderNumber, order.Status, order.TotalAmount)
	}
}

func (s *adminService) GetSalesFeedSnapshot(ctx context.Context) (*dto.AdminSalesFeedSnapshotResponse, error) {
	now := time.Now().UTC()
	todayStart := startOfDayUTC(now)
	db := s.db.WithContext(ctx)

	var totalOrdersToday int64
	if err := db.Model(&models.Order{}).
		Where("created_at >= ?", todayStart).
		Count(&totalOrdersToday).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	var totalRevenueToday float64
	if err := db.Model(&models.Order{}).
		Where("created_at >= ? AND status IN ?", todayStart, revenueOrderStatuses).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&totalRevenueToday).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	var statusRows []dto.AdminDashboardStatusCount
	if err := db.Model(&models.Order{}).
		Select("status, COUNT(*) as count").
		Where("created_at >= ?", todayStart).
		Group("status").
		Order("count DESC").
		Scan(&statusRows).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	type hourRow struct {
		Hour    time.Time
		Revenue float64
		Orders  int64
	}
	var hourRows []hourRow
	if err := db.Raw(`
		SELECT date_trunc('hour', created_at) AS hour,
		       COALESCE(SUM(CASE WHEN status IN ? THEN total_amount ELSE 0 END), 0) AS revenue,
		       COUNT(*) AS orders
		FROM orders
		WHERE created_at >= ? AND deleted_at IS NULL
		GROUP BY date_trunc('hour', created_at)
		ORDER BY hour ASC`, revenueOrderStatuses, todayStart).Scan(&hourRows).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	revenueSeries := make([]dto.AdminDashboardSeriesPoint, len(hourRows))
	for i, row := range hourRows {
		revenueSeries[i] = dto.AdminDashboardSeriesPoint{
			Date:    row.Hour.Format("15:04:05"),
			Revenue: row.Revenue,
			Orders:  row.Orders,
		}
	}

	var recentOrders []models.Order
	if err := db.Preload("User").
		Where("created_at >= ?", todayStart).
		Order("created_at DESC").
		Limit(20).
		Find(&recentOrders).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	recentEvents := make([]dto.AdminSalesFeedEvent, len(recentOrders))
	for i, order := range recentOrders {
		customerName := fmt.Sprintf("%s %s", order.User.FirstName, order.User.LastName)
		if customerName == " " {
			customerName = order.User.Email
		}
		if customerName == "" {
			customerName = "Customer"
		}

		recentEvents[i] = dto.AdminSalesFeedEvent{
			ID:        fmt.Sprintf("%d", order.ID),
			Type:      salesFeedEventType(order.Status),
			Title:     salesFeedEventTitle(order, customerName),
			Subtitle:  salesFeedEventSubtitle(order, customerName),
			Amount:    order.TotalAmount,
			Timestamp: order.CreatedAt.UnixMilli(),
		}
	}

	return &dto.AdminSalesFeedSnapshotResponse{
		TotalOrdersToday:  totalOrdersToday,
		TotalRevenueToday: totalRevenueToday,
		StatusCounts:      statusRows,
		RecentEvents:      recentEvents,
		RevenueSeries:     revenueSeries,
		GeneratedAt:       now,
	}, nil
}

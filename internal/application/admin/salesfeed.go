package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
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

func salesFeedEventTitle(order models.Order) string {
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

// GetSalesFeedSnapshot builds the admin sales feed snapshot response.
func (q *Queries) GetSalesFeedSnapshot(ctx context.Context) (*dto.AdminSalesFeedSnapshotResponse, error) {
	now := time.Now().UTC()
	todayStart := startOfDayUTC(now)

	totalOrdersToday, err := q.repo.CountOrdersSince(ctx, todayStart)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	totalRevenueToday, err := q.repo.SumRevenueSince(ctx, todayStart)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	statusRows, err := q.repo.OrderStatusCounts(ctx, todayStart)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	hourRows, err := q.repo.HourlyOrderSeries(ctx, todayStart)
	if err != nil {
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

	recentOrders, err := q.repo.OrdersSince(ctx, todayStart, 20)
	if err != nil {
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
			Title:     salesFeedEventTitle(order),
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

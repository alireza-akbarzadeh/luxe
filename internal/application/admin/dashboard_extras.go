package admin

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

const sparklinePointCount = 7

func conversionRate(paid, total int64) float64 {
	if total == 0 {
		return 0
	}
	return math.Round((float64(paid)/float64(total))*1000) / 10
}

func buildKpiSparklines(series []dto.AdminDashboardSeriesPoint) map[string][]float64 {
	n := sparklinePointCount
	if len(series) < n {
		n = len(series)
	}
	if n == 0 {
		return map[string][]float64{
			"revenue":       {},
			"orders":        {},
			"new_customers": {},
			"avg_order_value": {},
		}
	}

	slice := series[len(series)-n:]
	revenue := make([]float64, n)
	orders := make([]float64, n)
	aov := make([]float64, n)

	for i, point := range slice {
		revenue[i] = point.Revenue
		orders[i] = float64(point.Orders)
		if point.Orders > 0 {
			aov[i] = math.Round((point.Revenue/float64(point.Orders))*100) / 100
		}
	}

	return map[string][]float64{
		"revenue":         revenue,
		"orders":          orders,
		"new_customers":   orders,
		"avg_order_value": aov,
	}
}

func buildRecentActivity(orders []models.Order, auditLogs []models.AuditLog) []dto.AdminDashboardActivity {
	activities := make([]dto.AdminDashboardActivity, 0, len(orders)+len(auditLogs))

	for _, order := range orders {
		customerName := strings.TrimSpace(fmt.Sprintf("%s %s", order.User.FirstName, order.User.LastName))
		if customerName == "" {
			customerName = order.User.Email
		}
		activities = append(activities, dto.AdminDashboardActivity{
			ID:          fmt.Sprintf("order-%d", order.ID),
			Type:        "order",
			Title:       fmt.Sprintf("Order %s", order.OrderNumber),
			Description: fmt.Sprintf("%s · %s · $%.2f", order.Status, customerName, order.TotalAmount),
			Actor:       customerName,
			Href:        fmt.Sprintf("/dashboard/orders/%d", order.ID),
			CreatedAt:   order.CreatedAt,
		})
	}

	for _, log := range auditLogs {
		actor := log.User.Email
		if actor == "" {
			actor = fmt.Sprintf("User %d", log.UserID)
		}
		activities = append(activities, dto.AdminDashboardActivity{
			ID:          fmt.Sprintf("audit-%d", log.ID),
			Type:        "audit",
			Title:       fmt.Sprintf("%s %s", strings.ToUpper(log.Action), log.Resource),
			Description: log.Path,
			Actor:       actor,
			Href:        "",
			CreatedAt:   log.CreatedAt,
		})
	}

	sort.Slice(activities, func(i, j int) bool {
		return activities[i].CreatedAt.After(activities[j].CreatedAt)
	})

	if len(activities) > 12 {
		activities = activities[:12]
	}
	return activities
}

func buildAiInsights(
	kpis dto.AdminDashboardKPIs,
	conversion dto.AdminDashboardKPI,
	platform dto.AdminStatsResponse,
) []dto.AdminDashboardInsight {
	insights := make([]dto.AdminDashboardInsight, 0, 4)

	if kpis.Revenue.ChangePercent < -5 {
		insights = append(insights, dto.AdminDashboardInsight{
			ID:          "revenue-decline",
			Severity:    "warning",
			Title:       "Revenue is trending down",
			Body:        fmt.Sprintf("Gross revenue fell %.1f%% versus the previous period. Review campaigns and top sellers.", math.Abs(kpis.Revenue.ChangePercent)),
			ActionLabel: "View revenue report",
			ActionHref:  "/dashboard/reports/revenue",
		})
	} else if kpis.Revenue.ChangePercent > 10 {
		insights = append(insights, dto.AdminDashboardInsight{
			ID:       "revenue-growth",
			Severity: "success",
			Title:    "Strong revenue momentum",
			Body:     fmt.Sprintf("Revenue grew %.1f%% compared to the previous period.", kpis.Revenue.ChangePercent),
		})
	}

	if platform.LowStockProducts > 0 {
		insights = append(insights, dto.AdminDashboardInsight{
			ID:          "low-stock",
			Severity:    "warning",
			Title:       "Inventory attention needed",
			Body:        fmt.Sprintf("%d products are at or below their low-stock threshold.", platform.LowStockProducts),
			ActionLabel: "Review inventory",
			ActionHref:  "/dashboard/inventory",
		})
	}

	if platform.PendingOrders > 0 {
		insights = append(insights, dto.AdminDashboardInsight{
			ID:          "pending-orders",
			Severity:    "info",
			Title:       "Orders awaiting fulfillment",
			Body:        fmt.Sprintf("%d orders are pending payment or processing.", platform.PendingOrders),
			ActionLabel: "Open orders",
			ActionHref:  "/dashboard/orders",
		})
	}

	if conversion.Value < 50 && conversion.Value > 0 {
		insights = append(insights, dto.AdminDashboardInsight{
			ID:       "conversion-low",
			Severity: "warning",
			Title:    "Checkout conversion is soft",
			Body:     fmt.Sprintf("Only %.1f%% of orders in this period reached a paid state.", conversion.Value),
		})
	}

	if len(insights) == 0 {
		insights = append(insights, dto.AdminDashboardInsight{
			ID:       "all-clear",
			Severity: "success",
			Title:    "Platform looks healthy",
			Body:     "No urgent issues detected for this period. Keep monitoring KPIs and inventory levels.",
		})
	}

	return insights
}

func buildPlatformHealth(
	ctx context.Context,
	q *Queries,
	platform dto.AdminStatsResponse,
) (dto.AdminDashboardHealth, error) {
	latency, err := q.repo.PingDB(ctx)
	if err != nil {
		return dto.AdminDashboardHealth{
			Status:  "critical",
			Message: "Database connectivity check failed",
		}, nil
	}

	since := time.Now().UTC().Add(-24 * time.Hour)
	failed, err := q.repo.CountFailedWebhooksSince(ctx, since)
	if err != nil {
		return dto.AdminDashboardHealth{}, err
	}
	totalWebhooks, err := q.repo.CountWebhooksSince(ctx, since)
	if err != nil {
		return dto.AdminDashboardHealth{}, err
	}

	errorRate := 0.0
	if totalWebhooks > 0 {
		errorRate = math.Round((float64(failed)/float64(totalWebhooks))*1000) / 10
	}

	status := "healthy"
	message := "All systems operational"
	if platform.PendingOrders > 25 || platform.LowStockProducts > 10 {
		status = "degraded"
		message = "Elevated pending orders or inventory alerts"
	}
	if latency > 500 || errorRate > 10 {
		status = "degraded"
		message = "Elevated latency or webhook error rate"
	}
	if latency > 2000 || errorRate > 25 {
		status = "critical"
		message = "Critical latency or error rate detected"
	}

	return dto.AdminDashboardHealth{
		Status:       status,
		APILatencyMs: latency,
		ErrorRate:    errorRate,
		QueueDepth:   platform.PendingOrders,
		Message:      message,
	}, nil
}

func orderActivityType(status string) string {
	switch status {
	case constants.OrderStatusCancelled:
		return "cancellation"
	case constants.OrderStatusShipped:
		return "shipment"
	default:
		return "order"
	}
}

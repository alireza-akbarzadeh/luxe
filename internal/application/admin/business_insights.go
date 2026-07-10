package admin

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

// GetBusinessInsights builds categorized AI-style business insights for the admin insights page.
func (q *Queries) GetBusinessInsights(ctx context.Context, filters dto.AdminBusinessInsightsFilters) (*dto.AdminBusinessInsightsResponse, error) {
	days := dashboardPeriodDays(filters.Period)
	periodLabel := filters.Period
	if periodLabel == "" {
		periodLabel = "30d"
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
	revenueKPI := dashboardKPI(currentRevenue, previousRevenue)

	currentOrders, err := q.repo.CountOrdersSince(ctx, currentStart)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	currentPaid, err := q.repo.CountPaidOrdersSince(ctx, currentStart)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	conversion := conversionRate(currentPaid, currentOrders)

	atRiskCount, err := q.repo.CountCustomersBySegment(ctx, "at_risk")
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	vipCount, err := q.repo.CountCustomersBySegment(ctx, "vip")
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	topProducts, err := q.repo.TopProducts(ctx, currentStart, 3)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	lowStock, err := q.repo.LowStockProducts(ctx, 5)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	insights := buildBusinessInsights(
		revenueKPI,
		conversion,
		*platform,
		atRiskCount,
		vipCount,
		topProducts,
		lowStock,
	)

	revenueTrend := "stable"
	if revenueKPI.ChangePercent > 5 {
		revenueTrend = "up"
	} else if revenueKPI.ChangePercent < -5 {
		revenueTrend = "down"
	}

	riskCount := 0
	opportunityCount := 0
	for _, insight := range insights {
		switch insight.Severity {
		case "warning", "critical":
			riskCount++
		case "success":
			if insight.Category == "products" || insight.Category == "revenue" {
				opportunityCount++
			}
		}
	}

	return &dto.AdminBusinessInsightsResponse{
		Period:           periodLabel,
		GeneratedAt:      now,
		RevenueTrend:     revenueTrend,
		RiskCount:        riskCount,
		OpportunityCount: opportunityCount,
		Insights:         insights,
	}, nil
}

func buildBusinessInsights(
	revenue KPI,
	conversion float64,
	platform dto.AdminStatsResponse,
	atRiskCount int64,
	vipCount int64,
	topProducts []postgres.TopProductRow,
	lowStock []models.Product,
) []dto.AdminBusinessInsight {
	insights := make([]dto.AdminBusinessInsight, 0, 8)

	if revenue.ChangePercent < -5 {
		insights = append(insights, dto.AdminBusinessInsight{
			ID:          "revenue-decline",
			Category:    "revenue",
			Severity:    "warning",
			Title:       "Revenue is trending down",
			Body:        fmt.Sprintf("Gross revenue fell %.1f%% versus the previous period. Review promotions and top sellers.", math.Abs(revenue.ChangePercent)),
			ActionLabel: "Open sales analytics",
			ActionHref:  "/dashboard/analytics",
		})
	} else if revenue.ChangePercent > 10 {
		insights = append(insights, dto.AdminBusinessInsight{
			ID:          "revenue-growth",
			Category:    "revenue",
			Severity:    "success",
			Title:       "Strong revenue momentum",
			Body:        fmt.Sprintf("Revenue grew %.1f%% compared to the previous period — consider scaling winning campaigns.", revenue.ChangePercent),
			ActionLabel: "View revenue report",
			ActionHref:  "/dashboard/reports/revenue",
		})
	} else {
		insights = append(insights, dto.AdminBusinessInsight{
			ID:          "revenue-stable",
			Category:    "revenue",
			Severity:    "info",
			Title:       "Revenue is stable",
			Body:        fmt.Sprintf("Revenue changed %.1f%% versus the prior period. Monitor daily trends for early signals.", revenue.ChangePercent),
			ActionLabel: "Sales analytics",
			ActionHref:  "/dashboard/analytics",
		})
	}

	if len(topProducts) > 0 {
		top := topProducts[0]
		insights = append(insights, dto.AdminBusinessInsight{
			ID:          "top-product",
			Category:    "products",
			Severity:    "success",
			Title:       "Top product opportunity",
			Body:        fmt.Sprintf("%s (%s) drove $%.0f in revenue — ensure stock and feature it in promotions.", top.Name, top.SKU, top.Revenue),
			ActionLabel: "Manage product",
			ActionHref:  fmt.Sprintf("/dashboard/products/edit/%d", top.ID),
		})
	}

	if platform.LowStockProducts > 0 && len(lowStock) > 0 {
		insights = append(insights, dto.AdminBusinessInsight{
			ID:          "inventory-risk",
			Category:    "inventory",
			Severity:    "warning",
			Title:       "Inventory risk detected",
			Body:        fmt.Sprintf("%d products are at or below threshold. %s has only %d units left.", platform.LowStockProducts, lowStock[0].Name, lowStock[0].Stock),
			ActionLabel: "Review inventory",
			ActionHref:  "/dashboard/inventory",
		})
	}

	if atRiskCount > 0 {
		insights = append(insights, dto.AdminBusinessInsight{
			ID:          "churn-risk",
			Category:    "churn",
			Severity:    "critical",
			Title:       "Customers at churn risk",
			Body:        fmt.Sprintf("%d customers are tagged at-risk. Launch win-back email campaigns or targeted offers.", atRiskCount),
			ActionLabel: "View customers",
			ActionHref:  "/dashboard/customers",
		})
	}

	if vipCount > 0 {
		insights = append(insights, dto.AdminBusinessInsight{
			ID:          "vip-retention",
			Category:    "churn",
			Severity:    "info",
			Title:       "Protect VIP relationships",
			Body:        fmt.Sprintf("%d VIP customers contribute disproportionate revenue — prioritize support response times.", vipCount),
			ActionLabel: "CRM segments",
			ActionHref:  "/dashboard/customers",
		})
	}

	if conversion < 50 && conversion > 0 {
		insights = append(insights, dto.AdminBusinessInsight{
			ID:          "checkout-conversion",
			Category:    "revenue",
			Severity:    "warning",
			Title:       "Checkout conversion is soft",
			Body:        fmt.Sprintf("Only %.1f%% of orders reached a paid state this period. Review abandoned carts and payment failures.", conversion),
			ActionLabel: "View orders",
			ActionHref:  "/dashboard/orders",
		})
	}

	if platform.PendingOrders > 0 {
		insights = append(insights, dto.AdminBusinessInsight{
			ID:          "fulfillment-backlog",
			Category:    "inventory",
			Severity:    "info",
			Title:       "Fulfillment backlog",
			Body:        fmt.Sprintf("%d orders need processing or payment confirmation.", platform.PendingOrders),
			ActionLabel: "Fulfillment hub",
			ActionHref:  "/dashboard/fulfillment",
		})
	}

	return insights
}

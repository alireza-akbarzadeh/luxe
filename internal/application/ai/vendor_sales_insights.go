package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	apporder "github.com/alireza-akbarzadeh/luxe/internal/application/order"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

const TaskVendorSalesInsights = "vendor_sales_insights"

// VendorSalesSnapshot aggregates store sales facts for AI or rule-based insights.
type VendorSalesSnapshot struct {
	StoreName   string
	PeriodDays  int
	Current     apporder.VendorSalesSummary
	Prior       apporder.VendorSalesSummary
	Daily       []apporder.VendorDailySales
	TopProducts []apporder.VendorTopProduct
}

// VendorSalesInsightsQueries loads store sales facts for analytics.
type VendorSalesInsightsQueries interface {
	LoadSalesSnapshot(ctx context.Context, storeID uint, days int) (*VendorSalesSnapshot, error)
}

type vendorSalesInsightsPayload struct {
	Summary         string   `json:"summary"`
	Highlights      []string `json:"highlights"`
	Recommendations []string `json:"recommendations"`
	Warnings        []string `json:"warnings"`
}

// VendorSalesInsights returns AI-generated or rule-based sales analytics for sellers.
func (s *Service) VendorSalesInsights(
	ctx context.Context,
	storeID uint,
	days int,
	subjectKey string,
	queries VendorSalesInsightsQueries,
) (*dto.AiVendorSalesInsightsResponse, error) {
	if queries == nil {
		return nil, utils.ErrInternal(fmt.Errorf("vendor sales insights queries not configured"))
	}
	if storeID == 0 {
		return nil, utils.ErrBadRequest("store id required")
	}
	if days <= 0 {
		days = 30
	}
	if days > 90 {
		days = 90
	}

	snapshot, err := queries.LoadSalesSnapshot(ctx, storeID, days)
	if err != nil {
		return nil, err
	}

	metrics := buildVendorSalesMetrics(snapshot)
	response := &dto.AiVendorSalesInsightsResponse{
		PeriodDays:  days,
		Metrics:     metrics,
		DailySeries: mapDailySeries(snapshot.Daily),
		TopProducts: mapTopProducts(snapshot.TopProducts),
		Sources:     []string{"orders", "order_items"},
	}

	if !s.Enabled() {
		fallback := fallbackVendorSalesInsights(snapshot, metrics)
		response.AiEnabled = false
		response.Summary = fallback.Summary
		response.Highlights = fallback.Highlights
		response.Recommendations = fallback.Recommendations
		response.Warnings = fallback.Warnings
		return response, nil
	}

	if !s.adminRL.allow("vendor_sales_insights:" + subjectKey) {
		return nil, utils.ErrTooManyRequests()
	}

	facts := vendorSalesFacts(snapshot)
	system := `You are a concise e-commerce sales analyst for marketplace sellers.
Respond with JSON only using this exact shape:
{"summary":"...","highlights":["..."],"recommendations":["..."],"warnings":["..."]}
Rules:
- summary: 2-3 sentences on revenue momentum and order volume for the period.
- highlights: 2-4 positive or notable trends grounded in the metrics.
- recommendations: 2-3 actionable merchandising or fulfillment suggestions.
- warnings: 0-2 risks (declining revenue, concentration in one SKU) — empty array if none.
Use ONLY the sales facts below. Do not invent traffic, ad spend, or review data.
Plain text only — no markdown.

Sales facts:
` + facts

	user := fmt.Sprintf("Analyze the last %d days of store sales.", days)
	resp, err := s.complete(ctx, system, user, 900)
	if err != nil {
		fallback := fallbackVendorSalesInsights(snapshot, metrics)
		response.AiEnabled = false
		response.Summary = "AI is temporarily unavailable. Showing rule-based sales insights from your order data."
		response.Highlights = fallback.Highlights
		response.Recommendations = fallback.Recommendations
		response.Warnings = fallback.Warnings
		return response, nil
	}

	parsed, err := parseVendorSalesInsightsJSON(resp.Content)
	if err != nil {
		fallback := fallbackVendorSalesInsights(snapshot, metrics)
		response.AiEnabled = false
		response.Summary = fallback.Summary
		response.Highlights = fallback.Highlights
		response.Recommendations = fallback.Recommendations
		response.Warnings = fallback.Warnings
		return response, nil
	}

	response.AiEnabled = true
	response.Summary = strings.TrimSpace(parsed.Summary)
	response.Highlights = sanitizeStringList(parsed.Highlights)
	response.Recommendations = sanitizeStringList(parsed.Recommendations)
	response.Warnings = sanitizeStringList(parsed.Warnings)
	return response, nil
}

func buildVendorSalesMetrics(snapshot *VendorSalesSnapshot) dto.VendorSalesMetrics {
	if snapshot == nil {
		return dto.VendorSalesMetrics{}
	}
	return dto.VendorSalesMetrics{
		Revenue:         snapshot.Current.Revenue,
		OrderCount:      snapshot.Current.OrderCount,
		UnitsSold:       snapshot.Current.UnitsSold,
		AvgOrderValue:   snapshot.Current.AvgOrderValue,
		RevenueChangePct: pctChange(snapshot.Current.Revenue, snapshot.Prior.Revenue),
		OrdersChangePct:  pctChange(float64(snapshot.Current.OrderCount), float64(snapshot.Prior.OrderCount)),
	}
}

func mapDailySeries(series []apporder.VendorDailySales) []dto.VendorDailySalesPoint {
	points := make([]dto.VendorDailySalesPoint, len(series))
	for i, day := range series {
		points[i] = dto.VendorDailySalesPoint{
			Date:    day.Date,
			Revenue: day.Revenue,
			Orders:  day.Orders,
		}
	}
	return points
}

func mapTopProducts(products []apporder.VendorTopProduct) []dto.VendorTopProductSales {
	rows := make([]dto.VendorTopProductSales, len(products))
	for i, product := range products {
		rows[i] = dto.VendorTopProductSales{
			ProductID: product.ProductID,
			Name:      product.Name,
			Revenue:   product.Revenue,
			Units:     product.Units,
		}
	}
	return rows
}

func pctChange(current, prior float64) float64 {
	if prior == 0 {
		if current > 0 {
			return 100
		}
		return 0
	}
	change := ((current - prior) / prior) * 100
	return math.Round(change*10) / 10
}

func vendorSalesFacts(snapshot *VendorSalesSnapshot) string {
	if snapshot == nil {
		return "No sales data available."
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Store: %s\n", snapshot.StoreName)
	fmt.Fprintf(&b, "Period: last %d days\n", snapshot.PeriodDays)
	fmt.Fprintf(
		&b,
		"Current revenue: %.2f | orders: %d | units: %d | AOV: %.2f\n",
		snapshot.Current.Revenue,
		snapshot.Current.OrderCount,
		snapshot.Current.UnitsSold,
		snapshot.Current.AvgOrderValue,
	)
	fmt.Fprintf(
		&b,
		"Prior period revenue: %.2f | orders: %d\n",
		snapshot.Prior.Revenue,
		snapshot.Prior.OrderCount,
	)

	if len(snapshot.TopProducts) > 0 {
		b.WriteString("Top products:\n")
		for _, product := range snapshot.TopProducts {
			fmt.Fprintf(&b, "- %s revenue=%.2f units=%d\n", product.Name, product.Revenue, product.Units)
		}
	}

	return b.String()
}

func fallbackVendorSalesInsights(snapshot *VendorSalesSnapshot, metrics dto.VendorSalesMetrics) *vendorSalesInsightsPayload {
	highlights := make([]string, 0, 4)
	recommendations := make([]string, 0, 3)
	warnings := make([]string, 0, 2)

	if metrics.Revenue > 0 {
		highlights = append(highlights, fmt.Sprintf("Generated $%.2f in revenue across %d orders", metrics.Revenue, metrics.OrderCount))
	}
	if metrics.RevenueChangePct > 5 {
		highlights = append(highlights, fmt.Sprintf("Revenue is up %.1f%% versus the prior period", metrics.RevenueChangePct))
	} else if metrics.RevenueChangePct < -5 {
		warnings = append(warnings, fmt.Sprintf("Revenue declined %.1f%% versus the prior period", math.Abs(metrics.RevenueChangePct)))
	}
	if metrics.AvgOrderValue > 0 {
		highlights = append(highlights, fmt.Sprintf("Average order value is $%.2f", metrics.AvgOrderValue))
	}

	if len(snapshot.TopProducts) > 0 {
		top := snapshot.TopProducts[0]
		highlights = append(highlights, fmt.Sprintf("%s is your top seller with $%.2f revenue", top.Name, top.Revenue))
		if len(snapshot.TopProducts) == 1 && metrics.Revenue > 0 && top.Revenue/metrics.Revenue > 0.7 {
			warnings = append(warnings, "Revenue is highly concentrated in a single product")
		}
	}

	if metrics.OrderCount == 0 {
		recommendations = append(recommendations, "Run a limited-time offer on hero products to convert browsing into first orders")
	} else {
		recommendations = append(recommendations, "Bundle your top seller with complementary SKUs to lift average order value")
	}
	if metrics.UnitsSold > 0 && metrics.OrderCount > 0 && float64(metrics.UnitsSold)/float64(metrics.OrderCount) < 1.2 {
		recommendations = append(recommendations, "Add cross-sell prompts at checkout to increase units per order")
	}

	summary := fmt.Sprintf(
		"%s recorded $%.2f in revenue from %d orders over the last %d days.",
		snapshot.StoreName,
		metrics.Revenue,
		metrics.OrderCount,
		snapshot.PeriodDays,
	)

	return &vendorSalesInsightsPayload{
		Summary:         summary,
		Highlights:      highlights,
		Recommendations: recommendations,
		Warnings:        warnings,
	}
}

func parseVendorSalesInsightsJSON(content string) (*vendorSalesInsightsPayload, error) {
	content = extractJSONObject(content)
	var payload vendorSalesInsightsPayload
	if err := json.Unmarshal([]byte(content), &payload); err != nil {
		return nil, err
	}
	if strings.TrimSpace(payload.Summary) == "" {
		return nil, fmt.Errorf("missing summary")
	}
	return &payload, nil
}

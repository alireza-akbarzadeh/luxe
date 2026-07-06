package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

const TaskVendorInventoryForecast = "vendor_inventory_forecast"

// VendorInventorySnapshot aggregates stock and demand for forecasting.
type VendorInventorySnapshot struct {
	StoreName     string
	PeriodDays    int
	LowStockCount int64
	Products      []VendorProductDemand
}

// VendorInventoryForecastQueries loads inventory demand facts.
type VendorInventoryForecastQueries interface {
	LoadInventorySnapshot(ctx context.Context, storeID uint, days int) (*VendorInventorySnapshot, error)
}

type vendorInventoryForecastPayload struct {
	Summary         string   `json:"summary"`
	Priorities      []string `json:"priorities"`
	Recommendations []string `json:"recommendations"`
	Alerts          []string `json:"alerts"`
}

// VendorInventoryForecast returns AI-generated or rule-based stock forecasts for sellers.
func (s *Service) VendorInventoryForecast(
	ctx context.Context,
	storeID uint,
	days int,
	subjectKey string,
	queries VendorInventoryForecastQueries,
) (*dto.AiVendorInventoryForecastResponse, error) {
	if queries == nil {
		return nil, utils.ErrInternal(fmt.Errorf("vendor inventory forecast queries not configured"))
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

	snapshot, err := queries.LoadInventorySnapshot(ctx, storeID, days)
	if err != nil {
		return nil, err
	}

	forecasts := buildInventoryForecasts(snapshot.Products, days)
	response := &dto.AiVendorInventoryForecastResponse{
		PeriodDays:      days,
		LowStockCount:   snapshot.LowStockCount,
		Forecasts:       forecasts,
		CriticalCount:   countUrgency(forecasts, "critical"),
		WarningCount:    countUrgency(forecasts, "warning"),
		Sources:         []string{"products", "orders", "order_items"},
	}

	if !s.Enabled() {
		fallback := fallbackVendorInventoryForecast(snapshot, forecasts)
		response.AiEnabled = false
		response.Summary = fallback.Summary
		response.Priorities = fallback.Priorities
		response.Recommendations = fallback.Recommendations
		response.Alerts = fallback.Alerts
		return response, nil
	}

	if !s.adminRL.allow("vendor_inventory_forecast:" + subjectKey) {
		return nil, utils.ErrTooManyRequests()
	}

	facts := vendorInventoryFacts(snapshot, forecasts)
	system := `You are a concise inventory planner for marketplace sellers.
Respond with JSON only using this exact shape:
{"summary":"...","priorities":["..."],"recommendations":["..."],"alerts":["..."]}
Rules:
- summary: 2-3 sentences on stock risk and replenishment outlook.
- priorities: 2-4 urgent restock actions grounded in velocity and days-until-stockout.
- recommendations: 2-3 inventory hygiene tips (safety stock, slow movers).
- alerts: 0-3 stockout or overstock risks — empty array if none.
Use ONLY the inventory facts below. Do not invent supplier lead times beyond what is provided.
Plain text only — no markdown.

Inventory facts:
` + facts

	user := fmt.Sprintf("Forecast inventory needs for the next %d days based on recent sales velocity.", days)
	resp, err := s.complete(ctx, system, user, 900)
	if err != nil {
		fallback := fallbackVendorInventoryForecast(snapshot, forecasts)
		response.AiEnabled = false
		response.Summary = "AI is temporarily unavailable. Showing rule-based inventory forecasts from your catalog and sales."
		response.Priorities = fallback.Priorities
		response.Recommendations = fallback.Recommendations
		response.Alerts = fallback.Alerts
		return response, nil
	}

	parsed, err := parseVendorInventoryForecastJSON(resp.Content)
	if err != nil {
		fallback := fallbackVendorInventoryForecast(snapshot, forecasts)
		response.AiEnabled = false
		response.Summary = fallback.Summary
		response.Priorities = fallback.Priorities
		response.Recommendations = fallback.Recommendations
		response.Alerts = fallback.Alerts
		return response, nil
	}

	response.AiEnabled = true
	response.Summary = strings.TrimSpace(parsed.Summary)
	response.Priorities = sanitizeStringList(parsed.Priorities)
	response.Recommendations = sanitizeStringList(parsed.Recommendations)
	response.Alerts = sanitizeStringList(parsed.Alerts)
	return response, nil
}

func buildInventoryForecasts(products []VendorProductDemand, days int) []dto.VendorInventoryForecastItem {
	forecasts := make([]dto.VendorInventoryForecastItem, 0, len(products))
	periodDays := float64(days)
	if periodDays <= 0 {
		periodDays = 30
	}

	for _, product := range products {
		dailyVelocity := float64(product.UnitsSold) / periodDays
		item := dto.VendorInventoryForecastItem{
			ProductID:   product.ProductID,
			Name:        product.Name,
			Stock:       product.Stock,
			UnitsSold:   product.UnitsSold,
			DailyVelocity: math.Round(dailyVelocity*100) / 100,
			Urgency:     "no_demand",
		}

		targetStock := product.LowStockThreshold * 2
		if targetStock < 10 {
			targetStock = 10
		}
		coverageDays := 14.0
		if dailyVelocity > 0 {
			coverageStock := int(math.Ceil(dailyVelocity * coverageDays))
			if coverageStock > targetStock {
				targetStock = coverageStock
			}
			daysLeft := float64(product.Stock) / dailyVelocity
			rounded := math.Round(daysLeft*10) / 10
			item.DaysUntilStockout = &rounded

			switch {
			case product.Stock == 0:
				item.Urgency = "critical"
			case daysLeft < 7:
				item.Urgency = "critical"
			case daysLeft < 14:
				item.Urgency = "warning"
			default:
				item.Urgency = "ok"
			}
		} else if product.Stock <= product.LowStockThreshold {
			item.Urgency = "warning"
		}

		reorder := targetStock - product.Stock
		if reorder < 0 {
			reorder = 0
		}
		item.SuggestedReorderQty = reorder

		forecasts = append(forecasts, item)
	}

	return forecasts
}

func countUrgency(forecasts []dto.VendorInventoryForecastItem, urgency string) int {
	count := 0
	for _, item := range forecasts {
		if item.Urgency == urgency {
			count++
		}
	}
	return count
}

func vendorInventoryFacts(snapshot *VendorInventorySnapshot, forecasts []dto.VendorInventoryForecastItem) string {
	if snapshot == nil {
		return "No inventory data available."
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Store: %s\n", snapshot.StoreName)
	fmt.Fprintf(&b, "Analysis window: %d days\n", snapshot.PeriodDays)
	fmt.Fprintf(&b, "Low-stock SKUs: %d\n", snapshot.LowStockCount)
	fmt.Fprintf(&b, "Tracked products analyzed: %d\n", len(forecasts))

	critical := 0
	for _, item := range forecasts {
		if item.Urgency == "critical" {
			critical++
		}
	}
	fmt.Fprintf(&b, "Critical stockout risk SKUs: %d\n", critical)

	b.WriteString("Top risk SKUs:\n")
	shown := 0
	for _, item := range forecasts {
		if item.Urgency != "critical" && item.Urgency != "warning" {
			continue
		}
		if item.DaysUntilStockout != nil {
			fmt.Fprintf(&b, "- %s stock=%d velocity=%.2f/day days_left=%.1f reorder=%d\n",
				item.Name, item.Stock, item.DailyVelocity, *item.DaysUntilStockout, item.SuggestedReorderQty)
		} else {
			fmt.Fprintf(&b, "- %s stock=%d velocity=%.2f/day reorder=%d\n",
				item.Name, item.Stock, item.DailyVelocity, item.SuggestedReorderQty)
		}
		shown++
		if shown >= 8 {
			break
		}
	}

	return b.String()
}

func fallbackVendorInventoryForecast(
	snapshot *VendorInventorySnapshot,
	forecasts []dto.VendorInventoryForecastItem,
) *vendorInventoryForecastPayload {
	priorities := make([]string, 0, 4)
	recommendations := make([]string, 0, 3)
	alerts := make([]string, 0, 3)

	for _, item := range forecasts {
		if item.Urgency == "critical" && len(priorities) < 4 {
			if item.DaysUntilStockout != nil {
				priorities = append(priorities, fmt.Sprintf(
					"Restock %s — ~%.0f days of cover left (order %d units)",
					item.Name, *item.DaysUntilStockout, item.SuggestedReorderQty,
				))
			} else {
				priorities = append(priorities, fmt.Sprintf("Restock %s immediately (out of stock)", item.Name))
			}
		}
	}

	if snapshot.LowStockCount > 0 {
		alerts = append(alerts, fmt.Sprintf("%d products are at or below the low-stock threshold", snapshot.LowStockCount))
	}

	slowMovers := 0
	for _, item := range forecasts {
		if item.UnitsSold == 0 && item.Stock > item.SuggestedReorderQty {
			slowMovers++
		}
	}
	if slowMovers > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Review %d slow-moving SKUs before reordering", slowMovers))
	}
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Set safety stock to ~2 weeks of average daily sales for top sellers")
	}

	summary := fmt.Sprintf(
		"%s analyzed %d tracked SKUs over the last %d days.",
		snapshot.StoreName,
		len(forecasts),
		snapshot.PeriodDays,
	)
	if countUrgency(forecasts, "critical") > 0 {
		summary += fmt.Sprintf(" %d products need urgent replenishment.", countUrgency(forecasts, "critical"))
	}

	return &vendorInventoryForecastPayload{
		Summary:         summary,
		Priorities:      priorities,
		Recommendations: recommendations,
		Alerts:          alerts,
	}
}

func parseVendorInventoryForecastJSON(content string) (*vendorInventoryForecastPayload, error) {
	content = extractJSONObject(content)
	var payload vendorInventoryForecastPayload
	if err := json.Unmarshal([]byte(content), &payload); err != nil {
		return nil, err
	}
	if strings.TrimSpace(payload.Summary) == "" {
		return nil, fmt.Errorf("missing summary")
	}
	return &payload, nil
}

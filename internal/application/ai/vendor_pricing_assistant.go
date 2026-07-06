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

const TaskVendorPricingAssistant = "vendor_pricing_assistant"

// VendorPricingSnapshot aggregates catalog pricing and demand for recommendations.
type VendorPricingSnapshot struct {
	StoreName  string
	PeriodDays int
	Products   []VendorProductDemand
}

// VendorPricingAssistantQueries loads pricing demand facts.
type VendorPricingAssistantQueries interface {
	LoadPricingSnapshot(ctx context.Context, storeID uint, days int) (*VendorPricingSnapshot, error)
}

type vendorPricingAssistantPayload struct {
	Summary         string   `json:"summary"`
	Highlights      []string `json:"highlights"`
	Recommendations []string `json:"recommendations"`
	Warnings        []string `json:"warnings"`
}

// VendorPricingAssistant returns AI-generated or rule-based price recommendations for sellers.
func (s *Service) VendorPricingAssistant(
	ctx context.Context,
	storeID uint,
	days int,
	subjectKey string,
	queries VendorPricingAssistantQueries,
) (*dto.AiVendorPricingAssistantResponse, error) {
	if queries == nil {
		return nil, utils.ErrInternal(fmt.Errorf("vendor pricing assistant queries not configured"))
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

	snapshot, err := queries.LoadPricingSnapshot(ctx, storeID, days)
	if err != nil {
		return nil, err
	}

	suggestions := buildPricingSuggestions(snapshot.Products, days)
	response := &dto.AiVendorPricingAssistantResponse{
		PeriodDays:  days,
		Suggestions: suggestions,
		Sources:     []string{"products", "orders", "order_items"},
	}

	if !s.Enabled() {
		fallback := fallbackVendorPricingAssistant(snapshot, suggestions)
		response.AiEnabled = false
		response.Summary = fallback.Summary
		response.Highlights = fallback.Highlights
		response.Recommendations = fallback.Recommendations
		response.Warnings = fallback.Warnings
		return response, nil
	}

	if !s.adminRL.allow("vendor_pricing_assistant:" + subjectKey) {
		return nil, utils.ErrTooManyRequests()
	}

	facts := vendorPricingFacts(snapshot, suggestions)
	system := `You are a concise pricing strategist for marketplace sellers.
Respond with JSON only using this exact shape:
{"summary":"...","highlights":["..."],"recommendations":["..."],"warnings":["..."]}
Rules:
- summary: 2-3 sentences on overall pricing posture and demand signals.
- highlights: 2-4 notable pricing wins or demand patterns from the facts.
- recommendations: 2-3 actionable price moves (raise, discount, hold) tied to specific SKUs.
- warnings: 0-2 margin or clearance risks — empty array if none.
Use ONLY the pricing facts below. Do not invent competitor prices or ad spend.
Plain text only — no markdown.

Pricing facts:
` + facts

	user := fmt.Sprintf("Recommend pricing adjustments based on the last %d days of sales.", days)
	resp, err := s.complete(ctx, system, user, 900)
	if err != nil {
		fallback := fallbackVendorPricingAssistant(snapshot, suggestions)
		response.AiEnabled = false
		response.Summary = "AI is temporarily unavailable. Showing rule-based pricing suggestions from your catalog and sales."
		response.Highlights = fallback.Highlights
		response.Recommendations = fallback.Recommendations
		response.Warnings = fallback.Warnings
		return response, nil
	}

	parsed, err := parseVendorPricingAssistantJSON(resp.Content)
	if err != nil {
		fallback := fallbackVendorPricingAssistant(snapshot, suggestions)
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

func buildPricingSuggestions(products []VendorProductDemand, days int) []dto.VendorPricingSuggestion {
	suggestions := make([]dto.VendorPricingSuggestion, 0, len(products))
	periodDays := float64(days)
	if periodDays <= 0 {
		periodDays = 30
	}

	var maxVelocity float64
	for _, product := range products {
		velocity := float64(product.UnitsSold) / periodDays
		if velocity > maxVelocity {
			maxVelocity = velocity
		}
	}

	for _, product := range products {
		dailyVelocity := float64(product.UnitsSold) / periodDays
		item := dto.VendorPricingSuggestion{
			ProductID:    product.ProductID,
			Name:         product.Name,
			CurrentPrice: product.Price,
			UnitsSold:    product.UnitsSold,
			Revenue:      product.Revenue,
			Action:       "hold",
		}

		if product.Cost != nil && *product.Cost > 0 && product.Price > 0 {
			margin := ((product.Price - *product.Cost) / product.Price) * 100
			rounded := math.Round(margin*10) / 10
			item.MarginPct = &rounded
		}

		highDemand := maxVelocity > 0 && dailyVelocity >= maxVelocity*0.5 && product.UnitsSold >= 3
		lowDemand := product.UnitsSold == 0 && product.Stock > product.LowStockThreshold
		overstocked := product.UnitsSold > 0 && dailyVelocity < 0.1 && product.Stock > product.LowStockThreshold*2

		switch {
		case highDemand && product.Stock <= product.LowStockThreshold:
			item.Action = "increase"
			suggested := roundPrice(product.Price * 1.08)
			item.SuggestedPrice = &suggested
			item.Rationale = "Strong demand with thin inventory — test a modest price increase"
		case lowDemand || overstocked:
			item.Action = "decrease"
			suggested := roundPrice(product.Price * 0.9)
			if product.CompareAtPrice != nil && *product.CompareAtPrice > product.Price {
				suggested = product.Price
			}
			item.SuggestedPrice = &suggested
			item.Rationale = "Slow sales velocity — consider a promotional price to move stock"
		case product.CompareAtPrice != nil && *product.CompareAtPrice > product.Price:
			item.Action = "hold"
			item.Rationale = "Already discounted versus compare-at price — monitor conversion before further cuts"
		default:
			item.Action = "hold"
			item.Rationale = "Demand and inventory are balanced at the current price"
		}

		suggestions = append(suggestions, item)
	}

	return suggestions
}

func roundPrice(value float64) float64 {
	if value < 10 {
		return math.Round(value*100) / 100
	}
	return math.Round(value)
}

func vendorPricingFacts(snapshot *VendorPricingSnapshot, suggestions []dto.VendorPricingSuggestion) string {
	if snapshot == nil {
		return "No pricing data available."
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Store: %s\n", snapshot.StoreName)
	fmt.Fprintf(&b, "Analysis window: %d days\n", snapshot.PeriodDays)
	fmt.Fprintf(&b, "Active products analyzed: %d\n", len(suggestions))

	increase, decrease, hold := 0, 0, 0
	for _, item := range suggestions {
		switch item.Action {
		case "increase":
			increase++
		case "decrease":
			decrease++
		default:
			hold++
		}
	}
	fmt.Fprintf(&b, "Suggested increases: %d | decreases: %d | hold: %d\n", increase, decrease, hold)

	b.WriteString("Notable SKUs:\n")
	shown := 0
	for _, item := range suggestions {
		if item.Action == "hold" {
			continue
		}
		if item.SuggestedPrice != nil {
			fmt.Fprintf(&b, "- %s price=%.2f suggested=%.2f action=%s units=%d revenue=%.2f\n",
				item.Name, item.CurrentPrice, *item.SuggestedPrice, item.Action, item.UnitsSold, item.Revenue)
		} else {
			fmt.Fprintf(&b, "- %s price=%.2f action=%s units=%d revenue=%.2f\n",
				item.Name, item.CurrentPrice, item.Action, item.UnitsSold, item.Revenue)
		}
		shown++
		if shown >= 8 {
			break
		}
	}

	return b.String()
}

func fallbackVendorPricingAssistant(
	snapshot *VendorPricingSnapshot,
	suggestions []dto.VendorPricingSuggestion,
) *vendorPricingAssistantPayload {
	highlights := make([]string, 0, 4)
	recommendations := make([]string, 0, 3)
	warnings := make([]string, 0, 2)

	var totalRevenue float64
	for _, product := range snapshot.Products {
		totalRevenue += product.Revenue
	}
	if totalRevenue > 0 {
		highlights = append(highlights, fmt.Sprintf("$%.2f in attributed revenue over the last %d days", totalRevenue, snapshot.PeriodDays))
	}

	for _, item := range suggestions {
		if item.Action == "increase" && len(recommendations) < 3 {
			if item.SuggestedPrice != nil {
				recommendations = append(recommendations, fmt.Sprintf(
					"Raise %s from $%.2f to $%.2f — %s",
					item.Name, item.CurrentPrice, *item.SuggestedPrice, item.Rationale,
				))
			}
		}
		if item.Action == "decrease" && len(recommendations) < 3 {
			if item.SuggestedPrice != nil {
				recommendations = append(recommendations, fmt.Sprintf(
					"Discount %s from $%.2f to $%.2f — %s",
					item.Name, item.CurrentPrice, *item.SuggestedPrice, item.Rationale,
				))
			}
		}
		if item.MarginPct != nil && *item.MarginPct < 15 {
			warnings = append(warnings, fmt.Sprintf("%s margin is only %.1f%% — review cost or price", item.Name, *item.MarginPct))
		}
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Hold current prices and re-evaluate after the next sales week")
	}

	summary := fmt.Sprintf(
		"%s reviewed %d active products for pricing opportunities over %d days.",
		snapshot.StoreName,
		len(suggestions),
		snapshot.PeriodDays,
	)

	return &vendorPricingAssistantPayload{
		Summary:         summary,
		Highlights:      highlights,
		Recommendations: recommendations,
		Warnings:        warnings,
	}
}

func parseVendorPricingAssistantJSON(content string) (*vendorPricingAssistantPayload, error) {
	content = extractJSONObject(content)
	var payload vendorPricingAssistantPayload
	if err := json.Unmarshal([]byte(content), &payload); err != nil {
		return nil, err
	}
	if strings.TrimSpace(payload.Summary) == "" {
		return nil, fmt.Errorf("missing summary")
	}
	return &payload, nil
}

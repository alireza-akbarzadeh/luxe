package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

const (
	TaskDeliveryPrediction       = "delivery_prediction"
	minDeliveriesForStoreAverage = 5
)

// DeliveryStatsQueries loads shipment timing signals for AI delivery estimates.
type DeliveryStatsQueries interface {
	StoreDeliveryStats(ctx context.Context, storeID uint) (deliveredCount int64, avgDays float64, err error)
	ListActiveProviders(ctx context.Context) ([]models.ShippingProviders, error)
}

// DeliveryPrediction estimates delivery timing for PDP shoppers.
func (s *Service) DeliveryPrediction(
	ctx context.Context,
	subjectKey string,
	shipments DeliveryStatsQueries,
	req dto.AiDeliveryPredictionRequest,
) (*dto.AiDeliveryPredictionResponse, error) {
	if !s.Enabled() {
		return nil, utils.NewAppError(503, "AI is not enabled", nil)
	}
	if !s.chatRL.allow("delivery_prediction:" + subjectKey) {
		return nil, utils.ErrTooManyRequests()
	}
	if shipments == nil {
		return nil, utils.ErrInternal(fmt.Errorf("delivery stats queries not configured"))
	}

	product, err := s.products.GetDetailedByID(ctx, req.ProductID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("product not found")
		}
		return nil, utils.ErrInternal(err)
	}

	facts, sources := productFacts(product)

	if product.IsDigital {
		return &dto.AiDeliveryPredictionResponse{
			Speed:          "digital",
			DeliveryWindow: "Instant access",
			Summary:        "This is a digital product — delivery is instant after purchase with no physical shipping required.",
			Confidence:     "high",
			Highlights: []string{
				"No carrier or transit time",
				"Available immediately after checkout",
			},
			Sources: uniqueStrings(sources),
		}, nil
	}

	deliveryFacts := ""
	if product.StoreID > 0 {
		deliveredCount, avgDays, statsErr := shipments.StoreDeliveryStats(ctx, product.StoreID)
		if statsErr == nil {
			deliveryFacts = storeDeliveryFacts(deliveredCount, avgDays)
			if deliveredCount > 0 {
				sources = append(sources, "Store shipment history")
			}
		}
	}

	providers, providerErr := shipments.ListActiveProviders(ctx)
	if providerErr == nil && len(providers) > 0 {
		deliveryFacts += shippingProviderFacts(providers)
		sources = append(sources, "Shipping options")
	}

	system := `You estimate delivery timing for an online product page to set shopper expectations.
Respond with JSON only using this exact shape:
{"speed":"fast|standard|slow","delivery_window":"...","estimated_days_min":0,"estimated_days_max":0,"summary":"...","confidence":"high|medium|low","highlights":["..."],"factors":["..."]}
Rules:
- speed: fast = typically under 3 business days; standard = about 3-7; slow = over a week or unclear.
- delivery_window: short shopper-friendly phrase like "2-4 business days" or "5-8 business days".
- estimated_days_min/max: integers for business-day range when you can infer from facts; omit both only if truly unknown.
- summary: 2 sentences on when to expect delivery and what affects timing.
- confidence: high only with clear store shipping copy or reliable shipment history; low when facts are thin.
- highlights: 2-3 factual bullets shoppers care about (stock, store location, shipping tiers).
- factors: 2-4 bullets on what could speed up or delay delivery (backorder, distance, processing time).
Use ONLY the facts below. Do not invent carriers, guarantees, or exact dates.
Plain text only — no markdown.

Product facts:
` + facts + deliveryFacts

	user := "Estimate delivery timing and practical factors for this product."
	resp, err := s.complete(ctx, system, user, 900)
	if err != nil {
		return nil, utils.NewAppError(503, "AI provider unavailable", err)
	}

	parsed, err := parseDeliveryPredictionJSON(resp.Content)
	if err != nil {
		return nil, utils.NewAppError(503, "AI returned invalid delivery prediction", err)
	}

	result := &dto.AiDeliveryPredictionResponse{
		Speed:            parsed.Speed,
		DeliveryWindow:   parsed.DeliveryWindow,
		EstimatedDaysMin: parsed.EstimatedDaysMin,
		EstimatedDaysMax: parsed.EstimatedDaysMax,
		Summary:          parsed.Summary,
		Confidence:       parsed.Confidence,
		Highlights:       parsed.Highlights,
		Factors:          parsed.Factors,
		Sources:          uniqueStrings(sources),
	}

	utils.Log.WithFields(map[string]any{
		"product_id": req.ProductID,
		"subject":    subjectKey,
		"task":       TaskDeliveryPrediction,
		"speed":      result.Speed,
	}).Info("ai delivery prediction completed")

	return result, nil
}

func storeDeliveryFacts(deliveredCount int64, avgDays float64) string {
	if deliveredCount == 0 {
		return "\nStore shipment history: no completed deliveries recorded yet for this seller.\n"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "\nStore shipment history: %d delivered shipments", deliveredCount)
	if deliveredCount >= minDeliveriesForStoreAverage && avgDays > 0 {
		fmt.Fprintf(&b, " (average %.1f days from ship to delivery)", avgDays)
	} else if avgDays > 0 {
		fmt.Fprintf(&b, " (average %.1f days — sample still small)", avgDays)
	}
	b.WriteString(".\n")
	return b.String()
}

func shippingProviderFacts(providers []models.ShippingProviders) string {
	var b strings.Builder
	b.WriteString("\nActive shipping options:\n")
	limit := len(providers)
	if limit > 4 {
		limit = 4
	}
	for i := 0; i < limit; i++ {
		provider := providers[i]
		fmt.Fprintf(&b, "- %s", sanitizeAIText(provider.Name))
		if provider.Description != "" {
			fmt.Fprintf(&b, ": %s", sanitizeAIText(provider.Description))
		}
		if provider.Price > 0 {
			fmt.Fprintf(&b, " (from %.2f)", provider.Price)
		}
		b.WriteString("\n")
	}
	return b.String()
}

func parseDeliveryPredictionJSON(content string) (*dto.AiDeliveryPredictionResponse, error) {
	trimmed := strings.TrimSpace(content)
	if idx := strings.Index(trimmed, "{"); idx >= 0 {
		if end := strings.LastIndex(trimmed, "}"); end > idx {
			trimmed = trimmed[idx : end+1]
		}
	}

	var raw struct {
		Speed            string   `json:"speed"`
		DeliveryWindow   string   `json:"delivery_window"`
		EstimatedDaysMin *int     `json:"estimated_days_min"`
		EstimatedDaysMax *int     `json:"estimated_days_max"`
		Summary          string   `json:"summary"`
		Confidence       string   `json:"confidence"`
		Highlights       []string `json:"highlights"`
		Factors          []string `json:"factors"`
	}
	if err := json.Unmarshal([]byte(trimmed), &raw); err != nil {
		return nil, err
	}

	summary := sanitizeAIText(raw.Summary)
	if summary == "" {
		return nil, fmt.Errorf("empty summary")
	}

	window := sanitizeAIText(raw.DeliveryWindow)
	if window == "" {
		window = "Timing varies"
	}

	return &dto.AiDeliveryPredictionResponse{
		Speed:            normalizeDeliverySpeed(raw.Speed),
		DeliveryWindow:   window,
		EstimatedDaysMin: raw.EstimatedDaysMin,
		EstimatedDaysMax: raw.EstimatedDaysMax,
		Summary:          summary,
		Confidence:       normalizeConfidence(raw.Confidence),
		Highlights:       sanitizeStringList(raw.Highlights),
		Factors:          sanitizeStringList(raw.Factors),
	}, nil
}

func normalizeDeliverySpeed(speed string) string {
	switch strings.ToLower(strings.TrimSpace(speed)) {
	case "fast", "standard", "slow", "digital":
		return strings.ToLower(strings.TrimSpace(speed))
	default:
		return "standard"
	}
}

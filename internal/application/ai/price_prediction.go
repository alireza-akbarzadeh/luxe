package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

const (
	TaskPricePrediction        = "price_prediction"
	defaultPricePredictionDays = 90
)

// PriceHistoryQueries loads recorded price snapshots for AI trend analysis.
type PriceHistoryQueries interface {
	GetPriceHistory(ctx context.Context, productID uint, days int) ([]dto.PriceHistoryPoint, error)
}

// PricePrediction forecasts short-term price direction from recorded history.
func (s *Service) PricePrediction(
	ctx context.Context,
	subjectKey string,
	history PriceHistoryQueries,
	req dto.AiPricePredictionRequest,
) (*dto.AiPricePredictionResponse, error) {
	if !s.Enabled() {
		return nil, utils.NewAppError(503, "AI is not enabled", nil)
	}
	if !s.chatRL.allow("price_prediction:" + subjectKey) {
		return nil, utils.ErrTooManyRequests()
	}
	if history == nil {
		return nil, utils.ErrInternal(fmt.Errorf("price history queries not configured"))
	}

	product, err := s.products.GetDetailedByID(ctx, req.ProductID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("product not found")
		}
		return nil, utils.ErrInternal(err)
	}

	days := req.Days
	if days <= 0 {
		days = defaultPricePredictionDays
	}

	points, err := history.GetPriceHistory(ctx, req.ProductID, days)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	facts, sources := productFacts(product)
	historyFacts := priceHistoryFacts(points, product.Price, product.CompareAtPrice)
	if historyFacts != "" {
		facts += historyFacts
		sources = append(sources, "Recorded price history")
	}

	if len(points) < 2 {
		return &dto.AiPricePredictionResponse{
			Trend:          "unknown",
			Direction:      "flat",
			Summary:        "Not enough price history yet to estimate a trend. Check back after more snapshots are recorded.",
			Recommendation: "neutral",
			Confidence:     "low",
			Sources:        uniqueStrings(sources),
		}, nil
	}

	system := `You analyze e-commerce price history for shoppers on a product detail page.
Respond with JSON only using this exact shape:
{"trend":"rising|falling|stable|volatile","direction":"up|down|flat|mixed","summary":"...","predicted_range":"...","recommendation":"buy_now|wait|neutral","confidence":"high|medium|low","highlights":["..."]}
Rules:
- trend: overall pattern over the history window.
- direction: likely near-term move (next ~30 days) — infer conservatively from facts only.
- summary: 2 sentences explaining the trend and what it means for shoppers.
- predicted_range: optional short phrase like "$420–$460 in the next month" when history supports it; omit if too thin.
- recommendation: buy_now when price looks favorable vs recent range; wait when a drop seems likely; neutral otherwise.
- confidence: high only with clear multi-point trend; low when history is sparse or choppy.
- highlights: 2-3 factual bullets (recent low/high, sale dips, compare-at gaps). No invented promotions.
Use ONLY the facts below. Plain text — no markdown.

Facts:
` + facts

	user := "Predict short-term price direction and shopping timing for this listing."
	resp, err := s.complete(ctx, system, user, 800)
	if err != nil {
		return nil, utils.NewAppError(503, "AI provider unavailable", err)
	}

	parsed, err := parsePricePredictionJSON(resp.Content)
	if err != nil {
		return nil, utils.NewAppError(503, "AI returned invalid price prediction", err)
	}

	result := &dto.AiPricePredictionResponse{
		Trend:          parsed.Trend,
		Direction:      parsed.Direction,
		Summary:        parsed.Summary,
		PredictedRange: parsed.PredictedRange,
		Recommendation: parsed.Recommendation,
		Confidence:     parsed.Confidence,
		Highlights:     parsed.Highlights,
		Sources:        uniqueStrings(sources),
	}

	utils.Log.WithFields(map[string]any{
		"product_id": req.ProductID,
		"subject":    subjectKey,
		"task":       TaskPricePrediction,
		"trend":      result.Trend,
	}).Info("ai price prediction completed")

	return result, nil
}

func priceHistoryFacts(points []dto.PriceHistoryPoint, currentPrice float64, compareAt *float64) string {
	var b strings.Builder
	b.WriteString("\nPrice history:\n")
	fmt.Fprintf(&b, "- Current listing price: %.2f\n", currentPrice)
	if compareAt != nil && *compareAt > 0 {
		fmt.Fprintf(&b, "- Compare-at / MSRP: %.2f\n", *compareAt)
	}
	fmt.Fprintf(&b, "- Snapshots in window: %d\n", len(points))
	if len(points) == 0 {
		return b.String()
	}

	first := points[0]
	last := points[len(points)-1]
	if first.Price > 0 && last.Price > 0 {
		delta := last.Price - first.Price
		pct := 0.0
		if first.Price > 0 {
			pct = (delta / first.Price) * 100
		}
		fmt.Fprintf(&b, "- Window change: %.2f (%.1f%%) from %.2f to %.2f\n", delta, pct, first.Price, last.Price)
	}

	high, low := last.Price, last.Price
	for _, point := range points {
		if point.Price > high {
			high = point.Price
		}
		if point.Price < low {
			low = point.Price
		}
	}
	fmt.Fprintf(&b, "- Range in window: low %.2f, high %.2f\n", low, high)

	sample := points
	if len(sample) > 8 {
		sample = sample[len(sample)-8:]
	}
	b.WriteString("- Recent snapshots (oldest→newest in sample):\n")
	for _, point := range sample {
		date := "unknown"
		if !point.RecordedAt.IsZero() {
			date = point.RecordedAt.Format(time.RFC3339)
		}
		fmt.Fprintf(&b, "  • %s price=%.2f", date, point.Price)
		if point.CompareAtPrice != nil && *point.CompareAtPrice > 0 {
			fmt.Fprintf(&b, " compare_at=%.2f", *point.CompareAtPrice)
		}
		b.WriteString("\n")
	}

	return b.String()
}

func parsePricePredictionJSON(content string) (*dto.AiPricePredictionResponse, error) {
	trimmed := strings.TrimSpace(content)
	if idx := strings.Index(trimmed, "{"); idx >= 0 {
		if end := strings.LastIndex(trimmed, "}"); end > idx {
			trimmed = trimmed[idx : end+1]
		}
	}

	var raw struct {
		Trend          string   `json:"trend"`
		Direction      string   `json:"direction"`
		Summary        string   `json:"summary"`
		PredictedRange string   `json:"predicted_range"`
		Recommendation string   `json:"recommendation"`
		Confidence     string   `json:"confidence"`
		Highlights     []string `json:"highlights"`
	}
	if err := json.Unmarshal([]byte(trimmed), &raw); err != nil {
		return nil, err
	}

	summary := sanitizeAIText(raw.Summary)
	if summary == "" {
		return nil, fmt.Errorf("empty summary")
	}

	trend := normalizePriceTrend(raw.Trend)
	direction := normalizePriceDirection(raw.Direction)
	recommendation := normalizePriceRecommendation(raw.Recommendation)
	confidence := normalizeConfidence(raw.Confidence)

	return &dto.AiPricePredictionResponse{
		Trend:          trend,
		Direction:      direction,
		Summary:        summary,
		PredictedRange: sanitizeAIText(raw.PredictedRange),
		Recommendation: recommendation,
		Confidence:     confidence,
		Highlights:     sanitizeStringList(raw.Highlights),
	}, nil
}

func normalizePriceTrend(trend string) string {
	switch strings.ToLower(strings.TrimSpace(trend)) {
	case "rising", "falling", "stable", "volatile", "unknown":
		return strings.ToLower(strings.TrimSpace(trend))
	default:
		return "unknown"
	}
}

func normalizePriceDirection(direction string) string {
	switch strings.ToLower(strings.TrimSpace(direction)) {
	case "up", "down", "flat", "mixed":
		return strings.ToLower(strings.TrimSpace(direction))
	default:
		return "flat"
	}
}

func normalizePriceRecommendation(rec string) string {
	switch strings.ToLower(strings.TrimSpace(rec)) {
	case "buy_now", "wait", "neutral":
		return strings.ToLower(strings.TrimSpace(rec))
	default:
		return "neutral"
	}
}

func normalizeConfidence(confidence string) string {
	switch strings.ToLower(strings.TrimSpace(confidence)) {
	case "high", "medium", "low":
		return strings.ToLower(strings.TrimSpace(confidence))
	default:
		return "medium"
	}
}

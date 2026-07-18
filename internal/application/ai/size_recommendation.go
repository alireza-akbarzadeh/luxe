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
	TaskSizeRecommendation          = "size_recommendation"
	maxReviewsForSizeRecommendation = 20
)

// SizeRecommendation suggests a size from listing variants, reviews, and optional shopper profile.
func (s *Service) SizeRecommendation(
	ctx context.Context,
	subjectKey string,
	returns ReturnRiskQueries,
	reviews ReviewSummaryQueries,
	req dto.AiSizeRecommendationRequest,
) (*dto.AiSizeRecommendationResponse, error) {
	if !s.Enabled() {
		return nil, utils.NewAppError(503, "AI is not enabled", nil)
	}
	if !s.chatRL.allow("size_recommendation:" + subjectKey) {
		return nil, utils.ErrTooManyRequests()
	}

	product, err := s.products.GetDetailedByID(ctx, req.ProductID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("product not found")
		}
		return nil, utils.ErrInternal(err)
	}

	facts, sources := productFacts(product)
	sizes := productSizeValues(product)
	if len(sizes) == 0 {
		return &dto.AiSizeRecommendationResponse{
			FitNotes:   "unknown",
			Confidence: "high",
			Summary:    "This listing does not offer size variants, so a size recommendation is not applicable.",
			Sources:    uniqueStrings(sources),
		}, nil
	}

	facts += sizeListingFacts(sizes)
	sources = append(sources, "Available sizes")

	profileFacts := sizeShopperProfileFacts(req.Profile)
	if profileFacts != "" {
		facts += profileFacts
		sources = append(sources, "Shopper profile")
	}

	if returns != nil {
		orderCount, returnCount, returnReasons, statsErr := returns.ProductReturnStats(ctx, req.ProductID)
		if statsErr == nil {
			returnFacts := returnRiskFacts(orderCount, returnCount, returnReasons)
			if returnFacts != "" {
				facts += returnFacts
				sources = append(sources, "Return history")
			}
		}
	}

	if reviews != nil {
		reviewRows, total, summary, reviewErr := reviews.GetProductReviews(ctx, req.ProductID, maxReviewsForSizeRecommendation, 0)
		if reviewErr == nil && total > 0 {
			reviewFacts := sizeFitReviewFacts(summary, reviewRows)
			if reviewFacts != "" {
				facts += reviewFacts
				sources = append(sources, "Verified buyer reviews")
			}
		}
	}

	system := `You recommend the best size for an online product page.
Respond with JSON only using this exact shape:
{"recommended_size":"...","alternative_size":"...","fit_notes":"runs_small|true_to_size|runs_large|unknown","confidence":"high|medium|low","summary":"...","tips":["..."]}
Rules:
- recommended_size: must be one of the available sizes when possible; empty only if truly unknown.
- alternative_size: optional second-best size from available sizes; omit when not useful.
- fit_notes: how this product fits relative to label (runs small/large vs true to size).
- confidence: high when reviews + profile align; low when guessing from limited data.
- summary: 2 sentences explaining the pick in plain language.
- tips: 2-3 fit or measurement tips; mention returns/size chart when relevant.
Use ONLY the facts below. Do not invent measurements, size charts, or review quotes.
Plain text only — no markdown.

Facts:
` + facts

	user := "Recommend the best size for this shopper."
	resp, err := s.complete(ctx, system, user, 900)
	if err != nil {
		return nil, utils.NewAppError(503, "AI provider unavailable", err)
	}

	parsed, err := parseSizeRecommendationJSON(resp.Content)
	if err != nil {
		return nil, utils.NewAppError(503, "AI returned invalid size recommendation", err)
	}

	parsed.AvailableSizes = sizes
	parsed.Sources = uniqueStrings(sources)
	if parsed.RecommendedSize != "" && !sizeInList(parsed.RecommendedSize, sizes) {
		parsed.RecommendedSize = ""
	}
	if parsed.AlternativeSize != "" && !sizeInList(parsed.AlternativeSize, sizes) {
		parsed.AlternativeSize = ""
	}

	utils.Log.WithFields(map[string]any{
		"product_id": req.ProductID,
		"subject":    subjectKey,
		"task":       TaskSizeRecommendation,
		"size":       parsed.RecommendedSize,
	}).Info("ai size recommendation completed")

	return parsed, nil
}

func productSizeValues(product *models.Product) []string {
	if product == nil {
		return nil
	}

	seen := make(map[string]struct{})
	out := make([]string, 0)

	addValues := func(values []string) {
		for _, value := range values {
			value = strings.TrimSpace(value)
			if value == "" {
				continue
			}
			key := strings.ToLower(value)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, value)
		}
	}

	for _, attr := range product.Attributes {
		name := strings.ToLower(strings.TrimSpace(attr.Name))
		if name != "size" && name != "sizes" {
			continue
		}
		addValues(attr.Values)
	}
	addValues(product.Sizes)

	return out
}

func sizeListingFacts(sizes []string) string {
	return "\nAvailable sizes: " + strings.Join(sizes, ", ") + "\n"
}

func sizeShopperProfileFacts(profile *dto.AiSizeShopperProfile) string {
	if profile == nil {
		return ""
	}

	var b strings.Builder
	b.WriteString("\nShopper profile:\n")
	if profile.UsualSize != "" {
		fmt.Fprintf(&b, "- Usual size: %s\n", sanitizeAIText(profile.UsualSize))
	}
	if profile.HeightCm != nil && *profile.HeightCm > 0 {
		fmt.Fprintf(&b, "- Height: %d cm\n", *profile.HeightCm)
	}
	if profile.WeightKg != nil && *profile.WeightKg > 0 {
		fmt.Fprintf(&b, "- Weight: %d kg\n", *profile.WeightKg)
	}
	if profile.FitPreference != "" {
		fmt.Fprintf(&b, "- Fit preference: %s\n", sanitizeAIText(profile.FitPreference))
	}
	if b.Len() == len("\nShopper profile:\n") {
		return ""
	}
	return b.String()
}

func sizeFitReviewFacts(summary dto.ReviewSummary, reviews []models.Review) string {
	var b strings.Builder
	fmt.Fprintf(&b, "\nReviews: average %.1f from %d reviews\n", summary.Average, summary.Total)
	for i, review := range reviews {
		if i >= maxReviewsForSizeRecommendation {
			break
		}
		text := strings.ToLower(review.Title + " " + review.Comment)
		if !containsSizeFitSignal(text) {
			continue
		}
		fmt.Fprintf(&b, "- %d/5", review.Rating)
		if review.Comment != "" {
			fmt.Fprintf(&b, ": %s", sanitizeAIText(truncate(review.Comment, 140)))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func containsSizeFitSignal(text string) bool {
	keywords := []string{
		"size", "fit", "small", "large", "tight", "loose", "runs", "true to size",
		"snug", "roomy", "narrow", "wide", "length", "sizing",
	}
	for _, kw := range keywords {
		if strings.Contains(text, kw) {
			return true
		}
	}
	return false
}

func sizeInList(size string, sizes []string) bool {
	size = strings.ToLower(strings.TrimSpace(size))
	for _, candidate := range sizes {
		if strings.ToLower(strings.TrimSpace(candidate)) == size {
			return true
		}
	}
	return false
}

func parseSizeRecommendationJSON(content string) (*dto.AiSizeRecommendationResponse, error) {
	trimmed := strings.TrimSpace(content)
	if idx := strings.Index(trimmed, "{"); idx >= 0 {
		if end := strings.LastIndex(trimmed, "}"); end > idx {
			trimmed = trimmed[idx : end+1]
		}
	}

	var raw struct {
		RecommendedSize string   `json:"recommended_size"`
		AlternativeSize string   `json:"alternative_size"`
		FitNotes        string   `json:"fit_notes"`
		Confidence      string   `json:"confidence"`
		Summary         string   `json:"summary"`
		Tips            []string `json:"tips"`
	}
	if err := json.Unmarshal([]byte(trimmed), &raw); err != nil {
		return nil, err
	}

	summary := sanitizeAIText(raw.Summary)
	if summary == "" {
		return nil, fmt.Errorf("empty summary")
	}

	fitNotes := strings.ToLower(strings.TrimSpace(raw.FitNotes))
	switch fitNotes {
	case "runs_small", "true_to_size", "runs_large", "unknown":
	default:
		fitNotes = "unknown"
	}

	confidence := strings.ToLower(strings.TrimSpace(raw.Confidence))
	switch confidence {
	case "high", "medium", "low":
	default:
		confidence = "medium"
	}

	return &dto.AiSizeRecommendationResponse{
		RecommendedSize: sanitizeAIText(raw.RecommendedSize),
		AlternativeSize: sanitizeAIText(raw.AlternativeSize),
		FitNotes:        fitNotes,
		Confidence:      confidence,
		Summary:         summary,
		Tips:            sanitizeStringList(raw.Tips),
	}, nil
}

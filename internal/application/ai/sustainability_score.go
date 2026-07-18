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
	TaskSustainabilityScore          = "sustainability_score"
	maxReviewsForSustainabilityScore = 20
)

// SustainabilityScore estimates environmental and ethical signals for a product listing.
func (s *Service) SustainabilityScore(
	ctx context.Context,
	subjectKey string,
	reviews ReviewSummaryQueries,
	req dto.AiSustainabilityScoreRequest,
) (*dto.AiSustainabilityScoreResponse, error) {
	if !s.Enabled() {
		return nil, utils.NewAppError(503, "AI is not enabled", nil)
	}
	if !s.chatRL.allow("sustainability_score:" + subjectKey) {
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
	sustainFacts := sustainabilityListingFacts(product)
	if sustainFacts != "" {
		facts += sustainFacts
		sources = append(sources, "Product listing & specifications")
	}

	reviewFacts := ""
	if reviews != nil {
		reviewRows, total, summary, reviewErr := reviews.GetProductReviews(ctx, req.ProductID, maxReviewsForSustainabilityScore, 0)
		if reviewErr == nil && total > 0 {
			reviewFacts = sustainabilityReviewFacts(summary, reviewRows)
			sources = append(sources, "Verified buyer reviews")
		}
	}

	system := `You assess sustainability and ethical shopping signals for an online marketplace PDP.
Respond with JSON only using this exact shape:
{"score":0,"rating":"leading|good|mixed|low","summary":"...","pillars":[{"key":"materials|packaging|ethics|longevity","label":"...","score":0,"note":"..."}],"highlights":["..."],"watch_outs":["..."]}
Rules:
- score: integer 0-100 (higher = stronger sustainability signals in the listing).
- rating: leading (75+), good (55-74), mixed (35-54), low (below 35) — align with score.
- summary: 2 sentences on eco/ethics posture based ONLY on provided facts. Be honest when data is thin.
- pillars: exactly 4 items with keys materials, packaging, ethics, longevity.
- highlights: 2-4 positive sustainability signals stated in facts (certifications, materials, repairability, etc.).
- watch_outs: 0-3 gaps or unknowns (greenwashing risk, missing info, single-use packaging). Empty array if none.
Do NOT invent certifications (B Corp, Fair Trade, organic, etc.) unless explicitly in facts.
Plain text only — no markdown.

Facts:
` + facts + reviewFacts

	user := "Score sustainability and ethical shopping signals for this product."
	resp, err := s.complete(ctx, system, user, 1000)
	if err != nil {
		return nil, utils.NewAppError(503, "AI provider unavailable", err)
	}

	parsed, err := parseSustainabilityScoreJSON(resp.Content)
	if err != nil {
		return nil, utils.NewAppError(503, "AI returned invalid sustainability score", err)
	}

	result := &dto.AiSustainabilityScoreResponse{
		Score:      parsed.Score,
		Rating:     parsed.Rating,
		Summary:    parsed.Summary,
		Pillars:    parsed.Pillars,
		Highlights: parsed.Highlights,
		WatchOuts:  parsed.WatchOuts,
		Sources:    uniqueStrings(sources),
	}

	utils.Log.WithFields(map[string]any{
		"product_id": req.ProductID,
		"subject":    subjectKey,
		"task":       TaskSustainabilityScore,
		"score":      result.Score,
	}).Info("ai sustainability score completed")

	return result, nil
}

func sustainabilityListingFacts(product *models.Product) string {
	var b strings.Builder
	b.WriteString("\nSustainability-related listing data:\n")
	if product.IsDigital {
		b.WriteString("- Digital delivery (no physical shipping footprint)\n")
	}
	if len(product.Tags) > 0 {
		fmt.Fprintf(&b, "- Tags: %s\n", sanitizeAIText(strings.Join(product.Tags, ", ")))
	}
	if len(product.Attributes) > 0 {
		b.WriteString("- Attributes:\n")
		for _, attr := range product.Attributes {
			if len(attr.Values) == 0 {
				continue
			}
			fmt.Fprintf(&b, "  • %s: %s\n", sanitizeAIText(attr.Name), sanitizeAIText(strings.Join(attr.Values, ", ")))
		}
	}
	descSnippet := sanitizeAIText(truncate(product.Description, 400))
	if containsSustainabilitySignal(strings.ToLower(descSnippet)) {
		fmt.Fprintf(&b, "- Description excerpt: %s\n", descSnippet)
	}
	if product.Store != nil && product.Store.ShippingInfo != "" {
		fmt.Fprintf(&b, "- Seller shipping notes: %s\n", sanitizeAIText(truncate(product.Store.ShippingInfo, 200)))
	}
	return b.String()
}

func sustainabilityReviewFacts(summary dto.ReviewSummary, reviews []models.Review) string {
	var b strings.Builder
	fmt.Fprintf(&b, "\nProduct reviews: average %.1f from %d reviews\n", summary.Average, summary.Total)
	for i, review := range reviews {
		if i >= maxReviewsForSustainabilityScore {
			break
		}
		text := strings.ToLower(review.Title + " " + review.Comment)
		if !containsSustainabilitySignal(text) {
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

func containsSustainabilitySignal(text string) bool {
	keywords := []string{
		"sustainable", "sustainability", "eco", "organic", "recycled", "recycle",
		"biodegradable", "compostable", "vegan", "cruelty-free", "fair trade",
		"carbon", "plastic-free", "renewable", "ethical", "green", "waste",
		"packaging", "repair", "refill", "local", "handmade",
	}
	for _, kw := range keywords {
		if strings.Contains(text, kw) {
			return true
		}
	}
	return false
}

func parseSustainabilityScoreJSON(content string) (*dto.AiSustainabilityScoreResponse, error) {
	trimmed := strings.TrimSpace(content)
	if idx := strings.Index(trimmed, "{"); idx >= 0 {
		if end := strings.LastIndex(trimmed, "}"); end > idx {
			trimmed = trimmed[idx : end+1]
		}
	}

	var raw struct {
		Score   int    `json:"score"`
		Rating  string `json:"rating"`
		Summary string `json:"summary"`
		Pillars []struct {
			Key   string `json:"key"`
			Label string `json:"label"`
			Score int    `json:"score"`
			Note  string `json:"note"`
		} `json:"pillars"`
		Highlights []string `json:"highlights"`
		WatchOuts  []string `json:"watch_outs"`
	}
	if err := json.Unmarshal([]byte(trimmed), &raw); err != nil {
		return nil, err
	}

	summary := sanitizeAIText(raw.Summary)
	if summary == "" {
		return nil, fmt.Errorf("empty summary")
	}

	rating := strings.ToLower(strings.TrimSpace(raw.Rating))
	switch rating {
	case "leading", "good", "mixed", "low":
	default:
		rating = sustainabilityRatingFromScore(clampTrustScore(raw.Score))
	}

	pillars := make([]dto.AiSustainabilityPillar, 0, len(raw.Pillars))
	for _, pillar := range raw.Pillars {
		label := sanitizeAIText(pillar.Label)
		note := sanitizeAIText(pillar.Note)
		if label == "" && note == "" {
			continue
		}
		if label == "" {
			label = strings.TrimSpace(pillar.Key)
		}
		pillars = append(pillars, dto.AiSustainabilityPillar{
			Key:   strings.TrimSpace(pillar.Key),
			Label: label,
			Score: clampTrustScore(pillar.Score),
			Note:  note,
		})
	}

	return &dto.AiSustainabilityScoreResponse{
		Score:      clampTrustScore(raw.Score),
		Rating:     rating,
		Summary:    summary,
		Pillars:    pillars,
		Highlights: sanitizeStringList(raw.Highlights),
		WatchOuts:  sanitizeStringList(raw.WatchOuts),
	}, nil
}

func sustainabilityRatingFromScore(score int) string {
	switch {
	case score >= 75:
		return "leading"
	case score >= 55:
		return "good"
	case score >= 35:
		return "mixed"
	default:
		return "low"
	}
}

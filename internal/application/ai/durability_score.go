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
	TaskDurabilityScore          = "durability_score"
	maxReviewsForDurabilityScore = 20
)

// DurabilityScore estimates how well a product should hold up over time.
func (s *Service) DurabilityScore(
	ctx context.Context,
	subjectKey string,
	reviews ReviewSummaryQueries,
	req dto.AiDurabilityScoreRequest,
) (*dto.AiDurabilityScoreResponse, error) {
	if !s.Enabled() {
		return nil, utils.NewAppError(503, "AI is not enabled", nil)
	}
	if !s.chatRL.allow("durability_score:" + subjectKey) {
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
	attrFacts := durabilityAttributeFacts(product)
	if attrFacts != "" {
		facts += attrFacts
		sources = append(sources, "Product specifications")
	}

	reviewFacts := ""
	if reviews != nil {
		reviewRows, total, summary, reviewErr := reviews.GetProductReviews(ctx, req.ProductID, maxReviewsForDurabilityScore, 0)
		if reviewErr == nil && total > 0 {
			reviewFacts = durabilityReviewFacts(summary, reviewRows)
			sources = append(sources, "Verified buyer reviews")
		}
	}

	system := `You assess product durability and expected longevity for an online PDP.
Respond with JSON only using this exact shape:
{"score":0,"tier":"excellent|good|fair|limited","summary":"...","lifespan_estimate":"...","highlights":[{"label":"...","note":"..."}],"care_tips":["..."]}
Rules:
- score: integer 0-100 (higher = more durable / longer-lasting).
- tier: excellent (75+), good (55-74), fair (35-54), limited (below 35) — pick tier consistent with score.
- summary: 2 sentences on build quality and longevity expectations.
- lifespan_estimate: short phrase like "3-5 years with normal use" or "Seasonal wear; 1-2 years" when inferable; omit if unknown.
- highlights: 2-4 bullets on materials, construction, or brand/category signals from facts.
- care_tips: 2-3 practical maintenance tips when relevant; empty array if not applicable.
Use ONLY the facts below. Do not invent warranties, materials, or lab tests.
Plain text only — no markdown.

Facts:
` + facts + reviewFacts

	user := "Score durability and longevity for this product listing."
	resp, err := s.complete(ctx, system, user, 900)
	if err != nil {
		return nil, utils.NewAppError(503, "AI provider unavailable", err)
	}

	parsed, err := parseDurabilityScoreJSON(resp.Content)
	if err != nil {
		return nil, utils.NewAppError(503, "AI returned invalid durability score", err)
	}

	result := &dto.AiDurabilityScoreResponse{
		Score:            parsed.Score,
		Tier:             parsed.Tier,
		Summary:          parsed.Summary,
		LifespanEstimate: parsed.LifespanEstimate,
		Highlights:       parsed.Highlights,
		CareTips:         parsed.CareTips,
		Sources:          uniqueStrings(sources),
	}

	utils.Log.WithFields(map[string]any{
		"product_id": req.ProductID,
		"subject":    subjectKey,
		"task":       TaskDurabilityScore,
		"score":      result.Score,
	}).Info("ai durability score completed")

	return result, nil
}

func durabilityAttributeFacts(product *models.Product) string {
	var b strings.Builder
	b.WriteString("\nSpecifications:\n")
	if product.Weight != nil && *product.Weight > 0 {
		fmt.Fprintf(&b, "- Weight: %.2f\n", *product.Weight)
	}
	if product.IsDigital {
		b.WriteString("- Digital product (no physical wear)\n")
	}
	if len(product.Attributes) == 0 {
		b.WriteString("- No structured attributes listed\n")
		return b.String()
	}
	for _, attr := range product.Attributes {
		if len(attr.Values) == 0 {
			continue
		}
		fmt.Fprintf(&b, "- %s: %s\n", sanitizeAIText(attr.Name), sanitizeAIText(strings.Join(attr.Values, ", ")))
	}
	return b.String()
}

func durabilityReviewFacts(summary dto.ReviewSummary, reviews []models.Review) string {
	var b strings.Builder
	fmt.Fprintf(&b, "\nProduct reviews: average %.1f from %d reviews\n", summary.Average, summary.Total)
	for i, review := range reviews {
		if i >= maxReviewsForDurabilityScore {
			break
		}
		text := strings.ToLower(review.Title + " " + review.Comment)
		if !containsDurabilitySignal(text) {
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

func containsDurabilitySignal(text string) bool {
	keywords := []string{
		"durable", "durability", "sturdy", "solid", "quality", "build",
		"wear", "broke", "broken", "crack", "last", "lasting", "fragile",
		"premium", "cheap", "flimsy", "holds up", "long-lasting",
	}
	for _, kw := range keywords {
		if strings.Contains(text, kw) {
			return true
		}
	}
	return false
}

func parseDurabilityScoreJSON(content string) (*dto.AiDurabilityScoreResponse, error) {
	trimmed := strings.TrimSpace(content)
	if idx := strings.Index(trimmed, "{"); idx >= 0 {
		if end := strings.LastIndex(trimmed, "}"); end > idx {
			trimmed = trimmed[idx : end+1]
		}
	}

	var raw struct {
		Score            int    `json:"score"`
		Tier             string `json:"tier"`
		Summary          string `json:"summary"`
		LifespanEstimate string `json:"lifespan_estimate"`
		Highlights       []struct {
			Label string `json:"label"`
			Note  string `json:"note"`
		} `json:"highlights"`
		CareTips []string `json:"care_tips"`
	}
	if err := json.Unmarshal([]byte(trimmed), &raw); err != nil {
		return nil, err
	}

	summary := sanitizeAIText(raw.Summary)
	if summary == "" {
		return nil, fmt.Errorf("empty summary")
	}

	tier := strings.ToLower(strings.TrimSpace(raw.Tier))
	switch tier {
	case "excellent", "good", "fair", "limited":
	default:
		tier = tierFromScore(clampTrustScore(raw.Score))
	}

	highlights := make([]dto.AiDurabilityHighlight, 0, len(raw.Highlights))
	for _, item := range raw.Highlights {
		label := sanitizeAIText(item.Label)
		note := sanitizeAIText(item.Note)
		if label == "" && note == "" {
			continue
		}
		highlights = append(highlights, dto.AiDurabilityHighlight{
			Label: label,
			Note:  note,
		})
	}

	return &dto.AiDurabilityScoreResponse{
		Score:            clampTrustScore(raw.Score),
		Tier:             tier,
		Summary:          summary,
		LifespanEstimate: sanitizeAIText(raw.LifespanEstimate),
		Highlights:       highlights,
		CareTips:         sanitizeStringList(raw.CareTips),
	}, nil
}

func tierFromScore(score int) string {
	switch {
	case score >= 75:
		return "excellent"
	case score >= 55:
		return "good"
	case score >= 35:
		return "fair"
	default:
		return "limited"
	}
}

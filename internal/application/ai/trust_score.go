package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

const (
	TaskTrustScore           = "trust_score"
	maxReviewsForTrustScore  = 15
)

// TrustScore synthesizes buyer trust signals into a 0–100 score with factor notes.
func (s *Service) TrustScore(
	ctx context.Context,
	subjectKey string,
	returns ReturnRiskQueries,
	reviews ReviewSummaryQueries,
	req dto.AiTrustScoreRequest,
) (*dto.AiTrustScoreResponse, error) {
	if !s.Enabled() {
		return nil, utils.NewAppError(503, "AI is not enabled", nil)
	}
	if !s.chatRL.allow("trust_score:" + subjectKey) {
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
	storeFacts := trustScoreStoreFacts(product.Store)
	if storeFacts != "" {
		facts += storeFacts
		sources = append(sources, "Seller profile")
	}

	returnFacts := ""
	if returns != nil {
		orderCount, returnCount, returnReasons, statsErr := returns.ProductReturnStats(ctx, req.ProductID)
		if statsErr == nil {
			returnFacts = returnRiskFacts(orderCount, returnCount, returnReasons)
			if returnFacts != "" {
				sources = append(sources, "Order & return history")
			}
		}
	}

	reviewFacts := ""
	var reviewTotal int64
	var reviewAverage float64
	if reviews != nil {
		reviewRows, total, summary, reviewErr := reviews.GetProductReviews(ctx, req.ProductID, maxReviewsForTrustScore, 0)
		if reviewErr == nil && total > 0 {
			reviewTotal = total
			reviewAverage = summary.Average
			reviewFacts = trustScoreReviewFacts(summary, reviewRows)
			sources = append(sources, "Verified buyer reviews")
		}
	}

	system := `You compute a product trust score for an online marketplace PDP.
Respond with JSON only using this exact shape:
{"score":0,"confidence":"low|medium|high","summary":"...","factors":[{"key":"reviews|seller|listing|fulfillment","label":"...","score":0,"note":"..."}]}
Rules:
- score: integer 0-100 overall trust (higher = more trustworthy to buy).
- confidence: high when reviews + seller + order data exist; medium with partial data; low when mostly listing-only.
- summary: 2 sentences explaining the score in plain language for shoppers.
- factors: exactly 4 items with keys reviews, seller, listing, fulfillment. Each factor score is 0-100 with a one-line note.
- Weight reviews and fulfillment heavily when data exists; do not invent ratings, return rates, or verification.
Use ONLY the facts below. Plain text only — no markdown.

Facts:
` + facts + returnFacts + reviewFacts

	user := "Score how trustworthy this product listing is for a new buyer."
	resp, err := s.complete(ctx, system, user, 1000)
	if err != nil {
		return nil, utils.NewAppError(503, "AI provider unavailable", err)
	}

	parsed, err := parseTrustScoreJSON(resp.Content)
	if err != nil {
		return nil, utils.NewAppError(503, "AI returned invalid trust score", err)
	}

	result := &dto.AiTrustScoreResponse{
		Score:         parsed.Score,
		Confidence:    parsed.Confidence,
		Summary:       parsed.Summary,
		Factors:       parsed.Factors,
		Sources:       uniqueStrings(sources),
		ReviewCount:   reviewTotal,
		AverageRating: reviewAverage,
	}

	utils.Log.WithFields(map[string]any{
		"product_id": req.ProductID,
		"subject":    subjectKey,
		"task":       TaskTrustScore,
		"score":      result.Score,
	}).Info("ai trust score completed")

	return result, nil
}

func trustScoreStoreFacts(store *models.Store) string {
	if store == nil {
		return "\nSeller: unknown (no store profile linked).\n"
	}

	var b strings.Builder
	b.WriteString("\nSeller profile:\n")
	fmt.Fprintf(&b, "- Name: %s\n", sanitizeAIText(store.Name))
	fmt.Fprintf(&b, "- Verified seller: %t\n", store.IsVerified)
	if store.Rating > 0 {
		fmt.Fprintf(&b, "- Store rating: %.1f from %d store reviews\n", store.Rating, store.ReviewCount)
	}
	if store.ReturnPolicy != "" {
		fmt.Fprintf(&b, "- Return policy: %s\n", sanitizeAIText(truncate(store.ReturnPolicy, 300)))
	}
	if store.ShippingInfo != "" {
		fmt.Fprintf(&b, "- Shipping: %s\n", sanitizeAIText(truncate(store.ShippingInfo, 200)))
	}
	return b.String()
}

func trustScoreReviewFacts(summary dto.ReviewSummary, reviews []models.Review) string {
	var b strings.Builder
	fmt.Fprintf(&b, "\nProduct reviews: average %.1f from %d reviews\n", summary.Average, summary.Total)
	for i, review := range reviews {
		if i >= maxReviewsForTrustScore {
			break
		}
		verified := ""
		if review.IsVerified {
			verified = " [verified]"
		}
		fmt.Fprintf(&b, "- %d/5%s", review.Rating, verified)
		if review.Comment != "" {
			fmt.Fprintf(&b, ": %s", sanitizeAIText(truncate(review.Comment, 120)))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func parseTrustScoreJSON(content string) (*dto.AiTrustScoreResponse, error) {
	trimmed := strings.TrimSpace(content)
	if idx := strings.Index(trimmed, "{"); idx >= 0 {
		if end := strings.LastIndex(trimmed, "}"); end > idx {
			trimmed = trimmed[idx : end+1]
		}
	}

	var raw struct {
		Score      int    `json:"score"`
		Confidence string `json:"confidence"`
		Summary    string `json:"summary"`
		Factors    []struct {
			Key   string `json:"key"`
			Label string `json:"label"`
			Score int    `json:"score"`
			Note  string `json:"note"`
		} `json:"factors"`
	}
	if err := json.Unmarshal([]byte(trimmed), &raw); err != nil {
		return nil, err
	}

	summary := sanitizeAIText(raw.Summary)
	if summary == "" {
		return nil, fmt.Errorf("empty summary")
	}

	confidence := strings.ToLower(strings.TrimSpace(raw.Confidence))
	switch confidence {
	case "low", "medium", "high":
	default:
		confidence = "medium"
	}

	factors := make([]dto.AiTrustScoreFactor, 0, len(raw.Factors))
	for _, factor := range raw.Factors {
		label := sanitizeAIText(factor.Label)
		note := sanitizeAIText(factor.Note)
		if label == "" && note == "" {
			continue
		}
		if label == "" {
			label = strings.TrimSpace(factor.Key)
		}
		factors = append(factors, dto.AiTrustScoreFactor{
			Key:   strings.TrimSpace(factor.Key),
			Label: label,
			Score: clampTrustScore(factor.Score),
			Note:  note,
		})
	}

	return &dto.AiTrustScoreResponse{
		Score:      clampTrustScore(raw.Score),
		Confidence: confidence,
		Summary:    summary,
		Factors:    factors,
	}, nil
}

func clampTrustScore(score int) int {
	return int(math.Max(0, math.Min(100, float64(score))))
}

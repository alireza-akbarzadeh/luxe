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
	TaskReviewSummary       = "review_summary"
	minReviewsForAISummary    = 3
	maxReviewsForAISummary    = 30
)

// ReviewSummaryQueries loads approved product reviews for AI synthesis.
type ReviewSummaryQueries interface {
	GetProductReviews(ctx context.Context, productID uint, limit, offset int) ([]models.Review, int64, dto.ReviewSummary, error)
}

// ReviewSummary synthesizes buyer reviews into themes shoppers can scan quickly.
func (s *Service) ReviewSummary(
	ctx context.Context,
	subjectKey string,
	reviews ReviewSummaryQueries,
	req dto.AiReviewSummaryRequest,
) (*dto.AiReviewSummaryResponse, error) {
	if !s.Enabled() {
		return nil, utils.NewAppError(503, "AI is not enabled", nil)
	}
	if !s.chatRL.allow("review_summary:" + subjectKey) {
		return nil, utils.ErrTooManyRequests()
	}
	if reviews == nil {
		return nil, utils.ErrInternal(fmt.Errorf("review queries not configured"))
	}

	product, err := s.products.GetDetailedByID(ctx, req.ProductID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("product not found")
		}
		return nil, utils.ErrInternal(err)
	}

	reviewRows, total, summary, err := reviews.GetProductReviews(ctx, req.ProductID, maxReviewsForAISummary, 0)
	if err != nil {
		return nil, err
	}
	if total < minReviewsForAISummary {
		return nil, utils.ErrBadRequest("not enough reviews for AI summary")
	}

	facts := reviewFacts(product.Name, summary, reviewRows)
	system := `You summarize verified customer reviews for an online product page.
Respond with JSON only using this exact shape:
{"summary":"...","highlights":["..."],"watch_outs":["..."]}
Rules:
- summary: 2-3 sentences capturing overall buyer sentiment. Mention review count only if provided.
- highlights: 2-4 bullets on what buyers praise most often.
- watch_outs: 0-3 bullets on recurring concerns or caveats (omit array items if none; empty array is fine).
Use ONLY the review excerpts below. Do not invent features, specs, or policies.
Plain text only — no markdown, no numbering.

Reviews:
` + facts

	user := "Summarize what customers are saying about this product."
	resp, err := s.complete(ctx, system, user, 800)
	if err != nil {
		return nil, utils.NewAppError(503, "AI provider unavailable", err)
	}

	parsed, err := parseReviewSummaryJSON(resp.Content)
	if err != nil {
		return nil, utils.NewAppError(503, "AI returned invalid review summary", err)
	}

	result := &dto.AiReviewSummaryResponse{
		Summary:       parsed.Summary,
		Highlights:    parsed.Highlights,
		WatchOuts:     parsed.WatchOuts,
		ReviewCount:   total,
		AverageRating: summary.Average,
		Sources:       []string{"Verified buyer reviews"},
	}

	utils.Log.WithFields(map[string]any{
		"product_id": req.ProductID,
		"subject":    subjectKey,
		"task":       TaskReviewSummary,
		"reviews":    total,
	}).Info("ai review summary completed")

	return result, nil
}

func reviewFacts(productName string, summary dto.ReviewSummary, reviews []models.Review) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Product: %s\nAverage rating: %.1f from %d reviews\n\n", productName, summary.Average, summary.Total)
	for i, review := range reviews {
		if i >= maxReviewsForAISummary {
			break
		}
		verified := ""
		if review.IsVerified {
			verified = " [verified purchase]"
		}
		fmt.Fprintf(&b, "- %d/5%s", review.Rating, verified)
		if review.Title != "" {
			fmt.Fprintf(&b, " — %s", sanitizeAIText(review.Title))
		}
		if review.Comment != "" {
			fmt.Fprintf(&b, ": %s", sanitizeAIText(review.Comment))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func parseReviewSummaryJSON(content string) (*dto.AiReviewSummaryResponse, error) {
	trimmed := strings.TrimSpace(content)
	if idx := strings.Index(trimmed, "{"); idx >= 0 {
		if end := strings.LastIndex(trimmed, "}"); end > idx {
			trimmed = trimmed[idx : end+1]
		}
	}

	var raw struct {
		Summary    string   `json:"summary"`
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

	return &dto.AiReviewSummaryResponse{
		Summary:    summary,
		Highlights: sanitizeStringList(raw.Highlights),
		WatchOuts:  sanitizeStringList(raw.WatchOuts),
	}, nil
}

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
	TaskReturnRisk          = "return_risk"
	minOrdersForReturnRate  = 5
	maxReviewsForReturnRisk = 20
)

// ReturnRiskQueries loads return statistics for AI return-risk insights.
type ReturnRiskQueries interface {
	ProductReturnStats(ctx context.Context, productID uint) (orderCount int64, returnCount int64, reasons []string, err error)
}

// ReturnRisk explains return likelihood and practical tips for PDP shoppers.
func (s *Service) ReturnRisk(
	ctx context.Context,
	subjectKey string,
	returns ReturnRiskQueries,
	reviews ReviewSummaryQueries,
	req dto.AiReturnRiskRequest,
) (*dto.AiReturnRiskResponse, error) {
	if !s.Enabled() {
		return nil, utils.NewAppError(503, "AI is not enabled", nil)
	}
	if !s.chatRL.allow("return_risk:" + subjectKey) {
		return nil, utils.ErrTooManyRequests()
	}
	if returns == nil {
		return nil, utils.ErrInternal(fmt.Errorf("return queries not configured"))
	}

	product, err := s.products.GetDetailedByID(ctx, req.ProductID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("product not found")
		}
		return nil, utils.ErrInternal(err)
	}

	orderCount, returnCount, returnReasons, err := returns.ProductReturnStats(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}

	facts, sources := productFacts(product)
	returnFacts := returnRiskFacts(orderCount, returnCount, returnReasons)
	reviewFacts := ""
	if reviews != nil {
		reviewRows, _, summary, reviewErr := reviews.GetProductReviews(ctx, req.ProductID, maxReviewsForReturnRisk, 0)
		if reviewErr == nil && summary.Total > 0 {
			reviewFacts = returnRiskReviewFacts(summary, reviewRows)
			sources = append(sources, "Verified buyer reviews")
		}
	}
	if returnFacts != "" {
		sources = append(sources, "Return history")
	}

	system := `You assess return risk for an online product page to help shoppers buy with confidence.
Respond with JSON only using this exact shape:
{"risk_level":"low|medium|high","summary":"...","common_reasons":["..."],"tips":["..."]}
Rules:
- risk_level: low = straightforward purchase; medium = some fit/spec caveats; high = frequent mismatch or quality concerns.
- summary: 2-3 sentences. Be honest but not alarmist. Mention return policy when relevant.
- common_reasons: 0-4 bullets on why similar buyers return this type of item (use return stats and reviews when provided).
- tips: 2-4 actionable bullets to reduce return risk (size guides, read specs, check photos, etc.).
Use ONLY the facts below. Do not invent return rates, policies, or review quotes.
Plain text only — no markdown, no numbering.

Product facts:
` + facts + returnFacts + reviewFacts

	user := "Assess return risk and give practical guidance for this product."
	resp, err := s.complete(ctx, system, user, 900)
	if err != nil {
		return nil, utils.NewAppError(503, "AI provider unavailable", err)
	}

	parsed, err := parseReturnRiskJSON(resp.Content)
	if err != nil {
		return nil, utils.NewAppError(503, "AI returned invalid return risk insight", err)
	}

	result := &dto.AiReturnRiskResponse{
		RiskLevel:     parsed.RiskLevel,
		Summary:       parsed.Summary,
		CommonReasons: parsed.CommonReasons,
		Tips:          parsed.Tips,
		Sources:       uniqueStrings(sources),
	}

	if orderCount >= minOrdersForReturnRate {
		rate := float64(returnCount) / float64(orderCount) * 100
		rounded := float64(int(rate*10+0.5)) / 10
		result.ReturnRatePct = &rounded
		result.OrderSampleSize = orderCount
	}

	utils.Log.WithFields(map[string]any{
		"product_id": req.ProductID,
		"subject":    subjectKey,
		"task":       TaskReturnRisk,
		"orders":     orderCount,
		"returns":    returnCount,
	}).Info("ai return risk completed")

	return result, nil
}

func returnRiskFacts(orderCount, returnCount int64, reasons []string) string {
	if orderCount == 0 && returnCount == 0 && len(reasons) == 0 {
		return "\nReturn history: no completed orders with return data for this product yet.\n"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "\nReturn history: %d returns linked to orders containing this product", returnCount)
	if orderCount >= minOrdersForReturnRate {
		rate := float64(returnCount) / float64(orderCount) * 100
		fmt.Fprintf(&b, " across %d orders (approx. %.1f%% return rate)", orderCount, rate)
	} else if orderCount > 0 {
		fmt.Fprintf(&b, " across %d orders (sample too small for a reliable rate)", orderCount)
	}
	b.WriteString(".\n")
	if len(reasons) > 0 {
		b.WriteString("Reported return reasons:\n")
		for _, reason := range reasons {
			fmt.Fprintf(&b, "- %s\n", sanitizeAIText(reason))
		}
	}
	return b.String()
}

func returnRiskReviewFacts(summary dto.ReviewSummary, reviews []models.Review) string {
	var b strings.Builder
	fmt.Fprintf(&b, "\nReviews: average %.1f from %d reviews\n", summary.Average, summary.Total)
	for i, review := range reviews {
		if i >= maxReviewsForReturnRisk {
			break
		}
		if review.Rating > 3 {
			continue
		}
		fmt.Fprintf(&b, "- %d/5", review.Rating)
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

func parseReturnRiskJSON(content string) (*dto.AiReturnRiskResponse, error) {
	trimmed := strings.TrimSpace(content)
	if idx := strings.Index(trimmed, "{"); idx >= 0 {
		if end := strings.LastIndex(trimmed, "}"); end > idx {
			trimmed = trimmed[idx : end+1]
		}
	}

	var raw struct {
		RiskLevel     string   `json:"risk_level"`
		Summary       string   `json:"summary"`
		CommonReasons []string `json:"common_reasons"`
		Tips          []string `json:"tips"`
	}
	if err := json.Unmarshal([]byte(trimmed), &raw); err != nil {
		return nil, err
	}

	summary := sanitizeAIText(raw.Summary)
	if summary == "" {
		return nil, fmt.Errorf("empty summary")
	}

	level := strings.ToLower(strings.TrimSpace(raw.RiskLevel))
	switch level {
	case "low", "medium", "high":
	default:
		level = "medium"
	}

	return &dto.AiReturnRiskResponse{
		RiskLevel:     level,
		Summary:       summary,
		CommonReasons: sanitizeStringList(raw.CommonReasons),
		Tips:          sanitizeStringList(raw.Tips),
	}, nil
}

func uniqueStrings(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

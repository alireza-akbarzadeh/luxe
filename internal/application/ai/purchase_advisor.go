package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

const (
	TaskPurchaseAdvisor          = "purchase_advisor"
	maxReviewsForPurchaseAdvisor = 15
)

// PurchaseAdvisor synthesizes product, review, return, and price signals into buy guidance.
func (s *Service) PurchaseAdvisor(
	ctx context.Context,
	subjectKey string,
	returns ReturnRiskQueries,
	reviews ReviewSummaryQueries,
	history PriceHistoryQueries,
	req dto.AiPurchaseAdvisorRequest,
) (*dto.AiPurchaseAdvisorResponse, error) {
	if !s.Enabled() {
		return nil, utils.NewAppError(503, "AI is not enabled", nil)
	}
	if !s.chatRL.allow("purchase_advisor:" + subjectKey) {
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

	if returns != nil {
		orderCount, returnCount, returnReasons, statsErr := returns.ProductReturnStats(ctx, req.ProductID)
		if statsErr == nil {
			returnFacts := returnRiskFacts(orderCount, returnCount, returnReasons)
			if returnFacts != "" {
				facts += returnFacts
				sources = append(sources, "Order & return history")
			}
		}
	}

	if reviews != nil {
		reviewRows, total, summary, reviewErr := reviews.GetProductReviews(ctx, req.ProductID, maxReviewsForPurchaseAdvisor, 0)
		if reviewErr == nil && total > 0 {
			facts += trustScoreReviewFacts(summary, reviewRows)
			sources = append(sources, "Verified buyer reviews")
		}
	}

	if history != nil {
		points, histErr := history.GetPriceHistory(ctx, req.ProductID, defaultPricePredictionDays)
		if histErr == nil {
			historyFacts := priceHistoryFacts(points, product.Price, product.CompareAtPrice)
			if historyFacts != "" {
				facts += historyFacts
				sources = append(sources, "Recorded price history")
			}
		}
	}

	system := `You are a purchase advisor on an online marketplace product page.
Synthesize the facts below into a clear buy / wait / consider recommendation.
Respond with JSON only using this exact shape:
{"verdict":"buy|wait|consider","confidence":"high|medium|low","summary":"...","pros":["..."],"cons":["..."],"ideal_for":["..."],"considerations":["..."]}
Rules:
- verdict: buy = good time to purchase for most shoppers; wait = price, stock, or timing suggests patience; consider = mixed fit — only buy if needs match.
- confidence: high when reviews + price/return data support the call; low when mostly listing-only.
- summary: 2-3 sentences with the headline recommendation and why.
- pros: 2-4 bullets on strengths grounded in facts.
- cons: 1-3 honest caveats (fit, price, returns, stock) — omit if none.
- ideal_for: 2-3 shopper profiles who would be happy with this purchase.
- considerations: 2-4 checks before checkout (size, policy, timing, alternatives).
Use ONLY the facts below. Do not invent ratings, prices, policies, or guarantees.
Plain text only — no markdown, no numbering.

Facts:
` + facts

	user := "Should a shopper buy this product now? Give a decisive recommendation."
	resp, err := s.complete(ctx, system, user, 1100)
	if err != nil {
		return nil, utils.NewAppError(503, "AI provider unavailable", err)
	}

	parsed, err := parsePurchaseAdvisorJSON(resp.Content)
	if err != nil {
		return nil, utils.NewAppError(503, "AI returned invalid purchase advisor", err)
	}

	result := &dto.AiPurchaseAdvisorResponse{
		Verdict:        parsed.Verdict,
		Confidence:     parsed.Confidence,
		Summary:        parsed.Summary,
		Pros:           parsed.Pros,
		Cons:           parsed.Cons,
		IdealFor:       parsed.IdealFor,
		Considerations: parsed.Considerations,
		Sources:        uniqueStrings(sources),
	}

	utils.Log.WithFields(map[string]any{
		"product_id": req.ProductID,
		"subject":    subjectKey,
		"task":       TaskPurchaseAdvisor,
		"verdict":    result.Verdict,
	}).Info("ai purchase advisor completed")

	return result, nil
}

func parsePurchaseAdvisorJSON(content string) (*dto.AiPurchaseAdvisorResponse, error) {
	trimmed := strings.TrimSpace(content)
	if idx := strings.Index(trimmed, "{"); idx >= 0 {
		if end := strings.LastIndex(trimmed, "}"); end > idx {
			trimmed = trimmed[idx : end+1]
		}
	}

	var raw struct {
		Verdict        string   `json:"verdict"`
		Confidence     string   `json:"confidence"`
		Summary        string   `json:"summary"`
		Pros           []string `json:"pros"`
		Cons           []string `json:"cons"`
		IdealFor       []string `json:"ideal_for"`
		Considerations []string `json:"considerations"`
	}
	if err := json.Unmarshal([]byte(trimmed), &raw); err != nil {
		return nil, err
	}

	summary := sanitizeAIText(raw.Summary)
	if summary == "" {
		return nil, fmt.Errorf("empty summary")
	}

	verdict := strings.ToLower(strings.TrimSpace(raw.Verdict))
	switch verdict {
	case "buy", "wait", "consider":
	default:
		verdict = "consider"
	}

	confidence := strings.ToLower(strings.TrimSpace(raw.Confidence))
	switch confidence {
	case "high", "medium", "low":
	default:
		confidence = "medium"
	}

	return &dto.AiPurchaseAdvisorResponse{
		Verdict:        verdict,
		Confidence:     confidence,
		Summary:        summary,
		Pros:           sanitizeStringList(raw.Pros),
		Cons:           sanitizeStringList(raw.Cons),
		IdealFor:       sanitizeStringList(raw.IdealFor),
		Considerations: sanitizeStringList(raw.Considerations),
	}, nil
}

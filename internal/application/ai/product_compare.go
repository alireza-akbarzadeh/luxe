package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

const TaskProductCompare = "product_compare"

// CompareQueries loads product details for side-by-side comparison.
type CompareQueries interface {
	GetForCompare(ctx context.Context, productIDs []uint) ([]*dto.CompareProductResponse, error)
}

// CompareInsight explains trade-offs between 2–4 products using grounded catalog facts.
func (s *Service) CompareInsight(
	ctx context.Context,
	subjectKey string,
	compare CompareQueries,
	req dto.AiCompareInsightRequest,
) (*dto.AiCompareInsightResponse, error) {
	if !s.Enabled() {
		return nil, utils.NewAppError(503, "AI is not enabled", nil)
	}
	if !s.chatRL.allow("compare:" + subjectKey) {
		return nil, utils.ErrTooManyRequests()
	}
	if compare == nil {
		return nil, utils.ErrInternal(fmt.Errorf("compare queries not configured"))
	}

	products, err := compare.GetForCompare(ctx, req.ProductIDs)
	if err != nil {
		return nil, err
	}

	facts := compareProductFacts(products)
	system := `You help shoppers compare products side by side on a luxury e-commerce store.
Respond with JSON only using this exact shape:
{"summary":"...","recommendation":"...","best_for":[{"label":"...","product_name":"...","reason":"..."}],"tradeoffs":["..."]}
Rules:
- summary: 2-3 sentences comparing the products at a high level.
- recommendation: one clear sentence on which product to pick for most shoppers and why.
- best_for: 2-4 entries mapping a shopper goal (e.g. "Best price", "Highest rated") to a product name from the facts and a short reason.
- tradeoffs: 2-4 concise bullets about meaningful differences shoppers should weigh.
Use ONLY the product facts below. Do not invent specs, prices, ratings, or policies.
Product names in best_for must exactly match names from the facts.

Product facts:
` + facts

	user := fmt.Sprintf("Compare these %d products and help me decide which to buy.", len(products))
	resp, err := s.complete(ctx, system, user, 1100)
	if err != nil {
		return nil, utils.NewAppError(503, "AI provider unavailable", err)
	}

	insight, err := parseCompareInsightJSON(resp.Content)
	if err != nil {
		return nil, utils.NewAppError(503, "AI returned invalid comparison", err)
	}

	utils.Log.WithFields(map[string]any{
		"product_ids": req.ProductIDs,
		"subject":     subjectKey,
		"task":        TaskProductCompare,
	}).Info("ai compare insight completed")

	return insight, nil
}

func compareProductFacts(products []*dto.CompareProductResponse) string {
	var b strings.Builder
	for i, product := range products {
		if product == nil {
			continue
		}
		fmt.Fprintf(&b, "\nProduct %d:\n", i+1)
		fmt.Fprintf(&b, "- Name: %s\n", strings.TrimSpace(product.Name))
		if product.Brand != nil && product.Brand.Name != "" {
			fmt.Fprintf(&b, "- Brand: %s\n", product.Brand.Name)
		}
		if product.Category != nil && product.Category.Name != "" {
			fmt.Fprintf(&b, "- Category: %s\n", product.Category.Name)
		}
		fmt.Fprintf(&b, "- Price: %.2f\n", product.Price)
		if product.CompareAtPrice != nil {
			fmt.Fprintf(&b, "- Original price: %.2f\n", *product.CompareAtPrice)
		}
		if product.DiscountPercent > 0 {
			fmt.Fprintf(&b, "- Discount: %.0f%%\n", product.DiscountPercent)
		}
		fmt.Fprintf(&b, "- Rating: %.1f (%d reviews)\n", product.Rating, product.ReviewsCount)
		fmt.Fprintf(&b, "- In stock: %d units\n", product.Stock)
		if product.StoreName != "" {
			fmt.Fprintf(&b, "- Seller: %s\n", product.StoreName)
		}
		if product.ShippingInfo != "" {
			fmt.Fprintf(&b, "- Shipping: %s\n", truncate(product.ShippingInfo, 200))
		}
		if product.ReturnPolicy != "" {
			fmt.Fprintf(&b, "- Returns: %s\n", truncate(product.ReturnPolicy, 200))
		}
		if len(product.Tags) > 0 {
			fmt.Fprintf(&b, "- Tags: %s\n", strings.Join(product.Tags, ", "))
		}
		desc := truncate(product.Description, 400)
		if desc != "" {
			fmt.Fprintf(&b, "- Description: %s\n", desc)
		}
	}
	return b.String()
}

func parseCompareInsightJSON(content string) (*dto.AiCompareInsightResponse, error) {
	trimmed := strings.TrimSpace(content)
	if idx := strings.Index(trimmed, "{"); idx >= 0 {
		if end := strings.LastIndex(trimmed, "}"); end > idx {
			trimmed = trimmed[idx : end+1]
		}
	}

	var raw struct {
		Summary        string `json:"summary"`
		Recommendation string `json:"recommendation"`
		BestFor        []struct {
			Label       string `json:"label"`
			ProductName string `json:"product_name"`
			Reason      string `json:"reason"`
		} `json:"best_for"`
		Tradeoffs []string `json:"tradeoffs"`
	}
	if err := json.Unmarshal([]byte(trimmed), &raw); err != nil {
		return nil, err
	}

	insight := &dto.AiCompareInsightResponse{
		Summary:        sanitizeAIText(raw.Summary),
		Recommendation: sanitizeAIText(raw.Recommendation),
		Tradeoffs:      sanitizeStringList(raw.Tradeoffs),
	}
	for _, item := range raw.BestFor {
		label := sanitizeAIText(item.Label)
		name := sanitizeAIText(item.ProductName)
		reason := sanitizeAIText(item.Reason)
		if label == "" || name == "" || reason == "" {
			continue
		}
		insight.BestFor = append(insight.BestFor, dto.AiCompareBestFor{
			Label:       label,
			ProductName: name,
			Reason:      reason,
		})
	}

	if insight.Summary == "" && insight.Recommendation == "" {
		return nil, fmt.Errorf("empty comparison insight")
	}
	return insight, nil
}

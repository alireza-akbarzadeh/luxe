package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

const (
	TaskSmartCart         = "smart_cart"
	maxSmartCartProducts  = 20
	defaultSmartCartLimit = 12
)

type smartCartPayload struct {
	Summary     string   `json:"summary"`
	Tips        []string `json:"tips"`
	Warnings    []string `json:"warnings"`
	Gaps        []string `json:"gaps"`
	SearchQuery string   `json:"search_query"`
}

// SmartCart reviews cart contents and suggests checkout tips plus complementary picks.
func (s *Service) SmartCart(
	ctx context.Context,
	userID uint,
	subjectKey string,
	search SearchQueries,
	req dto.AiSmartCartRequest,
) (*dto.AiSmartCartResponse, error) {
	if !s.Enabled() {
		return nil, utils.NewAppError(503, "AI is not enabled", nil)
	}
	if userID == 0 {
		return nil, utils.ErrUnauthorized("authentication required")
	}
	if search == nil {
		return nil, utils.ErrInternal(fmt.Errorf("search not configured"))
	}
	if !s.chatRL.allow("smart_cart:" + subjectKey) {
		return nil, utils.ErrTooManyRequests()
	}

	productIDs := uniqueUints(req.ProductIDs)
	if len(productIDs) > maxSmartCartProducts {
		productIDs = productIDs[:maxSmartCartProducts]
	}

	products := make([]*models.Product, 0, len(productIDs))
	for _, id := range productIDs {
		product, err := s.products.GetDetailedByID(ctx, id)
		if err != nil || product == nil {
			continue
		}
		products = append(products, product)
	}
	if len(products) == 0 {
		return nil, utils.ErrBadRequest("no valid products in cart")
	}

	facts, sources := smartCartFacts(products, req.Subtotal, req.Context)
	system := `You review a shopper's cart before checkout.
Respond with JSON only using this exact shape:
{"summary":"...","tips":["..."],"warnings":["..."],"gaps":["..."],"search_query":"..."}
Rules:
- summary: 2-3 sentences on whether the cart looks cohesive and ready to checkout.
- tips: 2-4 practical checkout tips (sizing, bundles, shipping, timing).
- warnings: 0-3 issues (duplicate categories, missing essentials, inactive listings) — empty array if none.
- gaps: 0-3 complementary items the shopper might still need.
- search_query: short catalog keywords for ONE complementary product to add; empty if cart is complete.
Use ONLY the cart facts below. Do not invent prices, stock, or promotions.
Plain text only — no markdown.

Cart facts:
` + facts

	user := "Review this cart and help the shopper checkout with confidence."
	resp, err := s.complete(ctx, system, user, 1100)
	if err != nil {
		return nil, utils.NewAppError(503, "AI provider unavailable", err)
	}

	parsed, err := parseSmartCartJSON(resp.Content)
	if err != nil {
		return nil, utils.NewAppError(503, "AI returned invalid smart cart insight", err)
	}

	response := &dto.AiSmartCartResponse{
		Summary:  parsed.Summary,
		Tips:     sanitizeStringList(parsed.Tips),
		Warnings: sanitizeStringList(parsed.Warnings),
		Gaps:     sanitizeStringList(parsed.Gaps),
		Sources:  uniqueStrings(sources),
	}

	query := strings.TrimSpace(parsed.SearchQuery)
	if query == "" {
		return response, nil
	}

	inStock := true
	searchResult, err := search.GlobalSearch(ctx, dto.SearchRequest{
		Query:   query,
		Limit:   defaultSmartCartLimit,
		Offset:  0,
		InStock: &inStock,
	})
	if err != nil {
		return response, nil
	}

	cartIDSet := make(map[uint]struct{}, len(productIDs))
	for _, id := range productIDs {
		cartIDSet[id] = struct{}{}
	}

	response.Recommendations = make([]dto.AiRecommendedProduct, 0, len(searchResult.Products))
	for _, product := range searchResult.Products {
		if product.ID == 0 {
			continue
		}
		if _, inCart := cartIDSet[product.ID]; inCart {
			continue
		}
		response.Recommendations = append(response.Recommendations, dto.AiRecommendedProduct{
			Product: product,
			Reason:  "Complements your current cart.",
		})
		if len(response.Recommendations) >= 4 {
			break
		}
	}
	if len(response.Recommendations) > 0 {
		response.Sources = append(response.Sources, "Catalog search")
	}

	utils.Log.WithFields(map[string]any{
		"user_id": userID,
		"task":    TaskSmartCart,
		"items":   len(products),
	}).Info("ai smart cart completed")

	return response, nil
}

func smartCartFacts(products []*models.Product, subtotal float64, contextNote string) (string, []string) {
	var b strings.Builder
	sources := []string{"Cart items"}

	fmt.Fprintf(&b, "%d cart items:\n", len(products))
	var computedSubtotal float64
	for i, product := range products {
		if i >= maxSmartCartProducts {
			break
		}
		stock := "in stock"
		if product.Stock <= 0 {
			stock = "out of stock"
		}
		computedSubtotal += product.Price
		fmt.Fprintf(&b, "- %s | price %.2f | %s | category %s\n",
			sanitizeAIText(product.Name),
			product.Price,
			stock,
			productCategoryLabel(product),
		)
	}

	if subtotal > 0 {
		fmt.Fprintf(&b, "\nReported subtotal: %.2f\n", subtotal)
	} else if computedSubtotal > 0 {
		fmt.Fprintf(&b, "\nEstimated subtotal: %.2f\n", computedSubtotal)
	}

	if note := strings.TrimSpace(contextNote); note != "" {
		fmt.Fprintf(&b, "\nShopper note: %s\n", sanitizeAIText(note))
		sources = append(sources, "Shopper context")
	}

	return b.String(), sources
}

func parseSmartCartJSON(raw string) (*smartCartPayload, error) {
	raw = extractJSONObject(raw)
	var parsed smartCartPayload
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, err
	}
	if strings.TrimSpace(parsed.Summary) == "" {
		return nil, fmt.Errorf("missing summary")
	}
	return &parsed, nil
}

func uniqueUints(ids []uint) []uint {
	seen := make(map[uint]struct{}, len(ids))
	out := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

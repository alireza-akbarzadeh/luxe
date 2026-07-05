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
	TaskWishlistIntelligence   = "wishlist_intelligence"
	defaultWishlistIntelLimit  = 30
	maxWishlistIntelLimit      = 50
)

// WishlistIntelligenceQueries loads saved products for AI wishlist analysis.
type WishlistIntelligenceQueries interface {
	GetUserWishlist(userID uint, limit, offset int, sortBy string) ([]models.Product, int64, error)
}

// WishlistIntelligence prioritizes saved items using price, stock, and listing signals.
func (s *Service) WishlistIntelligence(
	ctx context.Context,
	userID uint,
	subjectKey string,
	wishlist WishlistIntelligenceQueries,
	req dto.AiWishlistIntelligenceRequest,
) (*dto.AiWishlistIntelligenceResponse, error) {
	if !s.Enabled() {
		return nil, utils.NewAppError(503, "AI is not enabled", nil)
	}
	if userID == 0 {
		return nil, utils.ErrUnauthorized("authentication required")
	}
	if wishlist == nil {
		return nil, utils.ErrInternal(fmt.Errorf("wishlist queries not configured"))
	}
	if !s.chatRL.allow("wishlist_intelligence:" + subjectKey) {
		return nil, utils.ErrTooManyRequests()
	}

	limit := req.Limit
	if limit <= 0 {
		limit = defaultWishlistIntelLimit
	}
	if limit > maxWishlistIntelLimit {
		limit = maxWishlistIntelLimit
	}

	products, total, err := wishlist.GetUserWishlist(userID, limit, 0, "price-desc")
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	if total == 0 || len(products) == 0 {
		return &dto.AiWishlistIntelligenceResponse{
			Summary:    "Your wishlist is empty. Save items you are considering to get personalized buy, watch, and wait guidance.",
			TotalItems: 0,
			Sources:    []string{"Wishlist"},
		}, nil
	}

	facts, estimatedSavings := wishlistFacts(products)
	sources := []string{"Saved wishlist items"}

	system := `You analyze a shopper's wishlist and prioritize what to do next.
Respond with JSON only using this exact shape:
{"summary":"...","highlights":["..."],"items":[{"product_id":0,"product_name":"...","priority":"buy_now|watch|wait|remove","reason":"..."}]}
Rules:
- summary: 2-3 sentences on overall wishlist health (deals, stock risks, stale items).
- highlights: 2-4 cross-list bullets (total savings, low stock, inactive listings).
- items: one entry per listed product with a short reason; priority buy_now for strong deals/in-stock must-haves, watch for price drops, wait for non-urgent items, remove for inactive/out-of-stock with no restock signal.
Use ONLY the wishlist facts below. Do not invent prices, stock, or promotions.
Plain text only — no markdown.

Wishlist facts:
` + facts

	user := "Prioritize this wishlist and tell the shopper what to buy, watch, or wait on."
	resp, err := s.complete(ctx, system, user, 1200)
	if err != nil {
		return nil, utils.NewAppError(503, "AI provider unavailable", err)
	}

	parsed, err := parseWishlistIntelligenceJSON(resp.Content, products)
	if err != nil {
		return nil, utils.NewAppError(503, "AI returned invalid wishlist intelligence", err)
	}

	parsed.TotalItems = int(total)
	if estimatedSavings > 0 {
		parsed.EstimatedSavings = estimatedSavings
	}
	parsed.Sources = uniqueStrings(sources)

	utils.Log.WithFields(map[string]any{
		"user_id": userID,
		"subject": subjectKey,
		"task":    TaskWishlistIntelligence,
		"items":   len(products),
	}).Info("ai wishlist intelligence completed")

	return parsed, nil
}

func wishlistFacts(products []models.Product) (string, float64) {
	var b strings.Builder
	var savings float64

	fmt.Fprintf(&b, "%d saved items:\n", len(products))
	for i, product := range products {
		if i >= maxWishlistIntelLimit {
			break
		}
		imageNote := ""
		if len(product.Images) == 0 {
			imageNote = " (no image)"
		}
		_ = imageNote

		fmt.Fprintf(&b, "- ID %d: %s | price %.2f", product.ID, sanitizeAIText(product.Name), product.Price)
		if product.CompareAtPrice != nil && *product.CompareAtPrice > product.Price {
			discount := *product.CompareAtPrice - product.Price
			savings += discount
			fmt.Fprintf(&b, " | was %.2f (%.0f%% off)", *product.CompareAtPrice, discount/ *product.CompareAtPrice*100)
		}
		fmt.Fprintf(&b, " | stock %d | status %s", product.Stock, sanitizeAIText(product.Status))
		if product.Category != nil && product.Category.Name != "" {
			fmt.Fprintf(&b, " | category %s", sanitizeAIText(product.Category.Name))
		}
		b.WriteString("\n")
	}

	return b.String(), savings
}

func parseWishlistIntelligenceJSON(content string, products []models.Product) (*dto.AiWishlistIntelligenceResponse, error) {
	trimmed := strings.TrimSpace(content)
	if idx := strings.Index(trimmed, "{"); idx >= 0 {
		if end := strings.LastIndex(trimmed, "}"); end > idx {
			trimmed = trimmed[idx : end+1]
		}
	}

	var raw struct {
		Summary    string `json:"summary"`
		Highlights []string `json:"highlights"`
		Items      []struct {
			ProductID   uint   `json:"product_id"`
			ProductName string `json:"product_name"`
			Priority    string `json:"priority"`
			Reason      string `json:"reason"`
		} `json:"items"`
	}
	if err := json.Unmarshal([]byte(trimmed), &raw); err != nil {
		return nil, err
	}

	summary := sanitizeAIText(raw.Summary)
	if summary == "" {
		return nil, fmt.Errorf("empty summary")
	}

	nameByID := make(map[uint]string, len(products))
	for _, product := range products {
		nameByID[product.ID] = product.Name
	}

	items := make([]dto.AiWishlistInsightItem, 0, len(raw.Items))
	for _, item := range raw.Items {
		if item.ProductID == 0 {
			continue
		}
		priority := strings.ToLower(strings.TrimSpace(item.Priority))
		switch priority {
		case "buy_now", "watch", "wait", "remove":
		default:
			priority = "watch"
		}
		name := sanitizeAIText(item.ProductName)
		if name == "" {
			name = nameByID[item.ProductID]
		}
		reason := sanitizeAIText(item.Reason)
		if reason == "" {
			continue
		}
		items = append(items, dto.AiWishlistInsightItem{
			ProductID:   item.ProductID,
			ProductName: name,
			Priority:    priority,
			Reason:      reason,
		})
	}

	return &dto.AiWishlistIntelligenceResponse{
		Summary:    summary,
		Highlights: sanitizeStringList(raw.Highlights),
		Items:      items,
	}, nil
}

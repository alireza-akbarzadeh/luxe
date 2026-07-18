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
	TaskPersonalizedNotifications = "personalized_notifications"
	defaultNotificationLimit      = 6
	maxNotificationLimit          = 8
)

type personalizedNotificationsPayload struct {
	Summary     string                         `json:"summary"`
	Suggestions []dto.AiNotificationSuggestion `json:"suggestions"`
}

// PersonalizedNotifications recommends alert types based on shopper activity.
func (s *Service) PersonalizedNotifications(
	ctx context.Context,
	userID uint,
	subjectKey string,
	memory ShoppingMemoryQueries,
	req dto.AiPersonalizedNotificationsRequest,
) (*dto.AiPersonalizedNotificationsResponse, error) {
	if !s.Enabled() {
		return nil, utils.NewAppError(503, "AI is not enabled", nil)
	}
	if userID == 0 {
		return nil, utils.ErrUnauthorized("authentication required")
	}
	if memory == nil {
		return nil, utils.ErrInternal(fmt.Errorf("shopping memory queries not configured"))
	}
	if !s.chatRL.allow("personalized_notifications:" + subjectKey) {
		return nil, utils.ErrTooManyRequests()
	}

	limit := req.Limit
	if limit <= 0 {
		limit = defaultNotificationLimit
	}
	if limit > maxNotificationLimit {
		limit = maxNotificationLimit
	}

	viewedIDs, _ := memory.ListRecentlyViewedProductIDs(ctx, userID, 15)
	wishlist, wishlistTotal, _ := memory.GetUserWishlist(userID, 15, 0, "price-desc")
	categoryIDs, _ := memory.ListFavoriteCategoryIDs(ctx, userID)

	if len(viewedIDs) == 0 && wishlistTotal == 0 && len(categoryIDs) == 0 {
		return &dto.AiPersonalizedNotificationsResponse{
			Summary: "Start browsing and saving items so Luxe can recommend the right order, price, and restock alerts for you.",
			Suggestions: []dto.AiNotificationSuggestion{
				{
					Type:        "order_updates",
					Title:       "Order updates",
					Description: "Shipping and delivery alerts for purchases you place.",
					Priority:    "high",
					Suggested:   true,
				},
			},
			Sources: []string{"Account defaults"},
		}, nil
	}

	facts, sources := personalizedNotificationFacts(ctx, s, viewedIDs, wishlist, categoryIDs, wishlistTotal)
	system := `You recommend notification preferences for a luxury e-commerce shopper.
Respond with JSON only using this exact shape:
{"summary":"...","suggestions":[{"type":"order_updates|price_drops|back_in_stock|style_picks|promotions|wishlist_alerts","title":"...","description":"...","priority":"high|medium|low","suggested":true}]}
Rules:
- summary: 2 sentences on why these alerts fit this shopper.
- suggestions: 3-6 items ranked by usefulness; type must be one of the allowed values.
- suggested: true when we recommend enabling that alert now.
Use ONLY the shopper facts below. Do not invent purchases or subscriptions.
Plain text only — no markdown.

Shopper facts:
` + facts

	user := "Recommend personalized notification types for this shopper."
	resp, err := s.complete(ctx, system, user, 1000)
	if err != nil {
		return nil, utils.NewAppError(503, "AI provider unavailable", err)
	}

	parsed, err := parsePersonalizedNotificationsJSON(resp.Content, limit)
	if err != nil {
		return nil, utils.NewAppError(503, "AI returned invalid notification suggestions", err)
	}

	return &dto.AiPersonalizedNotificationsResponse{
		Summary:     parsed.Summary,
		Suggestions: parsed.Suggestions,
		Sources:     uniqueStrings(sources),
	}, nil
}

func personalizedNotificationFacts(
	ctx context.Context,
	s *Service,
	viewedIDs []uint,
	wishlist []models.Product,
	categoryIDs []uint,
	wishlistTotal int64,
) (string, []string) {
	var b strings.Builder
	sources := []string{}

	if len(viewedIDs) > 0 {
		sources = append(sources, "Recently viewed")
		fmt.Fprintf(&b, "Recently viewed count: %d\n", len(viewedIDs))
	}
	if wishlistTotal > 0 {
		sources = append(sources, "Wishlist")
		fmt.Fprintf(&b, "Wishlist items: %d\n", wishlistTotal)
		for i, product := range wishlist {
			if i >= 5 {
				break
			}
			stock := "in stock"
			if product.Stock <= 0 {
				stock = "out of stock"
			}
			fmt.Fprintf(&b, "- saved %s | %.2f | %s\n", sanitizeAIText(product.Name), product.Price, stock)
		}
	}
	if len(categoryIDs) > 0 {
		sources = append(sources, "Favorite categories")
		fmt.Fprintf(&b, "Favorite category count: %d\n", len(categoryIDs))
	}

	if len(sources) == 0 {
		sources = append(sources, "Shopping activity")
	}

	return b.String(), sources
}

func parsePersonalizedNotificationsJSON(raw string, limit int) (*personalizedNotificationsPayload, error) {
	raw = extractJSONObject(raw)
	var parsed personalizedNotificationsPayload
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, err
	}
	if strings.TrimSpace(parsed.Summary) == "" {
		return nil, fmt.Errorf("missing summary")
	}

	allowedTypes := map[string]struct{}{
		"order_updates":   {},
		"price_drops":     {},
		"back_in_stock":   {},
		"style_picks":     {},
		"promotions":      {},
		"wishlist_alerts": {},
	}
	allowedPriority := map[string]struct{}{
		"high":   {},
		"medium": {},
		"low":    {},
	}

	clean := make([]dto.AiNotificationSuggestion, 0, len(parsed.Suggestions))
	for _, item := range parsed.Suggestions {
		itemType := strings.TrimSpace(strings.ToLower(item.Type))
		if _, ok := allowedTypes[itemType]; !ok {
			continue
		}
		priority := strings.TrimSpace(strings.ToLower(item.Priority))
		if _, ok := allowedPriority[priority]; !ok {
			priority = "medium"
		}
		title := strings.TrimSpace(item.Title)
		description := strings.TrimSpace(item.Description)
		if title == "" || description == "" {
			continue
		}
		clean = append(clean, dto.AiNotificationSuggestion{
			Type:        itemType,
			Title:       title,
			Description: description,
			Priority:    priority,
			Suggested:   item.Suggested,
		})
		if len(clean) >= limit {
			break
		}
	}

	if len(clean) == 0 {
		return nil, fmt.Errorf("no valid suggestions")
	}

	parsed.Suggestions = clean
	return &parsed, nil
}

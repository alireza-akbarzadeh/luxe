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
	TaskShoppingMemory         = "shopping_memory"
	defaultShoppingMemoryLimit = 20
	maxShoppingMemoryLimit     = 40
)

// ShoppingMemoryQueries loads personalized shopper signals for AI memory.
type ShoppingMemoryQueries interface {
	ListRecentlyViewedProductIDs(ctx context.Context, userID uint, limit int) ([]uint, error)
	ListFavoriteCategoryIDs(ctx context.Context, userID uint) ([]uint, error)
	GetUserWishlist(userID uint, limit, offset int, sortBy string) ([]models.Product, int64, error)
}

type shoppingMemoryPayload struct {
	Summary     string                       `json:"summary"`
	StyleNotes  []string                     `json:"style_notes"`
	Signals     []dto.AiShoppingMemorySignal `json:"signals"`
	SearchQuery string                       `json:"search_query"`
}

// ShoppingMemory summarizes taste from recently viewed items, wishlist, and favorite categories.
func (s *Service) ShoppingMemory(
	ctx context.Context,
	userID uint,
	subjectKey string,
	memory ShoppingMemoryQueries,
	search SearchQueries,
	req dto.AiShoppingMemoryRequest,
) (*dto.AiShoppingMemoryResponse, error) {
	if !s.Enabled() {
		return nil, utils.NewAppError(503, "AI is not enabled", nil)
	}
	if userID == 0 {
		return nil, utils.ErrUnauthorized("authentication required")
	}
	if memory == nil {
		return nil, utils.ErrInternal(fmt.Errorf("shopping memory queries not configured"))
	}
	if search == nil {
		return nil, utils.ErrInternal(fmt.Errorf("search not configured"))
	}
	if !s.chatRL.allow("shopping_memory:" + subjectKey) {
		return nil, utils.ErrTooManyRequests()
	}

	limit := req.Limit
	if limit <= 0 {
		limit = defaultShoppingMemoryLimit
	}
	if limit > maxShoppingMemoryLimit {
		limit = maxShoppingMemoryLimit
	}

	viewedIDs, _ := memory.ListRecentlyViewedProductIDs(ctx, userID, limit)
	wishlist, _, _ := memory.GetUserWishlist(userID, limit, 0, "price-desc")
	categoryIDs, _ := memory.ListFavoriteCategoryIDs(ctx, userID)

	if len(viewedIDs) == 0 && len(wishlist) == 0 && len(categoryIDs) == 0 {
		return &dto.AiShoppingMemoryResponse{
			Summary: "We do not have enough browsing history yet. View products, save favorites, or pick home categories to build your shopping memory.",
			Sources: []string{"Shopping activity"},
		}, nil
	}

	facts, sources := shoppingMemoryFacts(ctx, s, viewedIDs, wishlist, categoryIDs)
	if facts == "" {
		return &dto.AiShoppingMemoryResponse{
			Summary: "Keep browsing and saving items — your shopping memory will appear here.",
			Sources: uniqueStrings(sources),
		}, nil
	}

	system := `You summarize a shopper's taste from their recent activity on a luxury marketplace.
Respond with JSON only using this exact shape:
{"summary":"...","style_notes":["..."],"signals":[{"label":"...","detail":"..."}],"search_query":"..."}
Rules:
- summary: 2-3 sentences describing their style, price band, and categories in plain language.
- style_notes: 3-5 short bullets (colors, materials, brands, price sensitivity).
- signals: 3-6 labeled observations from the facts (e.g. "Recently viewed", "Wishlist focus").
- search_query: short catalog keywords for 4-6 products that fit their memory; never empty when facts exist.
Use ONLY the activity facts below. Do not invent purchases or sizes.
Plain text only — no markdown.

Activity facts:
` + facts

	user := "Summarize this shopper's memory and suggest a search query for matching products."
	resp, err := s.complete(ctx, system, user, 900)
	if err != nil {
		return nil, utils.NewAppError(503, "AI provider unavailable", err)
	}

	parsed, err := parseShoppingMemoryJSON(resp.Content)
	if err != nil {
		return nil, utils.NewAppError(503, "AI returned invalid shopping memory", err)
	}

	response := &dto.AiShoppingMemoryResponse{
		Summary:    parsed.Summary,
		StyleNotes: sanitizeStringList(parsed.StyleNotes),
		Signals:    parsed.Signals,
		Sources:    uniqueStrings(sources),
	}

	query := strings.TrimSpace(parsed.SearchQuery)
	if query == "" {
		return response, nil
	}

	inStock := true
	searchResult, err := search.GlobalSearch(ctx, dto.SearchRequest{
		Query:   query,
		Limit:   6,
		Offset:  0,
		InStock: &inStock,
	})
	if err != nil {
		return nil, err
	}

	response.Recommendations = make([]dto.AiRecommendedProduct, 0, len(searchResult.Products))
	for _, product := range searchResult.Products {
		response.Recommendations = append(response.Recommendations, dto.AiRecommendedProduct{
			Product: product,
			Reason:  "Matches your recent browsing and saved style.",
		})
	}
	response.Sources = append(response.Sources, "Catalog search")

	utils.Log.WithFields(map[string]any{
		"user_id": userID,
		"task":    TaskShoppingMemory,
	}).Info("ai shopping memory completed")

	return response, nil
}

func shoppingMemoryFacts(
	ctx context.Context,
	s *Service,
	viewedIDs []uint,
	wishlist []models.Product,
	categoryIDs []uint,
) (string, []string) {
	var b strings.Builder
	sources := []string{}

	if len(viewedIDs) > 0 {
		sources = append(sources, "Recently viewed")
		b.WriteString("\nRecently viewed products:\n")
		for _, id := range viewedIDs {
			product, err := s.products.GetDetailedByID(ctx, id)
			if err != nil || product == nil {
				continue
			}
			fmt.Fprintf(&b, "- %s | price %.2f | category %s\n",
				sanitizeAIText(product.Name),
				product.Price,
				productCategoryLabel(product),
			)
		}
	}

	if len(wishlist) > 0 {
		sources = append(sources, "Wishlist")
		b.WriteString("\nSaved wishlist items:\n")
		for _, product := range wishlist {
			fmt.Fprintf(&b, "- %s | price %.2f | stock %d\n",
				sanitizeAIText(product.Name),
				product.Price,
				product.Stock,
			)
		}
	}

	if len(categoryIDs) > 0 {
		sources = append(sources, "Favorite categories")
		b.WriteString("\nFavorite category ids: ")
		for i, id := range categoryIDs {
			if i > 0 {
				b.WriteString(", ")
			}
			fmt.Fprintf(&b, "%d", id)
		}
		b.WriteString("\n")
	}

	return b.String(), sources
}

func parseShoppingMemoryJSON(raw string) (*shoppingMemoryPayload, error) {
	raw = extractJSONObject(raw)
	var parsed shoppingMemoryPayload
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, err
	}
	parsed.Summary = strings.TrimSpace(parsed.Summary)
	if parsed.Summary == "" {
		return nil, fmt.Errorf("missing summary")
	}
	return &parsed, nil
}

func productCategoryLabel(product *models.Product) string {
	if product == nil || product.Category == nil {
		return "unknown"
	}
	return sanitizeAIText(product.Category.Name)
}

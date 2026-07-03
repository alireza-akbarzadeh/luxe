package bundle

import (
	"context"
	"fmt"
	"math"
	"strings"

	appai "github.com/alireza-akbarzadeh/luxe/internal/application/ai"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

// Service generates smart product bundles from catalog compatibility heuristics.
type Service struct {
	repo *postgres.BundleRepository
	ai   *appai.Service
}

// NewService wires smart bundle use cases.
func NewService(db *gorm.DB, ai *appai.Service) *Service {
	return &Service{
		repo: postgres.NewBundleRepository(db),
		ai:   ai,
	}
}

var intentKeywords = map[string][]string{
	"everyday":  {"essential", "classic", "everyday"},
	"workspace": {"desk", "office", "work", "organizer", "workspace"},
	"travel":    {"travel", "leather", "carry", "portable", "trip"},
	"gift":      {"gift", "premium", "limited", "heritage"},
}

type bundleBlueprint struct {
	key         string
	title       string
	description string
}

var blueprintByIntent = map[string][]bundleBlueprint{
	"everyday": {
		{key: "pairing", title: "Better together", description: "A popular pairing shoppers add with this piece."},
		{key: "complete", title: "Complete the set", description: "Round out your selection with a complementary pick."},
		{key: "upgrade", title: "Elevated duo", description: "Pair with a highly rated companion from the catalog."},
	},
	"workspace": {
		{key: "desk", title: "Minimal workspace kit", description: "Desk-ready pieces that work alongside your main item."},
		{key: "focus", title: "Focused work edit", description: "Calm, compatible accessories for a productive setup."},
		{key: "organize", title: "Organized desk duo", description: "Keep essentials within reach with this pairing."},
	},
	"travel": {
		{key: "carry", title: "Travel-ready set", description: "Compact companions built for life on the move."},
		{key: "pack", title: "Pack smarter", description: "Durable finishes and carry-friendly pairings."},
		{key: "trip", title: "Weekend away", description: "Essentials that travel well with your main piece."},
	},
	"gift": {
		{key: "gift-duo", title: "Gift-worthy duo", description: "A thoughtful pairing for gifting occasions."},
		{key: "premium", title: "Premium presentation", description: "Elevate the gift with a complementary luxury pick."},
		{key: "keepsake", title: "Keepsake pairing", description: "Timeless pieces that present beautifully together."},
	},
}

// SuggestForProduct returns smart bundles anchored on a single product.
func (s *Service) SuggestForProduct(ctx context.Context, productID uint, intent string, limit int) (dto.SmartBundlesResponse, error) {
	products, err := s.repo.GetProductsWithDetails(ctx, []uint{productID})
	if err != nil {
		return dto.SmartBundlesResponse{}, err
	}
	if len(products) == 0 {
		return dto.SmartBundlesResponse{}, utils.ErrNotFound("product not found")
	}
	return s.buildBundles(ctx, []models.Product{products[0]}, normalizeIntent(intent), limit, "")
}

// SuggestForCart returns bundles based on items already in the cart.
func (s *Service) SuggestForCart(ctx context.Context, req dto.SuggestSmartBundlesRequest) (dto.SmartBundlesResponse, error) {
	anchors, err := s.repo.GetProductsWithDetails(ctx, req.ProductIDs)
	if err != nil {
		return dto.SmartBundlesResponse{}, err
	}
	if len(anchors) == 0 {
		return dto.SmartBundlesResponse{Intent: normalizeIntent(req.Intent), Bundles: []dto.SmartBundleItem{}}, nil
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 3
	}
	return s.buildBundles(ctx, anchors, normalizeIntent(req.Intent), limit, "cart")
}

func (s *Service) buildBundles(
	ctx context.Context,
	anchors []models.Product,
	intent string,
	limit int,
	subjectPrefix string,
) (dto.SmartBundlesResponse, error) {
	if limit <= 0 {
		limit = 3
	}
	if limit > 6 {
		limit = 6
	}

	anchor := anchors[0]
	exclude := make([]uint, 0, len(anchors))
	for _, item := range anchors {
		exclude = append(exclude, item.ID)
	}

	keywords := intentKeywords[intent]
	candidates, err := s.repo.FindComplementCandidates(ctx, anchor, exclude, keywords, 16)
	if err != nil {
		return dto.SmartBundlesResponse{}, err
	}
	if len(candidates) == 0 {
		candidates, err = s.repo.FindComplementCandidates(ctx, anchor, exclude, nil, 16)
		if err != nil {
			return dto.SmartBundlesResponse{}, err
		}
	}

	blueprints := blueprintByIntent[intent]
	if len(blueprints) == 0 {
		blueprints = blueprintByIntent["everyday"]
	}

	used := make(map[uint]bool)
	for _, id := range exclude {
		used[id] = true
	}

	bundles := make([]dto.SmartBundleItem, 0, limit)
	candidateIdx := 0

	for i := 0; i < limit && i < len(blueprints); i++ {
		bp := blueprints[i]
		complements := pickComplements(candidates, &candidateIdx, used, 1)
		if len(complements) == 0 {
			break
		}

		products := append([]models.Product{anchor}, complements...)
		if len(anchors) > 1 {
			products = append(anchors, complements...)
		}

		item := dto.SmartBundleItem{
			ID:                 bundleID(subjectPrefix, intent, bp.key, anchor.ID),
			Title:              bp.title,
			Description:        bp.description,
			Intent:             intent,
			CompatibilityScore: scoreBundle(anchor, complements, intent),
			Subtotal:           sumPrices(products),
			Products:           toProductResponses(ctx, products),
		}
		bundles = append(bundles, item)
	}

	if s.ai != nil && s.ai.Enabled() && len(bundles) > 0 {
		enriched, enrichErr := s.ai.EnrichBundleCopy(ctx, subjectPrefix, anchor, intent, bundles)
		if enrichErr == nil {
			bundles = enriched
		}
	}

	return dto.SmartBundlesResponse{Intent: intent, Bundles: bundles}, nil
}

func pickComplements(candidates []*models.Product, idx *int, used map[uint]bool, count int) []models.Product {
	picked := make([]models.Product, 0, count)
	for *idx < len(candidates) && len(picked) < count {
		candidate := candidates[*idx]
		*idx++
		if candidate == nil || used[candidate.ID] {
			continue
		}
		used[candidate.ID] = true
		picked = append(picked, *candidate)
	}
	return picked
}

func scoreBundle(anchor models.Product, complements []models.Product, intent string) int {
	score := 55
	if len(complements) > 0 && complements[0].StoreID == anchor.StoreID {
		score += 15
	}
	if anchor.CategoryID != nil && len(complements) > 0 && complements[0].CategoryID != nil &&
		*complements[0].CategoryID == *anchor.CategoryID {
		score += 10
	}
	if len(complements) > 0 && complements[0].Rating >= 4 {
		score += 10
	}
	if intent != "everyday" {
		score += 5
	}
	if score > 98 {
		return 98
	}
	return score
}

func sumPrices(products []models.Product) float64 {
	var total float64
	for _, product := range products {
		total += product.Price
	}
	return math.Round(total*100) / 100
}

func toProductResponses(ctx context.Context, products []models.Product) []dto.ProductResponse {
	out := make([]dto.ProductResponse, len(products))
	for i := range products {
		out[i] = dto.ToProductResponse(ctx, products[i])
	}
	return out
}

func bundleID(prefix, intent, key string, anchorID uint) string {
	if prefix == "" {
		return fmt.Sprintf("%s-%s-%d", intent, key, anchorID)
	}
	return fmt.Sprintf("%s-%s-%s-%d", prefix, intent, key, anchorID)
}

func normalizeIntent(intent string) string {
	intent = strings.ToLower(strings.TrimSpace(intent))
	switch intent {
	case "workspace", "travel", "gift", "everyday":
		return intent
	default:
		return "everyday"
	}
}

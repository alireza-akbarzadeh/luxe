package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

const (
	TaskOutfitBuilder         = "outfit_builder"
	defaultOutfitSearchLimit  = 3
	maxOutfitPieces           = 6
)

type outfitBuilderPayload struct {
	Summary    string              `json:"summary"`
	StyleTheme string              `json:"style_theme"`
	Pieces     []outfitBuilderSlot `json:"pieces"`
	Tips       []string            `json:"tips"`
}

type outfitBuilderSlot struct {
	Role        string `json:"role"`
	Label       string `json:"label"`
	Reason      string `json:"reason"`
	IsAnchor    bool   `json:"is_anchor"`
	SearchQuery string `json:"search_query"`
}

// OutfitBuilder plans a complete look anchored on a catalog product and fills slots from search.
func (s *Service) OutfitBuilder(
	ctx context.Context,
	subjectKey string,
	search SearchQueries,
	req dto.AiOutfitBuilderRequest,
) (*dto.AiOutfitBuilderResponse, error) {
	if !s.Enabled() {
		return nil, utils.NewAppError(http.StatusServiceUnavailable, "AI is not enabled", nil)
	}
	if search == nil {
		return nil, utils.ErrInternal(fmt.Errorf("search not configured"))
	}
	if !s.chatRL.allow("outfit_builder:"+subjectKey) {
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
	if occasion := strings.TrimSpace(req.Occasion); occasion != "" {
		facts += "\nOccasion: " + sanitizeAIText(occasion) + "\n"
		sources = append(sources, "Occasion")
	}
	if note := strings.TrimSpace(req.Context); note != "" {
		facts += "\nShopper context: " + sanitizeAIText(note) + "\n"
		sources = append(sources, "Shopper context")
	}
	if req.BudgetMax > 0 {
		facts += fmt.Sprintf("\nBudget cap per additional piece: %.0f\n", req.BudgetMax)
		sources = append(sources, "Budget")
	}

	system := `You are a luxury personal stylist building a complete outfit or coordinated look around one anchor catalog product.
Respond with JSON only using this exact shape:
{"summary":"...","style_theme":"...","pieces":[{"role":"anchor","label":"...","reason":"...","is_anchor":true,"search_query":""},...],"tips":["..."]}
Rules:
- summary: 2-3 sentences describing the finished look.
- style_theme: short phrase (e.g. "Smart casual", "Evening minimal").
- pieces: 3-5 slots including exactly ONE anchor (is_anchor true) for the listing product; other slots need complementary items.
- role: anchor | top | bottom | layer | shoes | bag | accessory | jewelry | decor (pick what fits the category).
- label: shopper-friendly slot name.
- reason: one sentence why this slot completes the look.
- search_query: catalog keywords for non-anchor slots; empty for anchor.
- tips: 1-3 styling tips.
Use ONLY product facts below. Do not invent brands or items not inferable from context.
Plain text only — no markdown.

Product facts:
` + facts

	userPrompt := "Build a complete outfit or coordinated look using this product as the anchor."
	if strings.TrimSpace(req.Occasion) != "" {
		userPrompt += " Occasion: " + sanitizeAIText(req.Occasion) + "."
	}

	resp, err := s.complete(ctx, system, userPrompt, 1100)
	if err != nil {
		return nil, utils.NewAppError(http.StatusServiceUnavailable, "AI provider unavailable", err)
	}

	parsed, err := parseOutfitBuilderJSON(resp.Content)
	if err != nil {
		return nil, utils.NewAppError(http.StatusServiceUnavailable, "AI returned invalid outfit plan", err)
	}

	anchorDTO := dto.ToProductResponse(ctx, *product)

	usedProductIDs := map[uint]struct{}{req.ProductID: {}}
	pieces := make([]dto.AiOutfitBuilderPiece, 0, len(parsed.Pieces))
	inStock := true

	for _, slot := range parsed.Pieces {
		if len(pieces) >= maxOutfitPieces {
			break
		}

		label := strings.TrimSpace(sanitizeAIText(slot.Label))
		if label == "" {
			continue
		}

		role := strings.TrimSpace(strings.ToLower(sanitizeAIText(slot.Role)))
		if role == "" {
			role = "accessory"
		}

		piece := dto.AiOutfitBuilderPiece{
			Role:     role,
			Label:    label,
			Reason:   strings.TrimSpace(sanitizeAIText(slot.Reason)),
			IsAnchor: slot.IsAnchor,
		}

		if slot.IsAnchor {
			piece.Product = anchorDTO
			pieces = append(pieces, piece)
			continue
		}

		query := strings.TrimSpace(slot.SearchQuery)
		if query == "" {
			pieces = append(pieces, piece)
			continue
		}

		searchReq := dto.SearchRequest{
			Query:   query,
			Limit:   defaultOutfitSearchLimit + 2,
			Offset:  0,
			InStock: &inStock,
		}
		if req.BudgetMax > 0 {
			searchReq.MaxPrice = req.BudgetMax
		}

		searchResult, searchErr := search.GlobalSearch(ctx, searchReq)
		if searchErr != nil {
			pieces = append(pieces, piece)
			continue
		}

		for _, item := range searchResult.Products {
			if item.ID == 0 {
				continue
			}
			if _, seen := usedProductIDs[item.ID]; seen {
				continue
			}
			usedProductIDs[item.ID] = struct{}{}
			piece.Product = item
			break
		}

		pieces = append(pieces, piece)
	}

	if len(pieces) == 0 {
		pieces = []dto.AiOutfitBuilderPiece{{
			Role:     "anchor",
			Label:    strings.TrimSpace(product.Name),
			Reason:   "Your starting piece for this look.",
			IsAnchor: true,
			Product:  anchorDTO,
		}}
	}

	response := &dto.AiOutfitBuilderResponse{
		Summary:    sanitizeAIText(parsed.Summary),
		StyleTheme: strings.TrimSpace(sanitizeAIText(parsed.StyleTheme)),
		Pieces:     pieces,
		Tips:       sanitizeStringList(parsed.Tips),
		Sources:    uniqueStrings(sources),
	}
	if countCatalogPieces(pieces) > 1 {
		response.Sources = append(response.Sources, "Catalog search")
	}

	utils.Log.WithFields(map[string]any{
		"subject":    subjectKey,
		"task":       TaskOutfitBuilder,
		"product_id": req.ProductID,
		"pieces":     len(pieces),
	}).Info("ai outfit builder completed")

	return response, nil
}

func countCatalogPieces(pieces []dto.AiOutfitBuilderPiece) int {
	count := 0
	for _, piece := range pieces {
		if piece.Product.ID > 0 {
			count++
		}
	}
	return count
}

func parseOutfitBuilderJSON(raw string) (*outfitBuilderPayload, error) {
	raw = extractJSONObject(raw)
	var parsed outfitBuilderPayload
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, err
	}
	if strings.TrimSpace(parsed.Summary) == "" {
		return nil, fmt.Errorf("missing summary")
	}
	if len(parsed.Pieces) == 0 {
		return nil, fmt.Errorf("missing pieces")
	}
	return &parsed, nil
}

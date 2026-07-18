package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	aiint "github.com/alireza-akbarzadeh/luxe/internal/infrastructure/integrations/ai"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

const (
	TaskVirtualTryOn         = "virtual_try_on"
	defaultVirtualTryOnLimit = 4
)

type virtualTryOnPayload struct {
	Summary     string   `json:"summary"`
	FitNotes    string   `json:"fit_notes"`
	StyleMatch  string   `json:"style_match"`
	Confidence  string   `json:"confidence"`
	Tips        []string `json:"tips"`
	SearchQuery string   `json:"search_query"`
}

// VirtualTryOn analyzes a shopper photo against a wearable or personal product listing.
func (s *Service) VirtualTryOn(
	ctx context.Context,
	subjectKey string,
	search SearchQueries,
	req dto.AiVirtualTryOnRequest,
) (*dto.AiVirtualTryOnResponse, error) {
	if !s.Enabled() {
		return nil, utils.NewAppError(http.StatusServiceUnavailable, "AI is not enabled", nil)
	}
	if search == nil {
		return nil, utils.ErrInternal(fmt.Errorf("search not configured"))
	}
	if !s.chatRL.allow("virtual_try_on:" + subjectKey) {
		return nil, utils.ErrTooManyRequests()
	}

	userPhoto, err := normalizeVisualImageDataURL(req.PhotoBase64)
	if err != nil {
		return nil, utils.ErrBadRequest(err.Error())
	}

	product, err := s.products.GetDetailedByID(ctx, req.ProductID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("product not found")
		}
		return nil, utils.ErrInternal(err)
	}

	facts, sources := productFacts(product)
	if profile := strings.TrimSpace(req.SizeProfile); profile != "" {
		facts += "\nShopper size profile: " + sanitizeAIText(profile) + "\n"
		sources = append(sources, "Size profile")
	}
	if note := strings.TrimSpace(req.Context); note != "" {
		facts += "\nShopper note: " + sanitizeAIText(note) + "\n"
		sources = append(sources, "Shopper context")
	}

	images := []string{userPhoto}
	if len(product.Images) > 0 {
		if ref := strings.TrimSpace(product.Images[0]); ref != "" {
			images = append(images, ref)
		}
	}

	system := `You are a luxury personal stylist helping a shopper visualize wearing or using a catalog product.
The first image is the shopper; an optional second image is the product reference photo.
Respond with JSON only using this exact shape:
{"summary":"...","fit_notes":"...","style_match":"strong|good|mixed|weak","confidence":"high|medium|low","tips":["..."],"search_query":"..."}
Rules:
- summary: 2 sentences on how the product would likely look on/for this shopper.
- fit_notes: one sentence on size/proportion/style fit when inferable; empty if unknown.
- style_match: strong, good, mixed, or weak.
- confidence: high, medium, or low based on photo clarity and product type.
- tips: 2-4 styling or sizing tips.
- search_query: keywords for a better-matching alternative if style_match is mixed or weak; else empty.
Be respectful — describe styling only, not body judgment. Use product facts below.
Plain text only — no markdown.

Product facts:
` + facts

	resp, err := s.provider.Complete(ctx, aiint.CompletionRequest{
		Messages: []aiint.Message{
			{Role: "system", Content: system},
			{
				Role:    "user",
				Content: "How would this product suit me in the photo?",
				Images:  images,
			},
		},
		MaxTokens:   900,
		Temperature: 0.25,
	})
	if err != nil {
		return nil, utils.NewAppError(http.StatusServiceUnavailable, "AI provider unavailable", err)
	}

	parsed, err := parseVirtualTryOnJSON(resp.Content)
	if err != nil {
		return nil, utils.NewAppError(http.StatusServiceUnavailable, "AI returned invalid try-on guidance", err)
	}

	response := &dto.AiVirtualTryOnResponse{
		Summary:    parsed.Summary,
		FitNotes:   strings.TrimSpace(parsed.FitNotes),
		StyleMatch: parsed.StyleMatch,
		Confidence: parsed.Confidence,
		Tips:       sanitizeStringList(parsed.Tips),
		Sources:    uniqueStrings(sources),
	}

	query := strings.TrimSpace(parsed.SearchQuery)
	if query == "" || parsed.StyleMatch == "strong" || parsed.StyleMatch == "good" {
		return response, nil
	}

	inStock := true
	searchResult, err := search.GlobalSearch(ctx, dto.SearchRequest{
		Query:   query,
		Limit:   defaultVirtualTryOnLimit + 2,
		Offset:  0,
		InStock: &inStock,
	})
	if err != nil {
		return response, nil
	}

	response.Recommendations = make([]dto.AiRecommendedProduct, 0, defaultVirtualTryOnLimit)
	for _, item := range searchResult.Products {
		if item.ID == 0 || item.ID == req.ProductID {
			continue
		}
		response.Recommendations = append(response.Recommendations, dto.AiRecommendedProduct{
			Product: item,
			Reason:  "May suit your style better.",
		})
		if len(response.Recommendations) >= defaultVirtualTryOnLimit {
			break
		}
	}
	if len(response.Recommendations) > 0 {
		response.Sources = append(response.Sources, "Catalog search")
	}

	utils.Log.WithFields(map[string]any{
		"subject":    subjectKey,
		"task":       TaskVirtualTryOn,
		"product_id": req.ProductID,
	}).Info("ai virtual try-on completed")

	return response, nil
}

func parseVirtualTryOnJSON(raw string) (*virtualTryOnPayload, error) {
	raw = extractJSONObject(raw)
	var parsed virtualTryOnPayload
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, err
	}
	if strings.TrimSpace(parsed.Summary) == "" {
		return nil, fmt.Errorf("missing summary")
	}

	match := strings.TrimSpace(strings.ToLower(parsed.StyleMatch))
	switch match {
	case "strong", "good", "mixed", "weak":
		parsed.StyleMatch = match
	default:
		parsed.StyleMatch = "mixed"
	}

	confidence := strings.TrimSpace(strings.ToLower(parsed.Confidence))
	switch confidence {
	case "high", "medium", "low":
		parsed.Confidence = confidence
	default:
		parsed.Confidence = "medium"
	}

	return &parsed, nil
}

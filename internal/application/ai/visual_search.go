package ai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	aiint "github.com/alireza-akbarzadeh/luxe/internal/infrastructure/integrations/ai"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

const (
	TaskVisualSearch    = "visual_search"
	maxVisualImageBytes = 5 << 20 // 5 MiB decoded
)

type visualSearchIntent struct {
	Interpretation string  `json:"interpretation"`
	SearchQuery    string  `json:"search_query"`
	MinPrice       float64 `json:"min_price"`
	MaxPrice       float64 `json:"max_price"`
	MinRating      float64 `json:"min_rating"`
	Sort           string  `json:"sort"`
}

// VisualSearch analyzes a product photo and returns visually similar catalog matches.
func (s *Service) VisualSearch(
	ctx context.Context,
	subjectKey string,
	search SearchQueries,
	imageBase64 string,
) (*dto.AiVisualSearchResponse, error) {
	if !s.Enabled() {
		return nil, utils.NewAppError(http.StatusServiceUnavailable, "AI is not enabled", nil)
	}
	if search == nil {
		return nil, utils.ErrInternal(fmt.Errorf("search not configured"))
	}
	if !s.chatRL.allow("visual-search:" + subjectKey) {
		return nil, utils.ErrTooManyRequests()
	}

	dataURL, err := normalizeVisualImageDataURL(imageBase64)
	if err != nil {
		return nil, utils.ErrBadRequest(err.Error())
	}

	intent, err := s.extractVisualSearchIntent(ctx, dataURL)
	if err != nil {
		return nil, utils.NewAppError(http.StatusServiceUnavailable, "AI could not analyze this image", err)
	}

	response := &dto.AiVisualSearchResponse{
		Interpretation: intent.Interpretation,
		SearchQuery:    intent.SearchQuery,
		MinPrice:       intent.MinPrice,
		MaxPrice:       intent.MaxPrice,
		MinRating:      intent.MinRating,
		Sort:           normalizeSearchSort(intent.Sort),
	}

	query := strings.TrimSpace(intent.SearchQuery)
	if query == "" {
		return response, nil
	}

	inStock := true
	searchReq := dto.SearchRequest{
		Query:   query,
		Limit:   12,
		Offset:  0,
		Sort:    response.Sort,
		InStock: &inStock,
	}
	if intent.MinPrice > 0 {
		searchReq.MinPrice = intent.MinPrice
	}
	if intent.MaxPrice > 0 {
		searchReq.MaxPrice = intent.MaxPrice
	}
	if intent.MinRating > 0 {
		searchReq.MinRating = intent.MinRating
	}

	searchResult, err := search.GlobalSearch(ctx, searchReq)
	if err != nil {
		return nil, err
	}

	response.Products = searchResult.Products
	response.Total = searchResult.Total

	utils.Log.WithFields(map[string]any{
		"subject": subjectKey,
		"task":    TaskVisualSearch,
		"query":   query,
		"total":   searchResult.Total,
	}).Info("ai visual search completed")

	return response, nil
}

func (s *Service) extractVisualSearchIntent(ctx context.Context, dataURL string) (*visualSearchIntent, error) {
	system := `You analyze product photos for a luxury e-commerce catalog.
Describe what you see and return JSON only:
{"interpretation":"one sentence for the shopper","search_query":"short catalog keywords","min_price":0,"max_price":0,"min_rating":0,"sort":"rating_desc"}
Rules:
- interpretation: friendly summary of the item style/category/colors (max 20 words).
- search_query: 2-6 keywords to find similar products (category, material, color, style). Never empty.
- min_price / max_price: 0 when unknown.
- min_rating: 0-5 when quality cues matter; else 0.
- sort: rating_desc, price_asc, price_desc, newest, or popular.
Focus on the main product in the image. Ignore backgrounds and people unless they wear the shoppable item.`

	resp, err := s.provider.Complete(ctx, aiint.CompletionRequest{
		Messages: []aiint.Message{
			{Role: "system", Content: system},
			{
				Role:    "user",
				Content: "Find catalog products that look like this item.",
				Images:  []string{dataURL},
			},
		},
		MaxTokens:   500,
		Temperature: 0.2,
	})
	if err != nil {
		return nil, err
	}

	var intent visualSearchIntent
	trimmed := extractJSONObject(resp.Content)
	if err := json.Unmarshal([]byte(trimmed), &intent); err != nil {
		return nil, err
	}

	intent.Interpretation = sanitizeAIText(intent.Interpretation)
	intent.SearchQuery = sanitizeAIText(intent.SearchQuery)
	if intent.Interpretation == "" {
		intent.Interpretation = "Looking for products similar to your photo."
	}
	if intent.SearchQuery == "" {
		return nil, fmt.Errorf("empty search query from vision model")
	}
	return &intent, nil
}

func normalizeVisualImageDataURL(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("image is required")
	}

	var mime string
	var encoded string

	if strings.HasPrefix(trimmed, "data:") {
		comma := strings.Index(trimmed, ",")
		if comma < 0 {
			return "", fmt.Errorf("invalid image data URL")
		}
		header := trimmed[:comma]
		encoded = trimmed[comma+1:]
		if !strings.Contains(header, "base64") {
			return "", fmt.Errorf("image must be base64 encoded")
		}
		mime = strings.TrimPrefix(header, "data:")
		mime = strings.TrimSuffix(mime, ";base64")
	} else {
		encoded = trimmed
		mime = "image/jpeg"
	}

	switch mime {
	case "image/jpeg", "image/png", "image/webp", "image/gif":
	default:
		return "", fmt.Errorf("unsupported image type")
	}

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("invalid base64 image")
	}
	if len(decoded) == 0 {
		return "", fmt.Errorf("image is empty")
	}
	if len(decoded) > maxVisualImageBytes {
		return "", fmt.Errorf("image must be 5 MB or smaller")
	}

	return fmt.Sprintf("data:%s;base64,%s", mime, encoded), nil
}

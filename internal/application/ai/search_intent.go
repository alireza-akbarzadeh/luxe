package ai

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	aiint "github.com/alireza-akbarzadeh/luxe/internal/infrastructure/integrations/ai"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

const TaskSearchIntent = "search_intent"

type searchQueryIntent struct {
	IsIntentQuery  bool    `json:"is_intent_query"`
	Interpretation string  `json:"interpretation"`
	SearchQuery    string  `json:"search_query"`
	MinPrice       float64 `json:"min_price"`
	MaxPrice       float64 `json:"max_price"`
	MinRating      float64 `json:"min_rating"`
	Sort           string  `json:"sort"`
	InStock        *bool   `json:"in_stock"`
	OnSale         *bool   `json:"on_sale"`
	IsNew          *bool   `json:"is_new"`
	IsDigital      *bool   `json:"is_digital"`
}

// ParseSearchIntent turns a natural-language search phrase into catalog filters.
func (s *Service) ParseSearchIntent(
	ctx context.Context,
	subjectKey string,
	query string,
) (*dto.AiSearchIntentResponse, error) {
	clean := sanitizeAIText(query)
	if clean == "" {
		return nil, utils.ErrBadRequest("query is required")
	}

	fallback := &dto.AiSearchIntentResponse{
		IsIntentQuery: false,
		SearchQuery:   clean,
	}

	if !s.Enabled() {
		return fallback, nil
	}
	if !s.chatRL.allow("search-intent:" + subjectKey) {
		return nil, utils.ErrTooManyRequests()
	}

	intent, err := s.extractSearchQueryIntent(ctx, clean)
	if err != nil {
		return fallback, nil
	}

	return intentToDTO(intent), nil
}

func intentToDTO(intent *searchQueryIntent) *dto.AiSearchIntentResponse {
	if intent == nil {
		return &dto.AiSearchIntentResponse{IsIntentQuery: false}
	}
	return &dto.AiSearchIntentResponse{
		IsIntentQuery:  intent.IsIntentQuery,
		Interpretation: sanitizeAIText(intent.Interpretation),
		SearchQuery:    sanitizeAIText(intent.SearchQuery),
		MinPrice:       intent.MinPrice,
		MaxPrice:       intent.MaxPrice,
		MinRating:      intent.MinRating,
		Sort:           normalizeSearchSort(intent.Sort),
		InStock:        intent.InStock,
		OnSale:         intent.OnSale,
		IsNew:          intent.IsNew,
		IsDigital:      intent.IsDigital,
	}
}

func (s *Service) extractSearchQueryIntent(ctx context.Context, query string) (*searchQueryIntent, error) {
	system := `You parse e-commerce search queries into catalog filters. Return JSON only:
{"is_intent_query":true,"interpretation":"...","search_query":"...","min_price":0,"max_price":0,"min_rating":0,"sort":"rating_desc","in_stock":null,"on_sale":null,"is_new":null,"is_digital":null}
Rules:
- is_intent_query: true when the phrase expresses shopping intent (budget, occasion, style, constraints). false for simple product keywords (e.g. "nike", "watch", "sku-123").
- interpretation: one short sentence summarizing what the shopper wants (empty when is_intent_query is false).
- search_query: concise catalog keywords (2-6 words). For keyword-only queries, echo the query trimmed.
- min_price / max_price: numbers in store currency; 0 when unknown.
- min_rating: 0-5; 0 when unknown.
- sort: one of rating_desc, price_asc, price_desc, newest, popular, or empty for relevance.
- in_stock / on_sale / is_new / is_digital: true only when explicitly requested; null otherwise.
Do not invent brands. Do not request personal data.`

	resp, err := s.provider.Complete(ctx, aiint.CompletionRequest{
		Messages: []aiint.Message{
			{Role: "system", Content: system},
			{Role: "user", Content: query},
		},
		MaxTokens:   400,
		Temperature: 0.2,
	})
	if err != nil {
		return nil, err
	}

	payload, err := parseSearchQueryIntentJSON(resp.Content)
	if err != nil {
		return nil, err
	}

	if payload.SearchQuery == "" {
		payload.SearchQuery = query
	}
	if payload.IsIntentQuery && payload.Interpretation == "" {
		payload.Interpretation = "Showing results matching your request."
	}

	return payload, nil
}

func parseSearchQueryIntentJSON(content string) (*searchQueryIntent, error) {
	trimmed := extractJSONObject(content)
	var intent searchQueryIntent
	if err := json.Unmarshal([]byte(trimmed), &intent); err != nil {
		return nil, err
	}
	intent.SearchQuery = strings.TrimSpace(intent.SearchQuery)
	return &intent, nil
}

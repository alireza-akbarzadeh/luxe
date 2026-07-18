package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	aiint "github.com/alireza-akbarzadeh/luxe/internal/infrastructure/integrations/ai"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

const TaskShoppingAssistant = "shopping_assistant"

// SearchQueries runs catalog search for assistant recommendations.
type SearchQueries interface {
	GlobalSearch(ctx context.Context, req dto.SearchRequest) (*dto.SearchResponse, error)
}

type shoppingIntent struct {
	Reply             string   `json:"reply"`
	FollowUpQuestions []string `json:"follow_up_questions"`
	SearchQuery       string   `json:"search_query"`
	MinPrice          float64  `json:"min_price"`
	MaxPrice          float64  `json:"max_price"`
	Sort              string   `json:"sort"`
}

type productReasonsPayload struct {
	Reasons map[string]string `json:"reasons"`
}

// ShoppingAssistant handles store-wide conversational product discovery.
func (s *Service) ShoppingAssistant(
	ctx context.Context,
	subjectKey string,
	search SearchQueries,
	req dto.AiShoppingAssistantRequest,
) (*dto.AiShoppingAssistantResponse, error) {
	if !s.Enabled() {
		return nil, utils.NewAppError(503, "AI is not enabled", nil)
	}
	if search == nil {
		return nil, utils.ErrInternal(fmt.Errorf("search not configured"))
	}
	if !s.chatRL.allow("assistant:" + subjectKey) {
		return nil, utils.ErrTooManyRequests()
	}

	messages := sanitizeAssistantMessages(req.Messages)
	if len(messages) == 0 {
		return nil, utils.ErrBadRequest("at least one user message is required")
	}

	intent, err := s.extractShoppingIntent(ctx, messages)
	if err != nil {
		return nil, utils.NewAppError(503, "AI provider unavailable", err)
	}

	response := &dto.AiShoppingAssistantResponse{
		Reply:             intent.Reply,
		FollowUpQuestions: intent.FollowUpQuestions,
		Sources:           []string{"catalog search"},
	}

	query := strings.TrimSpace(intent.SearchQuery)
	if query == "" {
		return response, nil
	}

	inStock := true
	searchReq := dto.SearchRequest{
		Query:   query,
		Limit:   6,
		Offset:  0,
		Sort:    normalizeSearchSort(intent.Sort),
		InStock: &inStock,
	}
	if intent.MinPrice > 0 {
		searchReq.MinPrice = intent.MinPrice
	}
	if intent.MaxPrice > 0 {
		searchReq.MaxPrice = intent.MaxPrice
	}

	searchResult, err := search.GlobalSearch(ctx, searchReq)
	if err != nil {
		return nil, err
	}
	if len(searchResult.Products) == 0 {
		if len(response.FollowUpQuestions) == 0 {
			response.FollowUpQuestions = []string{
				"Would you like to adjust your budget?",
				"Do you have a preferred brand or color?",
			}
		}
		return response, nil
	}

	reasons, err := s.explainRecommendations(ctx, messages, searchResult.Products)
	if err != nil {
		reasons = map[uint]string{}
	}

	recommendations := make([]dto.AiRecommendedProduct, 0, len(searchResult.Products))
	for _, product := range searchResult.Products {
		reason := reasons[product.ID]
		if reason == "" {
			reason = "Matches your shopping request."
		}
		recommendations = append(recommendations, dto.AiRecommendedProduct{
			Product: product,
			Reason:  reason,
		})
	}
	response.Recommendations = recommendations

	return response, nil
}

func sanitizeAssistantMessages(messages []dto.AiChatMessage) []dto.AiChatMessage {
	out := make([]dto.AiChatMessage, 0, len(messages))
	for _, m := range messages {
		role := strings.TrimSpace(m.Role)
		if role != "user" && role != "assistant" {
			continue
		}
		content := sanitizeAIText(m.Content)
		if content == "" {
			continue
		}
		out = append(out, dto.AiChatMessage{Role: role, Content: content})
	}
	return out
}

func (s *Service) extractShoppingIntent(ctx context.Context, messages []dto.AiChatMessage) (*shoppingIntent, error) {
	system := `You are a luxury e-commerce shopping assistant.
Read the conversation and respond with JSON only:
{"reply":"...","follow_up_questions":["..."],"search_query":"...","min_price":0,"max_price":0,"sort":"rating_desc"}
Rules:
- reply: friendly, concise (2-4 sentences), acknowledge the shopper's intent.
- follow_up_questions: 0-2 short questions when key info is missing (budget, occasion, style). Empty array when ready to search.
- search_query: short catalog keywords to find products. Empty only when you must ask follow-ups first.
- min_price / max_price: numbers in store currency; use 0 when unknown.
- sort: one of rating_desc, price_asc, price_desc, newest, popular.
Do not invent inventory. Do not request personal contact info.`

	llmMessages := []aiint.Message{{Role: "system", Content: system}}
	for _, m := range messages {
		llmMessages = append(llmMessages, aiint.Message{Role: m.Role, Content: m.Content})
	}

	resp, err := s.provider.Complete(ctx, aiint.CompletionRequest{
		Messages:    llmMessages,
		MaxTokens:   700,
		Temperature: 0.35,
	})
	if err != nil {
		return nil, err
	}

	intent, err := parseShoppingIntentJSON(resp.Content)
	if err != nil {
		return nil, err
	}
	intent.Reply = sanitizeAIText(intent.Reply)
	intent.FollowUpQuestions = sanitizeStringList(intent.FollowUpQuestions)
	intent.SearchQuery = sanitizeAIText(intent.SearchQuery)
	if intent.Reply == "" {
		intent.Reply = "I'd be happy to help you find the right product."
	}
	return intent, nil
}

func (s *Service) explainRecommendations(
	ctx context.Context,
	messages []dto.AiChatMessage,
	products []dto.ProductResponse,
) (map[uint]string, error) {
	var catalog strings.Builder
	catalog.WriteString("Products:\n")
	for _, p := range products {
		fmt.Fprintf(&catalog, "- id=%d name=%s price=%.2f rating=%.1f\n", p.ID, p.Name, p.Price, p.Rating)
	}

	var conv strings.Builder
	for _, m := range messages {
		fmt.Fprintf(&conv, "%s: %s\n", m.Role, m.Content)
	}

	system := `Given the shopper conversation and product list, return JSON only:
{"reasons":{"<product_id>":"one short sentence why it fits"}}
Use only listed product ids. Plain text reasons, no markdown.`

	user := conv.String() + "\n" + catalog.String()
	resp, err := s.complete(ctx, system, user, 500)
	if err != nil {
		return nil, err
	}

	payload, err := parseProductReasonsJSON(resp.Content)
	if err != nil {
		return nil, err
	}

	out := make(map[uint]string, len(payload.Reasons))
	for idStr, reason := range payload.Reasons {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			continue
		}
		clean := sanitizeAIText(reason)
		if clean != "" {
			out[uint(id)] = clean
		}
	}
	return out, nil
}

func parseShoppingIntentJSON(content string) (*shoppingIntent, error) {
	trimmed := extractJSONObject(content)
	var intent shoppingIntent
	if err := json.Unmarshal([]byte(trimmed), &intent); err != nil {
		return nil, err
	}
	return &intent, nil
}

func parseProductReasonsJSON(content string) (*productReasonsPayload, error) {
	trimmed := extractJSONObject(content)
	var payload productReasonsPayload
	if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
		return nil, err
	}
	if payload.Reasons == nil {
		payload.Reasons = map[string]string{}
	}
	return &payload, nil
}

func extractJSONObject(content string) string {
	trimmed := strings.TrimSpace(content)
	if idx := strings.Index(trimmed, "{"); idx >= 0 {
		if end := strings.LastIndex(trimmed, "}"); end > idx {
			return trimmed[idx : end+1]
		}
	}
	return trimmed
}

func normalizeSearchSort(sort string) string {
	switch strings.TrimSpace(sort) {
	case "price_asc", "price_desc", "rating_desc", "newest", "popular":
		return sort
	default:
		return "rating_desc"
	}
}

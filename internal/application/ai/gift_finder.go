package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	aiint "github.com/alireza-akbarzadeh/luxe/internal/infrastructure/integrations/ai"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

const TaskGiftFinder = "gift_finder"

type giftFinderIntent struct {
	Reply             string   `json:"reply"`
	FollowUpQuestions []string `json:"follow_up_questions"`
	SearchQuery       string   `json:"search_query"`
	MinPrice          float64  `json:"min_price"`
	MaxPrice          float64  `json:"max_price"`
	Sort              string   `json:"sort"`
	GiftMessageIdeas  []string `json:"gift_message_ideas"`
}

// GiftFinder recommends catalog gifts from structured recipient, occasion, and budget inputs.
func (s *Service) GiftFinder(
	ctx context.Context,
	subjectKey string,
	search SearchQueries,
	req dto.AiGiftFinderRequest,
) (*dto.AiGiftFinderResponse, error) {
	if !s.Enabled() {
		return nil, utils.NewAppError(503, "AI is not enabled", nil)
	}
	if search == nil {
		return nil, utils.ErrInternal(fmt.Errorf("search not configured"))
	}
	if !s.chatRL.allow("gift_finder:" + subjectKey) {
		return nil, utils.ErrTooManyRequests()
	}

	intent, err := s.extractGiftFinderIntent(ctx, req)
	if err != nil {
		return nil, utils.NewAppError(503, "AI provider unavailable", err)
	}

	response := &dto.AiGiftFinderResponse{
		Reply:             intent.Reply,
		FollowUpQuestions: intent.FollowUpQuestions,
		GiftMessageIdeas:  sanitizeStringList(intent.GiftMessageIdeas),
		Sources:           []string{"catalog search"},
	}

	query := strings.TrimSpace(intent.SearchQuery)
	if query == "" {
		return response, nil
	}

	inStock := true
	searchReq := dto.SearchRequest{
		Query:   query,
		Limit:   8,
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
	if req.BudgetMin > 0 && searchReq.MinPrice == 0 {
		searchReq.MinPrice = req.BudgetMin
	}
	if req.BudgetMax > 0 && searchReq.MaxPrice == 0 {
		searchReq.MaxPrice = req.BudgetMax
	}

	searchResult, err := search.GlobalSearch(ctx, searchReq)
	if err != nil {
		return nil, err
	}
	if len(searchResult.Products) == 0 {
		if len(response.FollowUpQuestions) == 0 {
			response.FollowUpQuestions = []string{
				"Would a higher or lower budget work better?",
				"Any hobbies or brands they love?",
			}
		}
		return response, nil
	}

	reasons, err := s.explainGiftRecommendations(ctx, req, searchResult.Products)
	if err != nil {
		reasons = map[uint]string{}
	}

	recommendations := make([]dto.AiRecommendedProduct, 0, len(searchResult.Products))
	for _, product := range searchResult.Products {
		reason := reasons[product.ID]
		if reason == "" {
			reason = "A thoughtful pick for this gift occasion."
		}
		recommendations = append(recommendations, dto.AiRecommendedProduct{
			Product: product,
			Reason:  reason,
		})
	}
	response.Recommendations = recommendations

	return response, nil
}

func (s *Service) extractGiftFinderIntent(ctx context.Context, req dto.AiGiftFinderRequest) (*giftFinderIntent, error) {
	system := `You are a luxury e-commerce gift concierge.
Given structured gift-finder inputs, respond with JSON only:
{"reply":"...","follow_up_questions":["..."],"search_query":"...","min_price":0,"max_price":0,"sort":"rating_desc","gift_message_ideas":["..."]}
Rules:
- reply: warm, concise (2-4 sentences) acknowledging the gift context.
- follow_up_questions: 0-2 short questions ONLY when you cannot search yet (missing age, style, or gender context). Empty when ready.
- search_query: short catalog keywords for giftable products. Empty only when follow-ups are required first.
- min_price / max_price: numbers in store currency; honor shopper budget hints; use 0 when open-ended.
- sort: one of rating_desc, price_asc, price_desc, newest, popular.
- gift_message_ideas: 0-2 short note ideas for a gift card message (optional).
Focus on gift appropriateness, not generic shopping. Do not invent inventory.`

	userPrompt := buildGiftFinderPrompt(req)
	llmMessages := []aiint.Message{
		{Role: "system", Content: system},
		{Role: "user", Content: userPrompt},
	}

	resp, err := s.provider.Complete(ctx, aiint.CompletionRequest{
		Messages:    llmMessages,
		MaxTokens:   800,
		Temperature: 0.35,
	})
	if err != nil {
		return nil, err
	}

	intent, err := parseGiftFinderIntentJSON(resp.Content)
	if err != nil {
		return nil, err
	}
	intent.Reply = sanitizeAIText(intent.Reply)
	intent.FollowUpQuestions = sanitizeStringList(intent.FollowUpQuestions)
	intent.SearchQuery = sanitizeAIText(intent.SearchQuery)
	intent.GiftMessageIdeas = sanitizeStringList(intent.GiftMessageIdeas)
	if intent.Reply == "" {
		intent.Reply = "I'd love to help you find a thoughtful gift."
	}
	return intent, nil
}

func buildGiftFinderPrompt(req dto.AiGiftFinderRequest) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Recipient: %s\n", sanitizeAIText(req.Recipient))
	fmt.Fprintf(&b, "Occasion: %s\n", sanitizeAIText(req.Occasion))
	if req.BudgetMin > 0 || req.BudgetMax > 0 {
		fmt.Fprintf(&b, "Budget: %.0f", req.BudgetMin)
		if req.BudgetMax > 0 {
			fmt.Fprintf(&b, " – %.0f", req.BudgetMax)
		}
		b.WriteString("\n")
	}
	if interests := sanitizeAIText(req.Interests); interests != "" {
		fmt.Fprintf(&b, "Interests & style: %s\n", interests)
	}
	if notes := sanitizeAIText(req.AdditionalNotes); notes != "" {
		fmt.Fprintf(&b, "Extra notes: %s\n", notes)
	}
	for _, fa := range req.FollowUpAnswers {
		q := sanitizeAIText(fa.Question)
		a := sanitizeAIText(fa.Answer)
		if q != "" && a != "" {
			fmt.Fprintf(&b, "Q: %s\nA: %s\n", q, a)
		}
	}
	return b.String()
}

func (s *Service) explainGiftRecommendations(
	ctx context.Context,
	req dto.AiGiftFinderRequest,
	products []dto.ProductResponse,
) (map[uint]string, error) {
	var catalog strings.Builder
	catalog.WriteString("Products:\n")
	for _, p := range products {
		fmt.Fprintf(&catalog, "- id=%d name=%s price=%.2f rating=%.1f\n", p.ID, p.Name, p.Price, p.Rating)
	}

	system := `Given gift context and products, return JSON only:
{"reasons":{"<product_id>":"one short sentence why this is a good gift"}}
Use only listed product ids. Plain text, no markdown.`

	user := buildGiftFinderPrompt(req) + "\n" + catalog.String()
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

func parseGiftFinderIntentJSON(content string) (*giftFinderIntent, error) {
	trimmed := extractJSONObject(content)
	var intent giftFinderIntent
	if err := json.Unmarshal([]byte(trimmed), &intent); err != nil {
		return nil, err
	}
	return &intent, nil
}

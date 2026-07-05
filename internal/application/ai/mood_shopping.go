package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

const TaskMoodShopping = "mood_shopping"

type moodShoppingIntent struct {
	Reply             string   `json:"reply"`
	MoodTags          []string `json:"mood_tags"`
	StyleCues         []string `json:"style_cues"`
	FollowUpQuestions []string `json:"follow_up_questions"`
	SearchQuery       string   `json:"search_query"`
	MinPrice          float64  `json:"min_price"`
	MaxPrice          float64  `json:"max_price"`
	Sort              string   `json:"sort"`
}

// MoodShopping recommends catalog products aligned with a shopper's mood or vibe.
func (s *Service) MoodShopping(
	ctx context.Context,
	subjectKey string,
	search SearchQueries,
	req dto.AiMoodShoppingRequest,
) (*dto.AiMoodShoppingResponse, error) {
	if !s.Enabled() {
		return nil, utils.NewAppError(503, "AI is not enabled", nil)
	}
	if search == nil {
		return nil, utils.ErrInternal(fmt.Errorf("search not configured"))
	}
	if !s.chatRL.allow("mood_shopping:" + subjectKey) {
		return nil, utils.ErrTooManyRequests()
	}

	intent, err := s.extractMoodShoppingIntent(ctx, req)
	if err != nil {
		return nil, utils.NewAppError(503, "AI provider unavailable", err)
	}

	response := &dto.AiMoodShoppingResponse{
		Reply:             intent.Reply,
		MoodTags:          sanitizeStringList(intent.MoodTags),
		StyleCues:         sanitizeStringList(intent.StyleCues),
		FollowUpQuestions: sanitizeStringList(intent.FollowUpQuestions),
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

	for _, product := range searchResult.Products {
		response.Recommendations = append(response.Recommendations, dto.AiRecommendedProduct{
			Product: product,
			Reason:  "Matches your mood and style cues.",
		})
	}

	utils.Log.WithFields(map[string]any{
		"subject": subjectKey,
		"task":    TaskMoodShopping,
		"mood":    sanitizeAIText(req.Mood),
	}).Info("ai mood shopping completed")

	return response, nil
}

func (s *Service) extractMoodShoppingIntent(ctx context.Context, req dto.AiMoodShoppingRequest) (*moodShoppingIntent, error) {
	facts := moodShoppingFacts(req)
	system := `You help shoppers discover products that match their mood on a luxury marketplace.
Respond with JSON only using this exact shape:
{"reply":"...","mood_tags":["..."],"style_cues":["..."],"follow_up_questions":["..."],"search_query":"...","min_price":0,"max_price":0,"sort":"relevance|price_asc|price_desc|rating"}
Rules:
- reply: 2 sentences reflecting the mood and shopping approach.
- mood_tags: 2-4 short mood or vibe labels.
- style_cues: 3-5 product/style attributes to look for (colors, materials, silhouettes).
- follow_up_questions: 0-2 clarifiers only when mood is vague; otherwise empty array.
- search_query: short catalog keywords for mood-aligned products. Empty only when follow-ups are required first.
- min_price/max_price: use budget hints when provided; 0 when unknown.
- sort: relevance by default.
Plain text only — no markdown.

Mood inputs:
` + facts

	user := "Suggest a mood-aligned product discovery query."
	resp, err := s.complete(ctx, system, user, 900)
	if err != nil {
		return nil, err
	}
	return parseMoodShoppingJSON(resp.Content)
}

func moodShoppingFacts(req dto.AiMoodShoppingRequest) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Mood: %s\n", sanitizeAIText(req.Mood))
	if req.Context != "" {
		fmt.Fprintf(&b, "Context: %s\n", sanitizeAIText(req.Context))
	}
	if req.BudgetMin > 0 || req.BudgetMax > 0 {
		fmt.Fprintf(&b, "Budget: %.0f - %.0f\n", req.BudgetMin, req.BudgetMax)
	}
	return b.String()
}

func parseMoodShoppingJSON(raw string) (*moodShoppingIntent, error) {
	raw = extractJSONObject(raw)
	var parsed moodShoppingIntent
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, err
	}
	parsed.Reply = strings.TrimSpace(parsed.Reply)
	if parsed.Reply == "" {
		return nil, fmt.Errorf("missing reply")
	}
	return &parsed, nil
}

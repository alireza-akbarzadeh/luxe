package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

const TaskGoalShopping = "goal_shopping"

type goalShoppingIntent struct {
	Reply             string   `json:"reply"`
	Steps             []string `json:"steps"`
	FollowUpQuestions []string `json:"follow_up_questions"`
	SearchQuery       string   `json:"search_query"`
	MinPrice          float64  `json:"min_price"`
	MaxPrice          float64  `json:"max_price"`
	Sort              string   `json:"sort"`
}

// GoalShopping recommends products and a short plan for a stated shopping goal.
func (s *Service) GoalShopping(
	ctx context.Context,
	subjectKey string,
	search SearchQueries,
	req dto.AiGoalShoppingRequest,
) (*dto.AiGoalShoppingResponse, error) {
	if !s.Enabled() {
		return nil, utils.NewAppError(503, "AI is not enabled", nil)
	}
	if search == nil {
		return nil, utils.ErrInternal(fmt.Errorf("search not configured"))
	}
	if !s.chatRL.allow("goal_shopping:" + subjectKey) {
		return nil, utils.ErrTooManyRequests()
	}

	intent, err := s.extractGoalShoppingIntent(ctx, req)
	if err != nil {
		return nil, utils.NewAppError(503, "AI provider unavailable", err)
	}

	response := &dto.AiGoalShoppingResponse{
		Reply:             intent.Reply,
		Steps:             sanitizeStringList(intent.Steps),
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
			Reason:  "Aligned with your stated goal.",
		})
	}

	utils.Log.WithFields(map[string]any{
		"subject": subjectKey,
		"task":    TaskGoalShopping,
		"goal":    sanitizeAIText(req.Goal),
	}).Info("ai goal shopping completed")

	return response, nil
}

func (s *Service) extractGoalShoppingIntent(ctx context.Context, req dto.AiGoalShoppingRequest) (*goalShoppingIntent, error) {
	facts := goalShoppingFacts(req)
	system := `You help shoppers reach a specific shopping goal on a luxury marketplace.
Respond with JSON only using this exact shape:
{"reply":"...","steps":["..."],"follow_up_questions":["..."],"search_query":"...","min_price":0,"max_price":0,"sort":"relevance|price_asc|price_desc|rating"}
Rules:
- reply: 2 sentences acknowledging the goal and approach.
- steps: 3-5 actionable shopping steps (prioritize, compare, budget).
- follow_up_questions: 0-2 clarifiers only when the goal is vague; otherwise empty array.
- search_query: short catalog keywords to find goal-aligned products. Empty only when follow-ups are required first.
- min_price/max_price: use budget hints when provided; 0 when unknown.
- sort: relevance by default; price_asc for budget goals.
Plain text only — no markdown.

Goal inputs:
` + facts

	user := "Plan product discovery for this shopping goal."
	resp, err := s.complete(ctx, system, user, 900)
	if err != nil {
		return nil, err
	}
	return parseGoalShoppingJSON(resp.Content)
}

func goalShoppingFacts(req dto.AiGoalShoppingRequest) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Goal: %s\n", sanitizeAIText(req.Goal))
	if req.Timeline != "" {
		fmt.Fprintf(&b, "Timeline: %s\n", sanitizeAIText(req.Timeline))
	}
	if req.Preferences != "" {
		fmt.Fprintf(&b, "Preferences: %s\n", sanitizeAIText(req.Preferences))
	}
	if req.BudgetMin > 0 || req.BudgetMax > 0 {
		fmt.Fprintf(&b, "Budget: %.0f - %.0f\n", req.BudgetMin, req.BudgetMax)
	}
	return b.String()
}

func parseGoalShoppingJSON(raw string) (*goalShoppingIntent, error) {
	raw = extractJSONObject(raw)
	var parsed goalShoppingIntent
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, err
	}
	parsed.Reply = strings.TrimSpace(parsed.Reply)
	if parsed.Reply == "" {
		return nil, fmt.Errorf("missing reply")
	}
	return &parsed, nil
}

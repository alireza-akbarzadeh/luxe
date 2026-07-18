package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

const (
	TaskHouseholdShopping       = "household_shopping"
	maxHouseholdMembers         = 8
	defaultHouseholdSearchLimit = 5
)

type householdShoppingPayload struct {
	Summary string                         `json:"summary"`
	Members []householdMemberSearchPayload `json:"members"`
}

type householdMemberSearchPayload struct {
	MemberName  string `json:"member_name"`
	Summary     string `json:"summary"`
	SearchQuery string `json:"search_query"`
}

// HouseholdShopping recommends catalog picks for each household member profile.
func (s *Service) HouseholdShopping(
	ctx context.Context,
	userID uint,
	subjectKey string,
	search SearchQueries,
	req dto.AiHouseholdShoppingRequest,
) (*dto.AiHouseholdShoppingResponse, error) {
	if !s.Enabled() {
		return nil, utils.NewAppError(503, "AI is not enabled", nil)
	}
	if userID == 0 {
		return nil, utils.ErrUnauthorized("authentication required")
	}
	if search == nil {
		return nil, utils.ErrInternal(fmt.Errorf("search not configured"))
	}
	if !s.chatRL.allow("household_shopping:" + subjectKey) {
		return nil, utils.ErrTooManyRequests()
	}

	members := sanitizeHouseholdMembers(req.Members)
	if len(members) == 0 {
		return nil, utils.ErrBadRequest("at least one household member is required")
	}

	facts := householdFacts(members, req.Context, req.BudgetMin, req.BudgetMax)
	system := `You personalize luxury catalog shopping for each member of a household.
Respond with JSON only using this exact shape:
{"summary":"...","members":[{"member_name":"...","summary":"...","search_query":"..."}]}
Rules:
- summary: 2 sentences on how to shop for this household together.
- members: one entry per household member from the facts; member_name must match exactly.
- summary: 1-2 sentences on what fits that person.
- search_query: short catalog keywords for 3-5 products for that member; never empty.
Respect sizes, preferences, and interests from the facts. Do not invent people.
Plain text only — no markdown.

Household facts:
` + facts

	user := "Recommend a search query per household member for personalized product picks."
	resp, err := s.complete(ctx, system, user, 1200)
	if err != nil {
		return nil, utils.NewAppError(503, "AI provider unavailable", err)
	}

	parsed, err := parseHouseholdShoppingJSON(resp.Content, len(members))
	if err != nil {
		return nil, utils.NewAppError(503, "AI returned invalid household shopping plan", err)
	}

	response := &dto.AiHouseholdShoppingResponse{
		Summary: parsed.Summary,
		Sources: []string{"Household profiles", "Catalog search"},
	}

	inStock := true
	seenProduct := make(map[uint]struct{})
	for _, member := range parsed.Members {
		pick := dto.AiHouseholdMemberPick{
			MemberName: member.MemberName,
			Summary:    member.Summary,
		}

		query := strings.TrimSpace(member.SearchQuery)
		if query == "" {
			response.Members = append(response.Members, pick)
			continue
		}

		searchReq := dto.SearchRequest{
			Query:   query,
			Limit:   defaultHouseholdSearchLimit,
			Offset:  0,
			InStock: &inStock,
		}
		if req.BudgetMin > 0 {
			searchReq.MinPrice = req.BudgetMin
		}
		if req.BudgetMax > 0 {
			searchReq.MaxPrice = req.BudgetMax
		}

		searchResult, err := search.GlobalSearch(ctx, searchReq)
		if err != nil {
			response.Members = append(response.Members, pick)
			continue
		}

		for _, product := range searchResult.Products {
			if product.ID == 0 {
				continue
			}
			if _, ok := seenProduct[product.ID]; ok {
				continue
			}
			seenProduct[product.ID] = struct{}{}
			pick.Recommendations = append(pick.Recommendations, dto.AiRecommendedProduct{
				Product: product,
				Reason:  pick.Summary,
			})
			if len(pick.Recommendations) >= 4 {
				break
			}
		}

		response.Members = append(response.Members, pick)
	}

	utils.Log.WithFields(map[string]any{
		"user_id": userID,
		"task":    TaskHouseholdShopping,
		"members": len(members),
	}).Info("ai household shopping completed")

	return response, nil
}

func sanitizeHouseholdMembers(members []dto.AiHouseholdMemberProfile) []dto.AiHouseholdMemberProfile {
	clean := make([]dto.AiHouseholdMemberProfile, 0, len(members))
	for _, member := range members {
		name := strings.TrimSpace(member.Name)
		if name == "" {
			continue
		}
		clean = append(clean, dto.AiHouseholdMemberProfile{
			Name:         name,
			Relationship: strings.TrimSpace(member.Relationship),
			Sizes:        strings.TrimSpace(member.Sizes),
			Preferences:  strings.TrimSpace(member.Preferences),
			Interests:    strings.TrimSpace(member.Interests),
		})
		if len(clean) >= maxHouseholdMembers {
			break
		}
	}
	return clean
}

func householdFacts(
	members []dto.AiHouseholdMemberProfile,
	contextNote string,
	budgetMin, budgetMax float64,
) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%d household members:\n", len(members))
	for _, member := range members {
		fmt.Fprintf(&b, "- %s", sanitizeAIText(member.Name))
		if rel := member.Relationship; rel != "" {
			fmt.Fprintf(&b, " (%s)", sanitizeAIText(rel))
		}
		b.WriteString("\n")
		if member.Sizes != "" {
			fmt.Fprintf(&b, "  sizes: %s\n", sanitizeAIText(member.Sizes))
		}
		if member.Preferences != "" {
			fmt.Fprintf(&b, "  preferences: %s\n", sanitizeAIText(member.Preferences))
		}
		if member.Interests != "" {
			fmt.Fprintf(&b, "  interests: %s\n", sanitizeAIText(member.Interests))
		}
	}
	if note := strings.TrimSpace(contextNote); note != "" {
		fmt.Fprintf(&b, "\nShopping context: %s\n", sanitizeAIText(note))
	}
	if budgetMin > 0 || budgetMax > 0 {
		fmt.Fprintf(&b, "\nBudget range: %.0f - %.0f\n", budgetMin, budgetMax)
	}
	return b.String()
}

func parseHouseholdShoppingJSON(raw string, memberCount int) (*householdShoppingPayload, error) {
	raw = extractJSONObject(raw)
	var parsed householdShoppingPayload
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, err
	}
	if strings.TrimSpace(parsed.Summary) == "" {
		return nil, fmt.Errorf("missing summary")
	}

	clean := make([]householdMemberSearchPayload, 0, len(parsed.Members))
	for _, item := range parsed.Members {
		name := strings.TrimSpace(item.MemberName)
		summary := strings.TrimSpace(item.Summary)
		if name == "" || summary == "" {
			continue
		}
		clean = append(clean, householdMemberSearchPayload{
			MemberName:  name,
			Summary:     summary,
			SearchQuery: strings.TrimSpace(item.SearchQuery),
		})
		if len(clean) >= memberCount {
			break
		}
	}

	if len(clean) == 0 {
		return nil, fmt.Errorf("no valid member picks")
	}

	parsed.Members = clean
	return &parsed, nil
}

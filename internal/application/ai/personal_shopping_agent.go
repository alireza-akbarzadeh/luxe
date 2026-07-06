package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

const TaskPersonalShoppingAgent = "personal_shopping_agent"

// PersonalShoppingAgent runs a memory-aware shopping conversation for signed-in users.
func (s *Service) PersonalShoppingAgent(
	ctx context.Context,
	userID uint,
	subjectKey string,
	memory ShoppingMemoryQueries,
	search SearchQueries,
	req dto.AiPersonalShoppingAgentRequest,
) (*dto.AiPersonalShoppingAgentResponse, error) {
	if userID == 0 {
		return nil, utils.ErrUnauthorized("sign in required")
	}
	if search == nil {
		return nil, utils.ErrInternal(fmt.Errorf("search not configured"))
	}

	memorySummary := ""
	tasteSignals := []string{}
	sources := []string{"Catalog search"}

	if memory != nil {
		memResp, memErr := s.ShoppingMemory(ctx, userID, subjectKey, memory, dto.AiShoppingMemoryRequest{Limit: 12})
		if memErr == nil && memResp != nil {
			memorySummary = strings.TrimSpace(memResp.Summary)
			for _, signal := range memResp.Signals {
				label := strings.TrimSpace(signal.Label)
				if label != "" {
					tasteSignals = append(tasteSignals, label)
				}
			}
			sources = append(sources, "Shopping memory")
		}
	}

	enriched := req.Messages
	if memorySummary != "" || req.Goal != "" {
		prefix := "Shopper context:\n"
		if memorySummary != "" {
			prefix += "Taste memory: " + memorySummary + "\n"
		}
		if req.Goal != "" {
			prefix += "Active goal: " + strings.TrimSpace(req.Goal) + "\n"
		}
		enriched = append([]dto.AiChatMessage{{Role: "user", Content: prefix}}, enriched...)
	}

	assistantResp, err := s.ShoppingAssistant(ctx, subjectKey, search, dto.AiShoppingAssistantRequest{
		Messages: enriched,
	})
	if err != nil {
		return nil, err
	}

	return &dto.AiPersonalShoppingAgentResponse{
		Reply:             assistantResp.Reply,
		MemorySummary:     memorySummary,
		TasteSignals:      tasteSignals,
		FollowUpQuestions: assistantResp.FollowUpQuestions,
		Recommendations:   assistantResp.Recommendations,
		Sources:           uniqueStrings(append(sources, assistantResp.Sources...)),
	}, nil
}

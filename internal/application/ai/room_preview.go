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
	TaskRoomPreview         = "room_preview"
	defaultRoomPreviewLimit = 4
)

type roomPreviewPayload struct {
	Summary       string   `json:"summary"`
	PlacementTips []string `json:"placement_tips"`
	ScaleAdvice   string   `json:"scale_advice"`
	HarmonyNotes  []string `json:"harmony_notes"`
	Warnings      []string `json:"warnings"`
	SearchQuery   string   `json:"search_query"`
}

// RoomPreview analyzes a room photo and suggests how a catalog product would fit the space.
func (s *Service) RoomPreview(
	ctx context.Context,
	subjectKey string,
	search SearchQueries,
	req dto.AiRoomPreviewRequest,
) (*dto.AiRoomPreviewResponse, error) {
	if !s.Enabled() {
		return nil, utils.NewAppError(http.StatusServiceUnavailable, "AI is not enabled", nil)
	}
	if search == nil {
		return nil, utils.ErrInternal(fmt.Errorf("search not configured"))
	}
	if !s.chatRL.allow("room_preview:" + subjectKey) {
		return nil, utils.ErrTooManyRequests()
	}

	roomImage, err := normalizeVisualImageDataURL(req.RoomImageBase64)
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
	if note := strings.TrimSpace(req.Context); note != "" {
		facts += "\nShopper note: " + sanitizeAIText(note) + "\n"
		sources = append(sources, "Shopper context")
	}

	system := `You are a luxury interior styling assistant previewing how a catalog product fits a shopper's room photo.
Respond with JSON only using this exact shape:
{"summary":"...","placement_tips":["..."],"scale_advice":"...","harmony_notes":["..."],"warnings":["..."],"search_query":"..."}
Rules:
- summary: 2-3 sentences on overall fit in this room.
- placement_tips: 2-4 practical placement ideas (wall, corner, lighting, spacing).
- scale_advice: one sentence on whether the product scale suits the room.
- harmony_notes: 2-3 notes on color/style harmony with existing decor.
- warnings: 0-2 issues (clutter, scale mismatch, lighting) — empty array if none.
- search_query: short keywords for ONE complementary decor item; empty if none needed.
Use ONLY the room photo and product facts below. Do not invent dimensions you cannot infer.
Plain text only — no markdown.

Product facts:
` + facts

	resp, err := s.provider.Complete(ctx, aiint.CompletionRequest{
		Messages: []aiint.Message{
			{Role: "system", Content: system},
			{
				Role:    "user",
				Content: "Preview how this product would look in my room.",
				Images:  []string{roomImage},
			},
		},
		MaxTokens:   900,
		Temperature: 0.25,
	})
	if err != nil {
		return nil, utils.NewAppError(http.StatusServiceUnavailable, "AI provider unavailable", err)
	}

	parsed, err := parseRoomPreviewJSON(resp.Content)
	if err != nil {
		return nil, utils.NewAppError(http.StatusServiceUnavailable, "AI returned invalid room preview", err)
	}

	response := &dto.AiRoomPreviewResponse{
		Summary:       parsed.Summary,
		PlacementTips: sanitizeStringList(parsed.PlacementTips),
		ScaleAdvice:   strings.TrimSpace(parsed.ScaleAdvice),
		HarmonyNotes:  sanitizeStringList(parsed.HarmonyNotes),
		Warnings:      sanitizeStringList(parsed.Warnings),
		Sources:       uniqueStrings(sources),
	}

	query := strings.TrimSpace(parsed.SearchQuery)
	if query == "" {
		return response, nil
	}

	inStock := true
	searchResult, err := search.GlobalSearch(ctx, dto.SearchRequest{
		Query:   query,
		Limit:   defaultRoomPreviewLimit + 2,
		Offset:  0,
		InStock: &inStock,
	})
	if err != nil {
		return response, nil
	}

	response.Recommendations = make([]dto.AiRecommendedProduct, 0, defaultRoomPreviewLimit)
	for _, item := range searchResult.Products {
		if item.ID == 0 || item.ID == req.ProductID {
			continue
		}
		response.Recommendations = append(response.Recommendations, dto.AiRecommendedProduct{
			Product: item,
			Reason:  "Complements this room layout.",
		})
		if len(response.Recommendations) >= defaultRoomPreviewLimit {
			break
		}
	}
	if len(response.Recommendations) > 0 {
		response.Sources = append(response.Sources, "Catalog search")
	}

	utils.Log.WithFields(map[string]any{
		"subject":    subjectKey,
		"task":       TaskRoomPreview,
		"product_id": req.ProductID,
	}).Info("ai room preview completed")

	return response, nil
}

func parseRoomPreviewJSON(raw string) (*roomPreviewPayload, error) {
	raw = extractJSONObject(raw)
	var parsed roomPreviewPayload
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, err
	}
	if strings.TrimSpace(parsed.Summary) == "" {
		return nil, fmt.Errorf("missing summary")
	}
	return &parsed, nil
}

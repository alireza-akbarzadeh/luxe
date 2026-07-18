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
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

const (
	TaskInteractiveViewer = "interactive_viewer"
	maxViewerHotspots     = 8
)

type interactiveViewerPayload struct {
	Summary  string                     `json:"summary"`
	Hotspots []interactiveViewerHotspot `json:"hotspots"`
}

type interactiveViewerHotspot struct {
	ID          string  `json:"id"`
	Label       string  `json:"label"`
	Description string  `json:"description"`
	XPercent    float64 `json:"x_percent"`
	YPercent    float64 `json:"y_percent"`
}

// InteractiveViewer analyzes a product photo and returns feature hotspots for PDP exploration.
func (s *Service) InteractiveViewer(
	ctx context.Context,
	subjectKey string,
	req dto.AiInteractiveViewerRequest,
) (*dto.AiInteractiveViewerResponse, error) {
	if !s.Enabled() {
		return nil, utils.NewAppError(http.StatusServiceUnavailable, "AI is not enabled", nil)
	}
	if !s.chatRL.allow("interactive_viewer:" + subjectKey) {
		return nil, utils.ErrTooManyRequests()
	}

	product, err := s.products.GetDetailedByID(ctx, req.ProductID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("product not found")
		}
		return nil, utils.ErrInternal(err)
	}

	if len(product.Images) == 0 {
		return nil, utils.ErrBadRequest("product has no images")
	}

	imageIndex := 0
	if req.ImageIndex != nil {
		imageIndex = *req.ImageIndex
	}
	if imageIndex < 0 || imageIndex >= len(product.Images) {
		return nil, utils.ErrBadRequest("invalid image index")
	}

	imageRef := strings.TrimSpace(product.Images[imageIndex])
	if imageRef == "" {
		return nil, utils.ErrBadRequest("product image unavailable")
	}

	facts, sources := productFacts(product)
	facts += viewerAttributeFacts(product)

	system := `You are a luxury product specialist creating an interactive photo viewer for shoppers.
Analyze the product image and identify 3-6 visible features worth highlighting as hotspots.
Respond with JSON only using this exact shape:
{"summary":"...","hotspots":[{"id":"feature-1","label":"...","description":"...","x_percent":50,"y_percent":40}]}
Rules:
- summary: one sentence inviting the shopper to explore the photo.
- hotspots: 3-6 items max; id unique slug (feature-1, dial, clasp, etc.).
- label: 2-5 words for the hotspot pin.
- description: 1-2 sentences about that visible detail using listing facts when possible.
- x_percent and y_percent: numbers 5-95 for hotspot center on THIS image (0=left/top, 100=right/bottom).
- Place hotspots on visible parts only; spread them across the product.
Use listing facts below; do not invent materials or specs not supported by facts.
Plain text only — no markdown.

Product facts:
` + facts

	resp, err := s.provider.Complete(ctx, aiint.CompletionRequest{
		Messages: []aiint.Message{
			{Role: "system", Content: system},
			{
				Role:    "user",
				Content: fmt.Sprintf("Create interactive hotspots for product image %d.", imageIndex+1),
				Images:  []string{imageRef},
			},
		},
		MaxTokens:   900,
		Temperature: 0.2,
	})
	if err != nil {
		return nil, utils.NewAppError(http.StatusServiceUnavailable, "AI provider unavailable", err)
	}

	parsed, err := parseInteractiveViewerJSON(resp.Content)
	if err != nil {
		return nil, utils.NewAppError(http.StatusServiceUnavailable, "AI returned invalid viewer data", err)
	}

	hotspots := make([]dto.AiProductViewerHotspot, 0, len(parsed.Hotspots))
	for i, spot := range parsed.Hotspots {
		if i >= maxViewerHotspots {
			break
		}
		id := strings.TrimSpace(spot.ID)
		if id == "" {
			id = fmt.Sprintf("feature-%d", i+1)
		}
		label := strings.TrimSpace(spot.Label)
		desc := strings.TrimSpace(spot.Description)
		if label == "" || desc == "" {
			continue
		}
		hotspots = append(hotspots, dto.AiProductViewerHotspot{
			ID:          sanitizeAIText(id),
			Label:       sanitizeAIText(label),
			Description: sanitizeAIText(desc),
			XPercent:    clampPercent(spot.XPercent),
			YPercent:    clampPercent(spot.YPercent),
		})
	}

	response := &dto.AiInteractiveViewerResponse{
		Summary:    sanitizeAIText(parsed.Summary),
		ImageIndex: imageIndex,
		Hotspots:   hotspots,
		Sources:    uniqueStrings(sources),
	}

	utils.Log.WithFields(map[string]any{
		"subject":     subjectKey,
		"task":        TaskInteractiveViewer,
		"product_id":  req.ProductID,
		"image_index": imageIndex,
		"hotspots":    len(hotspots),
	}).Info("ai interactive viewer completed")

	return response, nil
}

func viewerAttributeFacts(product *models.Product) string {
	if len(product.Attributes) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("\nStructured attributes:\n")
	for _, attr := range product.Attributes {
		if len(attr.Values) == 0 {
			continue
		}
		fmt.Fprintf(&b, "- %s: %s\n", sanitizeAIText(attr.Name), sanitizeAIText(strings.Join(attr.Values, ", ")))
	}
	return b.String()
}

func clampPercent(value float64) float64 {
	if value < 5 {
		return 5
	}
	if value > 95 {
		return 95
	}
	return value
}

func parseInteractiveViewerJSON(raw string) (*interactiveViewerPayload, error) {
	raw = extractJSONObject(raw)
	var parsed interactiveViewerPayload
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, err
	}
	if strings.TrimSpace(parsed.Summary) == "" {
		return nil, fmt.Errorf("missing summary")
	}
	return &parsed, nil
}

package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

const TaskBundleSuggest = "bundle_suggest"

// EnrichBundleCopy rewrites bundle titles and descriptions using grounded product facts.
func (s *Service) EnrichBundleCopy(
	ctx context.Context,
	subjectPrefix string,
	anchor models.Product,
	intent string,
	bundles []dto.SmartBundleItem,
) ([]dto.SmartBundleItem, error) {
	if !s.Enabled() || len(bundles) == 0 {
		return bundles, nil
	}

	subjectKey := "bundle:" + subjectPrefix + ":" + fmt.Sprint(anchor.ID)
	if !s.chatRL.allow(subjectKey) {
		return bundles, nil
	}

	facts := bundleFacts(anchor, intent, bundles)
	system := `You help shoppers discover compatible product bundles on a luxury e-commerce store.
Respond with JSON only using this exact shape:
{"bundles":[{"title":"...","description":"..."}]}
Rules:
- Return exactly one entry per bundle in the same order as the input list.
- title: short shopper-facing headline (max 8 words).
- description: one sentence on why these products work together for the given intent.
Use ONLY the product facts below. Do not invent specs, prices, or policies.

` + facts

	user := fmt.Sprintf("Rewrite copy for %d smart bundles with intent %q.", len(bundles), intent)
	resp, err := s.complete(ctx, system, user, 700)
	if err != nil {
		return bundles, err
	}

	copies, err := parseBundleCopyJSON(resp.Content)
	if err != nil || len(copies) == 0 {
		return bundles, nil
	}

	out := make([]dto.SmartBundleItem, len(bundles))
	copy(out, bundles)
	for i := range out {
		if i >= len(copies) {
			break
		}
		if strings.TrimSpace(copies[i].Title) != "" {
			out[i].Title = strings.TrimSpace(copies[i].Title)
		}
		if strings.TrimSpace(copies[i].Description) != "" {
			out[i].Description = strings.TrimSpace(copies[i].Description)
		}
	}

	utils.Log.WithFields(map[string]any{
		"anchor_id": anchor.ID,
		"intent":    intent,
		"task":      TaskBundleSuggest,
	}).Info("ai bundle copy enriched")

	return out, nil
}

func bundleFacts(anchor models.Product, intent string, bundles []dto.SmartBundleItem) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Intent: %s\nAnchor product: %s (category: %s, price: %.2f)\n",
		intent, anchor.Name, categoryName(anchor), anchor.Price)
	for i, bundle := range bundles {
		b.WriteString(fmt.Sprintf("\nBundle %d products:\n", i+1))
		for _, product := range bundle.Products {
			name := strings.TrimSpace(product.Name)
			if name == "" {
				continue
			}
			fmt.Fprintf(&b, "- %s (%.2f)\n", name, product.Price)
		}
	}
	return b.String()
}

func categoryName(product models.Product) string {
	if product.Category != nil && product.Category.Name != "" {
		return product.Category.Name
	}
	return "uncategorized"
}

type bundleCopyItem struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type bundleCopyPayload struct {
	Bundles []bundleCopyItem `json:"bundles"`
}

func parseBundleCopyJSON(raw string) ([]bundleCopyItem, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var payload bundleCopyPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, err
	}
	return payload.Bundles, nil
}

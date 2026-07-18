package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

const (
	TaskProductConfigurator   = "product_configurator"
	defaultConfiguratorAddOns = 3
)

type productConfiguratorPayload struct {
	Summary     string                         `json:"summary"`
	Selections  []productConfiguratorSelection `json:"selections"`
	Tips        []string                       `json:"tips"`
	SearchQuery string                         `json:"search_query"`
}

type productConfiguratorSelection struct {
	Attribute string `json:"attribute"`
	Value     string `json:"value"`
	Reason    string `json:"reason"`
}

// ProductConfigurator recommends variant selections and optional complementary add-ons.
func (s *Service) ProductConfigurator(
	ctx context.Context,
	subjectKey string,
	search SearchQueries,
	req dto.AiProductConfiguratorRequest,
) (*dto.AiProductConfiguratorResponse, error) {
	if !s.Enabled() {
		return nil, utils.NewAppError(http.StatusServiceUnavailable, "AI is not enabled", nil)
	}
	if search == nil {
		return nil, utils.ErrInternal(fmt.Errorf("search not configured"))
	}
	if !s.chatRL.allow("product_configurator:" + subjectKey) {
		return nil, utils.ErrTooManyRequests()
	}

	product, err := s.products.GetDetailedByID(ctx, req.ProductID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("product not found")
		}
		return nil, utils.ErrInternal(err)
	}

	optionMap := configuratorOptionMap(product)
	if len(optionMap) == 0 {
		return nil, utils.ErrBadRequest("product has no configurable options")
	}

	facts, sources := productFacts(product)
	facts += configuratorOptionsFacts(optionMap)

	userNote := strings.TrimSpace(req.Context)
	if userNote != "" {
		facts += "\nShopper context: " + sanitizeAIText(userNote) + "\n"
		sources = append(sources, "Shopper context")
	}
	if len(req.Preferences) > 0 {
		facts += "\nCurrent selections:\n"
		for key, value := range req.Preferences {
			key = strings.TrimSpace(key)
			value = strings.TrimSpace(value)
			if key == "" || value == "" {
				continue
			}
			facts += fmt.Sprintf("- %s: %s\n", sanitizeAIText(key), sanitizeAIText(value))
		}
		sources = append(sources, "Current selections")
	}

	system := `You are a luxury product configurator helping a shopper choose variant options.
Respond with JSON only using this exact shape:
{"summary":"...","selections":[{"attribute":"color","value":"Navy","reason":"..."}],"tips":["..."],"search_query":"..."}
Rules:
- summary: 1-2 sentences explaining the recommended build.
- selections: one entry per configurable attribute listed below; value MUST be from that attribute's allowed values.
- reason: short phrase why that value fits the shopper context.
- tips: 1-3 care or styling tips for this configuration.
- search_query: keywords for ONE complementary accessory; empty if none needed.
Use ONLY allowed option values from the listing. Do not invent new colors or sizes.
Plain text only — no markdown.

Product facts:
` + facts

	resp, err := s.complete(ctx, system, "Recommend the best configuration for this shopper.", 900)
	if err != nil {
		return nil, utils.NewAppError(http.StatusServiceUnavailable, "AI provider unavailable", err)
	}

	parsed, err := parseProductConfiguratorJSON(resp.Content)
	if err != nil {
		return nil, utils.NewAppError(http.StatusServiceUnavailable, "AI returned invalid configuration", err)
	}

	selections := sanitizeConfiguratorSelections(parsed.Selections, optionMap)
	response := &dto.AiProductConfiguratorResponse{
		Summary:    sanitizeAIText(parsed.Summary),
		Selections: selections,
		Tips:       sanitizeStringList(parsed.Tips),
		Sources:    uniqueStrings(sources),
	}

	query := strings.TrimSpace(parsed.SearchQuery)
	if query != "" {
		inStock := true
		searchResult, err := search.GlobalSearch(ctx, dto.SearchRequest{
			Query:   query,
			Limit:   defaultConfiguratorAddOns + 2,
			Offset:  0,
			InStock: &inStock,
		})
		if err == nil {
			response.AddOns = make([]dto.AiRecommendedProduct, 0, defaultConfiguratorAddOns)
			for _, item := range searchResult.Products {
				if item.ID == 0 || item.ID == req.ProductID {
					continue
				}
				response.AddOns = append(response.AddOns, dto.AiRecommendedProduct{
					Product: item,
					Reason:  "Pairs with this configuration.",
				})
				if len(response.AddOns) >= defaultConfiguratorAddOns {
					break
				}
			}
			if len(response.AddOns) > 0 {
				response.Sources = append(response.Sources, "Catalog search")
			}
		}
	}

	utils.Log.WithFields(map[string]any{
		"subject":    subjectKey,
		"task":       TaskProductConfigurator,
		"product_id": req.ProductID,
		"selections": len(selections),
	}).Info("ai product configurator completed")

	return response, nil
}

func configuratorOptionMap(product *models.Product) map[string][]string {
	options := make(map[string][]string)

	appendOption := func(name string, values []string) {
		name = strings.TrimSpace(strings.ToLower(name))
		if name == "" || len(values) == 0 {
			return
		}
		clean := make([]string, 0, len(values))
		seen := make(map[string]struct{}, len(values))
		for _, value := range values {
			value = strings.TrimSpace(value)
			if value == "" {
				continue
			}
			key := strings.ToLower(value)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			clean = append(clean, value)
		}
		if len(clean) == 0 {
			return
		}
		if existing, ok := options[name]; ok {
			options[name] = mergeUniqueValues(existing, clean)
			return
		}
		options[name] = clean
	}

	for _, attr := range product.Attributes {
		if len(attr.Values) <= 1 {
			continue
		}
		appendOption(attr.Name, attr.Values)
	}
	if len(product.Colors) > 0 {
		appendOption("color", product.Colors)
	}
	if len(product.Sizes) > 0 {
		appendOption("size", product.Sizes)
	}

	return options
}

func mergeUniqueValues(base, extra []string) []string {
	seen := make(map[string]struct{}, len(base)+len(extra))
	out := make([]string, 0, len(base)+len(extra))
	for _, value := range append(base, extra...) {
		key := strings.ToLower(strings.TrimSpace(value))
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, strings.TrimSpace(value))
	}
	return out
}

func configuratorOptionsFacts(optionMap map[string][]string) string {
	if len(optionMap) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("\nConfigurable options (use exact values):\n")
	for name, values := range optionMap {
		fmt.Fprintf(&b, "- %s: %s\n", sanitizeAIText(name), sanitizeAIText(strings.Join(values, ", ")))
	}
	return b.String()
}

func sanitizeConfiguratorSelections(
	raw []productConfiguratorSelection,
	optionMap map[string][]string,
) []dto.AiProductConfiguratorSelection {
	out := make([]dto.AiProductConfiguratorSelection, 0, len(raw))
	seenAttrs := make(map[string]struct{}, len(raw))

	for _, item := range raw {
		attrKey := strings.TrimSpace(strings.ToLower(item.Attribute))
		if attrKey == "" {
			continue
		}
		if _, dup := seenAttrs[attrKey]; dup {
			continue
		}

		allowed, ok := optionMap[attrKey]
		if !ok {
			for name, values := range optionMap {
				if name == attrKey || strings.TrimSpace(strings.ToLower(name)) == attrKey {
					allowed = values
					attrKey = name
					ok = true
					break
				}
			}
		}
		if !ok {
			continue
		}

		value := matchAllowedValue(item.Value, allowed)
		if value == "" {
			continue
		}

		seenAttrs[attrKey] = struct{}{}
		out = append(out, dto.AiProductConfiguratorSelection{
			Attribute: attrKey,
			Value:     value,
			Reason:    strings.TrimSpace(sanitizeAIText(item.Reason)),
		})
	}

	return out
}

func matchAllowedValue(candidate string, allowed []string) string {
	candidate = strings.TrimSpace(candidate)
	if candidate == "" {
		return ""
	}
	lower := strings.ToLower(candidate)
	for _, value := range allowed {
		if strings.EqualFold(value, candidate) {
			return value
		}
		if strings.Contains(strings.ToLower(value), lower) || strings.Contains(lower, strings.ToLower(value)) {
			return value
		}
	}
	return ""
}

func parseProductConfiguratorJSON(raw string) (*productConfiguratorPayload, error) {
	raw = extractJSONObject(raw)
	var parsed productConfiguratorPayload
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, err
	}
	if strings.TrimSpace(parsed.Summary) == "" {
		return nil, fmt.Errorf("missing summary")
	}
	return &parsed, nil
}

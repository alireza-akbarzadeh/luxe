package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

const TaskCompatibilityCheck = "compatibility_check"

// CompatibilityCheck scores how well two catalog products work together.
func (s *Service) CompatibilityCheck(
	ctx context.Context,
	subjectKey string,
	req dto.AiCompatibilityCheckRequest,
) (*dto.AiCompatibilityCheckResponse, error) {
	if !s.Enabled() {
		return nil, utils.NewAppError(503, "AI is not enabled", nil)
	}
	if !s.chatRL.allow("compatibility:" + subjectKey) {
		return nil, utils.ErrTooManyRequests()
	}
	if req.ProductIDA == req.ProductIDB {
		return nil, utils.ErrBadRequest("choose two different products")
	}

	productA, err := s.products.GetDetailedByID(ctx, req.ProductIDA)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("product A not found")
		}
		return nil, utils.ErrInternal(err)
	}
	productB, err := s.products.GetDetailedByID(ctx, req.ProductIDB)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("product B not found")
		}
		return nil, utils.ErrInternal(err)
	}

	score := scoreProductPair(productA, productB)
	compatibility := compatibilityLabel(score)
	facts := pairCompatibilityFacts(productA, productB, score)
	sources := []string{"Catalog attributes", "Category & store signals"}

	system := `You are a universal product compatibility engine for a luxury marketplace.
Explain how well two products work together for a shopper.
Respond with JSON only:
{"summary":"...","works_well":["..."],"concerns":["..."],"category":"pairing type"}
Rules:
- summary: 2-3 sentences on the pairing.
- works_well: 2-4 bullets when score is decent; can be shorter when poor fit.
- concerns: 0-3 honest caveats (style clash, duplicate function, different stores).
- category: short label like "Desk pairing", "Travel set", "Style clash".
Use ONLY the facts below.

Facts:
` + facts

	user := "How compatible are these two products?"
	resp, err := s.complete(ctx, system, user, 900)
	if err != nil {
		return ruleBasedCompatibility(productA, productB, score, compatibility, sources), nil
	}

	parsed, parseErr := parseCompatibilityJSON(resp.Content)
	if parseErr != nil {
		return ruleBasedCompatibility(productA, productB, score, compatibility, sources), nil
	}

	result := &dto.AiCompatibilityCheckResponse{
		Score:         score,
		Summary:       parsed.Summary,
		WorksWell:     parsed.WorksWell,
		Concerns:      parsed.Concerns,
		Category:      parsed.Category,
		Compatibility: compatibility,
		Sources:       uniqueStrings(sources),
	}

	utils.Log.WithFields(map[string]any{
		"product_a": req.ProductIDA,
		"product_b": req.ProductIDB,
		"subject":   subjectKey,
		"task":      TaskCompatibilityCheck,
		"score":     score,
	}).Info("ai compatibility check completed")

	return result, nil
}

func scoreProductPair(a, b *models.Product) int {
	score := 50
	if a.StoreID == b.StoreID && a.StoreID > 0 {
		score += 15
	}
	if a.CategoryID != nil && b.CategoryID != nil && *a.CategoryID == *b.CategoryID {
		score += 10
	}
	brandA := brandName(a)
	brandB := brandName(b)
	if brandA != "" && brandA == brandB {
		score += 8
	}
	if a.Rating >= 4 && b.Rating >= 4 {
		score += 7
	}
	if a.IsDigital != b.IsDigital {
		score -= 10
	}
	if score > 98 {
		return 98
	}
	if score < 12 {
		return 12
	}
	return score
}

func compatibilityLabel(score int) string {
	switch {
	case score >= 80:
		return "excellent"
	case score >= 65:
		return "good"
	case score >= 45:
		return "mixed"
	default:
		return "poor"
	}
}

func pairCompatibilityFacts(a, b *models.Product, score int) string {
	catA := categoryName(a)
	catB := categoryName(b)
	return fmt.Sprintf(
		"Product A: %s | category: %s | brand: %s | price: %.2f | store_id: %d\nProduct B: %s | category: %s | brand: %s | price: %.2f | store_id: %d\nHeuristic score: %d/100\n",
		a.Name, catA, brandName(a), a.Price, a.StoreID,
		b.Name, catB, brandName(b), b.Price, b.StoreID,
		score,
	)
}

func brandName(product *models.Product) string {
	if product.Brand != nil {
		return strings.TrimSpace(product.Brand.Name)
	}
	return ""
}

func categoryName(product *models.Product) string {
	if product.Category != nil {
		return product.Category.Name
	}
	return ""
}

func ruleBasedCompatibility(a, b *models.Product, score int, compatibility string, sources []string) *dto.AiCompatibilityCheckResponse {
	summary := fmt.Sprintf(
		"%s and %s score %d/100 for compatibility based on category, store, and quality signals.",
		a.Name,
		b.Name,
		score,
	)
	works := []string{}
	concerns := []string{}
	if a.StoreID == b.StoreID && a.StoreID > 0 {
		works = append(works, "Same seller — easier combined shipping and returns.")
	}
	if score < 45 {
		concerns = append(concerns, "These items may not complement each other — double-check your use case.")
	}

	return &dto.AiCompatibilityCheckResponse{
		Score:         score,
		Summary:       summary,
		WorksWell:     works,
		Concerns:      concerns,
		Category:      "Catalog pairing",
		Compatibility: compatibility,
		Sources:       uniqueStrings(sources),
	}
}

func parseCompatibilityJSON(content string) (*dto.AiCompatibilityCheckResponse, error) {
	trimmed := strings.TrimSpace(content)
	if idx := strings.Index(trimmed, "{"); idx >= 0 {
		if end := strings.LastIndex(trimmed, "}"); end > idx {
			trimmed = trimmed[idx : end+1]
		}
	}

	var raw struct {
		Summary   string   `json:"summary"`
		WorksWell []string `json:"works_well"`
		Concerns  []string `json:"concerns"`
		Category  string   `json:"category"`
	}
	if err := json.Unmarshal([]byte(trimmed), &raw); err != nil {
		return nil, err
	}

	summary := sanitizeAIText(raw.Summary)
	if summary == "" {
		return nil, fmt.Errorf("empty summary")
	}

	return &dto.AiCompatibilityCheckResponse{
		Summary:   summary,
		WorksWell: sanitizeStringList(raw.WorksWell),
		Concerns:  sanitizeStringList(raw.Concerns),
		Category:  sanitizeAIText(raw.Category),
	}, nil
}

package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

const TaskNegotiation = "negotiation"

// Negotiation evaluates a shopper offer against listing price and store norms.
func (s *Service) Negotiation(
	ctx context.Context,
	subjectKey string,
	req dto.AiNegotiationRequest,
) (*dto.AiNegotiationResponse, error) {
	if !s.Enabled() {
		return nil, utils.NewAppError(503, "AI is not enabled", nil)
	}
	if !s.chatRL.allow("negotiation:" + subjectKey) {
		return nil, utils.ErrTooManyRequests()
	}

	product, err := s.products.GetDetailedByID(ctx, req.ProductID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("product not found")
		}
		return nil, utils.ErrInternal(err)
	}

	listPrice := product.Price
	offer := req.OfferedPrice
	discountPct := 0.0
	if listPrice > 0 {
		discountPct = math.Round((1 - offer/listPrice) * 100)
	}

	facts, sources := productFacts(product)
	offerFacts := fmt.Sprintf(
		"\nList price: %.2f\nShopper offer: %.2f\nRequested discount: %.0f%%\nShopper message: %s\n",
		listPrice,
		offer,
		discountPct,
		strings.TrimSpace(req.Message),
	)

	system := `You are a fair marketplace negotiation assistant helping shoppers make reasonable offers.
Respond with JSON only:
{"verdict":"accept|counter|decline","counter_price":0,"confidence":"high|medium|low","summary":"...","tips":["..."]}
Rules:
- accept: offer is within ~5% of list or clearly fair for the product condition and category.
- counter: offer is low but reasonable — suggest a counter_price between offer and list.
- decline: offer is far below value; counter_price may be omitted.
- summary: 2 sentences explaining the recommendation in plain language.
- tips: 2-3 negotiation tips for the shopper (polite message, bundle idea, timing).
Use ONLY the facts below. Do not promise the seller will accept.

Facts:
` + facts + offerFacts

	user := "Evaluate this offer and guide the shopper."
	resp, err := s.complete(ctx, system, user, 900)
	if err != nil {
		return ruleBasedNegotiation(product, offer, discountPct, sources)
	}

	parsed, parseErr := parseNegotiationJSON(resp.Content)
	if parseErr != nil {
		return ruleBasedNegotiation(product, offer, discountPct, sources)
	}

	result := &dto.AiNegotiationResponse{
		Verdict:    parsed.Verdict,
		Summary:    parsed.Summary,
		Confidence: parsed.Confidence,
		Tips:       parsed.Tips,
		Sources:    uniqueStrings(sources),
	}
	if parsed.CounterPrice != nil && *parsed.CounterPrice > 0 {
		cp := math.Round(*parsed.CounterPrice*100) / 100
		result.CounterPrice = &cp
	}

	utils.Log.WithFields(map[string]any{
		"product_id": req.ProductID,
		"subject":    subjectKey,
		"task":       TaskNegotiation,
		"verdict":    result.Verdict,
	}).Info("ai negotiation completed")

	return result, nil
}

func ruleBasedNegotiation(product *models.Product, offer, discountPct float64, sources []string) (*dto.AiNegotiationResponse, error) {
	listPrice := product.Price
	verdict := "counter"
	confidence := "medium"
	var counter *float64

	switch {
	case discountPct <= 5:
		verdict = "accept"
		confidence = "high"
	case discountPct >= 35:
		verdict = "decline"
		confidence = "high"
	default:
		mid := math.Round((offer+listPrice)/2*100) / 100
		counter = &mid
	}

	summary := fmt.Sprintf(
		"Your offer of %.2f is %.0f%% below the list price of %.2f. ",
		offer,
		discountPct,
		listPrice,
	)
	switch verdict {
	case "accept":
		summary += "This is close to the listed price and may be worth submitting to the seller."
	case "decline":
		summary += "The gap is large — consider raising your offer or watching for a promotion."
	default:
		summary += "A counter around the midpoint could be a reasonable next step."
	}

	return &dto.AiNegotiationResponse{
		Verdict:      verdict,
		CounterPrice: counter,
		Summary:      summary,
		Confidence:   confidence,
		Tips: []string{
			"Mention why you love the item and your timeline to buy.",
			"Ask about bundle discounts if you are buying multiple pieces.",
		},
		Sources: uniqueStrings(sources),
	}, nil
}

func parseNegotiationJSON(content string) (*dto.AiNegotiationResponse, error) {
	trimmed := strings.TrimSpace(content)
	if idx := strings.Index(trimmed, "{"); idx >= 0 {
		if end := strings.LastIndex(trimmed, "}"); end > idx {
			trimmed = trimmed[idx : end+1]
		}
	}

	var raw struct {
		Verdict      string   `json:"verdict"`
		CounterPrice *float64 `json:"counter_price"`
		Confidence   string   `json:"confidence"`
		Summary      string   `json:"summary"`
		Tips         []string `json:"tips"`
	}
	if err := json.Unmarshal([]byte(trimmed), &raw); err != nil {
		return nil, err
	}

	summary := sanitizeAIText(raw.Summary)
	if summary == "" {
		return nil, fmt.Errorf("empty summary")
	}

	verdict := strings.ToLower(strings.TrimSpace(raw.Verdict))
	if verdict != "accept" && verdict != "counter" && verdict != "decline" {
		verdict = "counter"
	}

	return &dto.AiNegotiationResponse{
		Verdict:      verdict,
		CounterPrice: raw.CounterPrice,
		Confidence:   normalizeConfidence(raw.Confidence),
		Summary:      summary,
		Tips:         sanitizeStringList(raw.Tips),
	}, nil
}

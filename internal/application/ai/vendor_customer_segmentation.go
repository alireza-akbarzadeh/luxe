package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	apporder "github.com/alireza-akbarzadeh/luxe/internal/application/order"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

const TaskVendorCustomerSegments = "vendor_customer_segments"

// VendorCustomerSnapshot aggregates buyer behavior for segmentation.
type VendorCustomerSnapshot struct {
	StoreName  string
	PeriodDays int
	Customers  []apporder.VendorStoreCustomer
}

// VendorCustomerSegmentationQueries loads store customer facts.
type VendorCustomerSegmentationQueries interface {
	LoadCustomerSnapshot(ctx context.Context, storeID uint, days int) (*VendorCustomerSnapshot, error)
}

type vendorCustomerSegmentsPayload struct {
	Summary         string   `json:"summary"`
	Highlights      []string `json:"highlights"`
	Recommendations []string `json:"recommendations"`
	CampaignIdeas   []string `json:"campaign_ideas"`
}

// VendorCustomerSegments returns AI-generated or rule-based customer segments for sellers.
func (s *Service) VendorCustomerSegments(
	ctx context.Context,
	storeID uint,
	days int,
	subjectKey string,
	queries VendorCustomerSegmentationQueries,
) (*dto.AiVendorCustomerSegmentsResponse, error) {
	if queries == nil {
		return nil, utils.ErrInternal(fmt.Errorf("vendor customer segmentation queries not configured"))
	}
	if storeID == 0 {
		return nil, utils.ErrBadRequest("store id required")
	}
	if days <= 0 {
		days = 365
	}
	if days > 730 {
		days = 730
	}

	snapshot, err := queries.LoadCustomerSnapshot(ctx, storeID, days)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	members, summaries := buildCustomerSegments(snapshot.Customers, now)
	response := &dto.AiVendorCustomerSegmentsResponse{
		PeriodDays: days,
		Segments:   summaries,
		Customers:  members,
		Sources:    []string{"orders", "users"},
	}

	if !s.Enabled() {
		fallback := fallbackVendorCustomerSegments(snapshot, summaries)
		response.AiEnabled = false
		response.Summary = fallback.Summary
		response.Highlights = fallback.Highlights
		response.Recommendations = fallback.Recommendations
		response.CampaignIdeas = fallback.CampaignIdeas
		return response, nil
	}

	if !s.adminRL.allow("vendor_customer_segments:" + subjectKey) {
		return nil, utils.ErrTooManyRequests()
	}

	facts := vendorCustomerFacts(snapshot, summaries)
	system := `You are a concise customer marketing strategist for marketplace sellers.
Respond with JSON only using this exact shape:
{"summary":"...","highlights":["..."],"recommendations":["..."],"campaign_ideas":["..."]}
Rules:
- summary: 2-3 sentences on customer base health and segment mix.
- highlights: 2-4 notable patterns in spend, repeat rate, or churn risk.
- recommendations: 2-3 retention or upsell actions tied to segments.
- campaign_ideas: 2-3 targeted campaign concepts (email/promo) for specific segments.
Use ONLY the customer facts below. Do not invent ad spend or email open rates.
Plain text only — no markdown.

Customer facts:
` + facts

	user := fmt.Sprintf("Analyze customer segments from the last %d days of store orders.", days)
	resp, err := s.complete(ctx, system, user, 900)
	if err != nil {
		fallback := fallbackVendorCustomerSegments(snapshot, summaries)
		response.AiEnabled = false
		response.Summary = "AI is temporarily unavailable. Showing rule-based customer segments from your order history."
		response.Highlights = fallback.Highlights
		response.Recommendations = fallback.Recommendations
		response.CampaignIdeas = fallback.CampaignIdeas
		return response, nil
	}

	parsed, err := parseVendorCustomerSegmentsJSON(resp.Content)
	if err != nil {
		fallback := fallbackVendorCustomerSegments(snapshot, summaries)
		response.AiEnabled = false
		response.Summary = fallback.Summary
		response.Highlights = fallback.Highlights
		response.Recommendations = fallback.Recommendations
		response.CampaignIdeas = fallback.CampaignIdeas
		return response, nil
	}

	response.AiEnabled = true
	response.Summary = strings.TrimSpace(parsed.Summary)
	response.Highlights = sanitizeStringList(parsed.Highlights)
	response.Recommendations = sanitizeStringList(parsed.Recommendations)
	response.CampaignIdeas = sanitizeStringList(parsed.CampaignIdeas)
	return response, nil
}

func assignCustomerSegment(customer apporder.VendorStoreCustomer, now time.Time) string {
	daysSinceLast := now.Sub(customer.LastOrderAt).Hours() / 24

	if customer.TotalSpend >= 500 || customer.OrderCount >= 5 {
		if daysSinceLast > 90 {
			return "at_risk"
		}
		return "vip"
	}
	if customer.OrderCount >= 3 {
		if daysSinceLast > 90 {
			return "at_risk"
		}
		return "loyal"
	}
	if customer.OrderCount == 2 {
		if daysSinceLast > 90 {
			return "at_risk"
		}
		return "repeat"
	}
	if customer.OrderCount == 1 {
		if daysSinceLast <= 30 {
			return "new"
		}
		return "one_time"
	}
	return "one_time"
}

func segmentLabel(segment string) string {
	switch segment {
	case "vip":
		return "VIP"
	case "loyal":
		return "Loyal repeat"
	case "repeat":
		return "Repeat buyer"
	case "new":
		return "New customer"
	case "at_risk":
		return "At risk"
	default:
		return "One-time"
	}
}

func buildCustomerSegments(
	customers []apporder.VendorStoreCustomer,
	now time.Time,
) ([]dto.VendorCustomerSegmentMember, []dto.VendorCustomerSegmentSummary) {
	summaryMap := map[string]*dto.VendorCustomerSegmentSummary{
		"vip":      {Segment: "vip", Label: segmentLabel("vip")},
		"loyal":    {Segment: "loyal", Label: segmentLabel("loyal")},
		"repeat":   {Segment: "repeat", Label: segmentLabel("repeat")},
		"new":      {Segment: "new", Label: segmentLabel("new")},
		"at_risk":  {Segment: "at_risk", Label: segmentLabel("at_risk")},
		"one_time": {Segment: "one_time", Label: segmentLabel("one_time")},
	}

	members := make([]dto.VendorCustomerSegmentMember, 0, len(customers))
	for _, customer := range customers {
		segment := assignCustomerSegment(customer, now)
		avgOrder := 0.0
		if customer.OrderCount > 0 {
			avgOrder = customer.TotalSpend / float64(customer.OrderCount)
		}

		members = append(members, dto.VendorCustomerSegmentMember{
			UserID:        customer.UserID,
			Name:          strings.TrimSpace(customer.FirstName + " " + customer.LastName),
			Email:         customer.Email,
			OrderCount:    customer.OrderCount,
			TotalSpend:    customer.TotalSpend,
			AvgOrderValue: avgOrder,
			LastOrderAt:   customer.LastOrderAt.Format(time.RFC3339),
			Segment:       segment,
		})

		if summary, ok := summaryMap[segment]; ok {
			summary.Count++
			summary.TotalSpend += customer.TotalSpend
		}
	}

	summaries := make([]dto.VendorCustomerSegmentSummary, 0, len(summaryMap))
	order := []string{"vip", "loyal", "repeat", "new", "at_risk", "one_time"}
	for _, key := range order {
		if summary := summaryMap[key]; summary.Count > 0 {
			summaries = append(summaries, *summary)
		}
	}
	return members, summaries
}

func vendorCustomerFacts(snapshot *VendorCustomerSnapshot, summaries []dto.VendorCustomerSegmentSummary) string {
	if snapshot == nil {
		return "No customer data available."
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Store: %s\n", snapshot.StoreName)
	fmt.Fprintf(&b, "Analysis window: %d days\n", snapshot.PeriodDays)
	fmt.Fprintf(&b, "Unique customers: %d\n", len(snapshot.Customers))

	for _, summary := range summaries {
		fmt.Fprintf(&b, "Segment %s (%s): %d customers, $%.2f spend\n",
			summary.Segment, summary.Label, summary.Count, summary.TotalSpend)
	}

	return b.String()
}

func fallbackVendorCustomerSegments(
	snapshot *VendorCustomerSnapshot,
	summaries []dto.VendorCustomerSegmentSummary,
) *vendorCustomerSegmentsPayload {
	highlights := make([]string, 0, 4)
	recommendations := make([]string, 0, 3)
	campaignIdeas := make([]string, 0, 3)

	for _, summary := range summaries {
		if summary.Segment == "vip" && summary.Count > 0 {
			highlights = append(highlights, fmt.Sprintf("%d VIP customers drove $%.2f in spend", summary.Count, summary.TotalSpend))
		}
		if summary.Segment == "at_risk" && summary.Count > 0 {
			highlights = append(highlights, fmt.Sprintf("%d customers have not ordered recently", summary.Count))
			recommendations = append(recommendations, "Send a win-back offer to at-risk customers with a limited-time discount")
			campaignIdeas = append(campaignIdeas, "We miss you — 15% off your next order")
		}
	}

	for _, summary := range summaries {
		if summary.Segment == "new" && summary.Count > 0 {
			campaignIdeas = append(campaignIdeas, "Welcome series with complementary product recommendations")
		}
		if summary.Segment == "one_time" && summary.Count > 0 {
			recommendations = append(recommendations, fmt.Sprintf("Convert %d one-time buyers with a second-purchase incentive", summary.Count))
		}
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Reward loyal repeat buyers with early access to new arrivals")
	}
	if len(campaignIdeas) == 0 {
		campaignIdeas = append(campaignIdeas, "Segment-specific email with best sellers for each buyer tier")
	}

	summary := fmt.Sprintf(
		"%s identified %d unique customers across %d days of orders.",
		snapshot.StoreName,
		len(snapshot.Customers),
		snapshot.PeriodDays,
	)

	return &vendorCustomerSegmentsPayload{
		Summary:         summary,
		Highlights:      highlights,
		Recommendations: recommendations,
		CampaignIdeas:   campaignIdeas,
	}
}

func parseVendorCustomerSegmentsJSON(content string) (*vendorCustomerSegmentsPayload, error) {
	content = extractJSONObject(content)
	var payload vendorCustomerSegmentsPayload
	if err := json.Unmarshal([]byte(content), &payload); err != nil {
		return nil, err
	}
	if strings.TrimSpace(payload.Summary) == "" {
		return nil, fmt.Errorf("missing summary")
	}
	return &payload, nil
}

package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

const TaskVendorDashboard = "vendor_dashboard"

// VendorDashboardOrderLine is a recent order row for vendor AI context.
type VendorDashboardOrderLine struct {
	OrderNumber string
	Status      string
	Subtotal    float64
	ItemCount   int
}

// VendorDashboardStockLine is a low-stock product row for vendor AI context.
type VendorDashboardStockLine struct {
	Name  string
	Stock int
}

// VendorDashboardSnapshot aggregates store facts for AI or rule-based insights.
type VendorDashboardSnapshot struct {
	StoreName        string
	OrderTotal       int64
	OrdersByStatus   map[string]int64
	ProductTotal     int64
	ProductsByStatus map[string]int64
	LowStockCount    int64
	RecentOrders     []VendorDashboardOrderLine
	LowStockItems    []VendorDashboardStockLine
}

// VendorDashboardQueries loads store operational facts for the vendor dashboard.
type VendorDashboardQueries interface {
	LoadSnapshot(ctx context.Context, storeID uint) (*VendorDashboardSnapshot, error)
}

type vendorDashboardPayload struct {
	Summary       string   `json:"summary"`
	HealthScore   int      `json:"health_score"`
	Priorities    []string `json:"priorities"`
	Opportunities []string `json:"opportunities"`
	Alerts        []string `json:"alerts"`
}

// VendorDashboard returns AI-generated or rule-based store insights for sellers.
func (s *Service) VendorDashboard(
	ctx context.Context,
	storeID uint,
	subjectKey string,
	queries VendorDashboardQueries,
) (*dto.AiVendorDashboardResponse, error) {
	if queries == nil {
		return nil, utils.ErrInternal(fmt.Errorf("vendor dashboard queries not configured"))
	}
	if storeID == 0 {
		return nil, utils.ErrBadRequest("store id required")
	}

	snapshot, err := queries.LoadSnapshot(ctx, storeID)
	if err != nil {
		return nil, err
	}

	if !s.Enabled() {
		resp := fallbackVendorDashboard(snapshot)
		resp.AiEnabled = false
		return resp, nil
	}

	if !s.adminRL.allow("vendor_dashboard:" + subjectKey) {
		return nil, utils.ErrTooManyRequests()
	}

	facts := vendorDashboardFacts(snapshot)
	system := `You are a concise e-commerce operations advisor for marketplace sellers.
Respond with JSON only using this exact shape:
{"summary":"...","health_score":0,"priorities":["..."],"opportunities":["..."],"alerts":["..."]}
Rules:
- summary: 2-3 sentences on overall store health and momentum.
- health_score: integer 0-100 reflecting fulfillment, catalog, and inventory health.
- priorities: 2-4 urgent actions the seller should take today (fulfillment, stock, catalog gaps).
- opportunities: 1-3 growth ideas (merchandising, pricing, campaigns) grounded in the facts.
- alerts: 0-3 risk flags (stockouts, backlog, inactive SKUs) — empty array if none.
Use ONLY the store facts below. Do not invent revenue, traffic, or review metrics.
Plain text only — no markdown.

Store facts:
` + facts

	user := "Generate today's vendor dashboard briefing."
	resp, err := s.complete(ctx, system, user, 900)
	if err != nil {
		fallback := fallbackVendorDashboard(snapshot)
		fallback.AiEnabled = false
		fallback.Summary = "AI is temporarily unavailable. Showing rule-based insights from your store data."
		return fallback, nil
	}

	parsed, err := parseVendorDashboardJSON(resp.Content)
	if err != nil {
		fallback := fallbackVendorDashboard(snapshot)
		fallback.AiEnabled = false
		return fallback, nil
	}

	health := parsed.HealthScore
	if health < 0 {
		health = 0
	}
	if health > 100 {
		health = 100
	}

	return &dto.AiVendorDashboardResponse{
		AiEnabled:     true,
		Summary:       strings.TrimSpace(parsed.Summary),
		HealthScore:   health,
		Priorities:    sanitizeStringList(parsed.Priorities),
		Opportunities: sanitizeStringList(parsed.Opportunities),
		Alerts:        sanitizeStringList(parsed.Alerts),
		Sources:       []string{"orders", "products", "inventory"},
	}, nil
}

func vendorDashboardFacts(snapshot *VendorDashboardSnapshot) string {
	if snapshot == nil {
		return "No store data available."
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Store: %s\n", snapshot.StoreName)
	fmt.Fprintf(&b, "Orders total: %d\n", snapshot.OrderTotal)
	for status, count := range snapshot.OrdersByStatus {
		fmt.Fprintf(&b, "Orders %s: %d\n", status, count)
	}
	fmt.Fprintf(&b, "Products total: %d\n", snapshot.ProductTotal)
	for status, count := range snapshot.ProductsByStatus {
		fmt.Fprintf(&b, "Products %s: %d\n", status, count)
	}
	fmt.Fprintf(&b, "Low-stock SKUs: %d\n", snapshot.LowStockCount)

	if len(snapshot.LowStockItems) > 0 {
		b.WriteString("Low-stock examples:\n")
		for _, item := range snapshot.LowStockItems {
			fmt.Fprintf(&b, "- %s (%d units)\n", item.Name, item.Stock)
		}
	}

	if len(snapshot.RecentOrders) > 0 {
		b.WriteString("Recent orders:\n")
		for _, order := range snapshot.RecentOrders {
			fmt.Fprintf(
				&b,
				"- %s status=%s subtotal=%.2f items=%d\n",
				order.OrderNumber,
				order.Status,
				order.Subtotal,
				order.ItemCount,
			)
		}
	}

	return b.String()
}

func fallbackVendorDashboard(snapshot *VendorDashboardSnapshot) *dto.AiVendorDashboardResponse {
	priorities := make([]string, 0, 4)
	opportunities := make([]string, 0, 3)
	alerts := make([]string, 0, 3)

	pending := snapshot.OrdersByStatus["pending"]
	processing := snapshot.OrdersByStatus["processing"]
	if pending > 0 {
		priorities = append(priorities, fmt.Sprintf("Confirm or fulfill %d pending orders", pending))
	}
	if processing > 0 {
		priorities = append(priorities, fmt.Sprintf("Ship %d orders in processing", processing))
	}
	if snapshot.LowStockCount > 0 {
		priorities = append(priorities, fmt.Sprintf("Restock %d low-inventory products", snapshot.LowStockCount))
		alerts = append(alerts, fmt.Sprintf("%d products are below the low-stock threshold", snapshot.LowStockCount))
	}

	draftCount := snapshot.ProductsByStatus["draft"]
	if draftCount > 0 {
		opportunities = append(opportunities, fmt.Sprintf("Publish %d draft products to grow catalog coverage", draftCount))
	}
	if snapshot.OrderTotal == 0 && snapshot.ProductTotal > 0 {
		opportunities = append(opportunities, "Launch a featured promotion to drive first orders")
	}
	if len(opportunities) == 0 {
		opportunities = append(opportunities, "Review top sellers and refresh hero product imagery")
	}

	health := 88
	if snapshot.LowStockCount > 5 {
		health -= 12
	}
	if pending+processing > 10 {
		health -= 8
	}
	if health < 40 {
		health = 40
	}

	summary := fmt.Sprintf(
		"%s has %d products and %d orders on record.",
		snapshot.StoreName,
		snapshot.ProductTotal,
		snapshot.OrderTotal,
	)
	if pending+processing > 0 {
		summary += fmt.Sprintf(" %d orders need fulfillment attention.", pending+processing)
	}

	return &dto.AiVendorDashboardResponse{
		AiEnabled:     false,
		Summary:       summary,
		HealthScore:   health,
		Priorities:    priorities,
		Opportunities: opportunities,
		Alerts:        alerts,
		Sources:       []string{"orders", "products", "inventory"},
	}
}

func parseVendorDashboardJSON(content string) (*vendorDashboardPayload, error) {
	content = extractJSONObject(content)
	var payload vendorDashboardPayload
	if err := json.Unmarshal([]byte(content), &payload); err != nil {
		return nil, err
	}
	if strings.TrimSpace(payload.Summary) == "" {
		return nil, fmt.Errorf("missing summary")
	}
	return &payload, nil
}

package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

const (
	TaskReplenishmentReminders   = "replenishment_reminders"
	defaultReplenishmentOrders   = 12
	maxReplenishmentOrders       = 20
	defaultReplenishmentLimit    = 6
	maxReplenishmentLimit        = 8
	defaultReplenishmentSearch   = 6
)

// ReplenishmentQueries loads recent purchase history for reorder suggestions.
type ReplenishmentQueries interface {
	ListRecentOrders(ctx context.Context, userID uint, limit int) ([]models.Order, error)
}

type replenishmentPayload struct {
	Summary   string                      `json:"summary"`
	Reminders []dto.AiReplenishmentReminder `json:"reminders"`
}

// ReplenishmentReminders suggests when to reorder consumables and repeat purchases.
func (s *Service) ReplenishmentReminders(
	ctx context.Context,
	userID uint,
	subjectKey string,
	orders ReplenishmentQueries,
	search SearchQueries,
	req dto.AiReplenishmentRemindersRequest,
) (*dto.AiReplenishmentRemindersResponse, error) {
	if !s.Enabled() {
		return nil, utils.NewAppError(503, "AI is not enabled", nil)
	}
	if userID == 0 {
		return nil, utils.ErrUnauthorized("authentication required")
	}
	if orders == nil {
		return nil, utils.ErrInternal(fmt.Errorf("order queries not configured"))
	}
	if search == nil {
		return nil, utils.ErrInternal(fmt.Errorf("search not configured"))
	}
	if !s.chatRL.allow("replenishment_reminders:"+subjectKey) {
		return nil, utils.ErrTooManyRequests()
	}

	limit := req.Limit
	if limit <= 0 {
		limit = defaultReplenishmentLimit
	}
	if limit > maxReplenishmentLimit {
		limit = maxReplenishmentLimit
	}

	orderLimit := defaultReplenishmentOrders
	recentOrders, err := orders.ListRecentOrders(ctx, userID, orderLimit)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	purchaseLines := flattenPurchaseLines(recentOrders)
	if len(purchaseLines) == 0 {
		return &dto.AiReplenishmentRemindersResponse{
			Summary: "Place your first order and Luxe will learn which items you tend to reorder — skincare, coffee, pet supplies, and more.",
			Sources: []string{"Order history"},
		}, nil
	}

	facts, sources := replenishmentFacts(purchaseLines)
	system := `You suggest replenishment reminders for a luxury e-commerce shopper based on purchase history.
Respond with JSON only using this exact shape:
{"summary":"...","reminders":[{"product_id":0,"product_name":"...","category":"...","days_since_order":0,"urgency":"watch|due|soon|overdue","message":"...","search_query":"..."}]}
Rules:
- summary: 2 sentences on overall reorder posture.
- reminders: 2-6 items that are likely consumables or repeat purchases; skip one-off luxury buys unless due again.
- urgency: watch (monitor), due (reorder soon), soon (within ~2 weeks), overdue (past typical cycle).
- days_since_order: integer from the facts.
- product_id: use id from facts when reordering the same sku; 0 if suggesting a substitute category.
- search_query: short catalog keywords to find the item or a close substitute.
Use ONLY the purchase facts below. Do not invent orders or dates.
Plain text only — no markdown.

Purchase facts:
` + facts

	user := "Suggest replenishment reminders based on this purchase history."
	resp, err := s.complete(ctx, system, user, 1100)
	if err != nil {
		return nil, utils.NewAppError(503, "AI provider unavailable", err)
	}

	parsed, err := parseReplenishmentJSON(resp.Content, limit)
	if err != nil {
		return nil, utils.NewAppError(503, "AI returned invalid replenishment reminders", err)
	}

	response := &dto.AiReplenishmentRemindersResponse{
		Summary:   parsed.Summary,
		Reminders: parsed.Reminders,
		Sources:   uniqueStrings(sources),
	}

	seenProduct := make(map[uint]struct{})
	inStock := true
	for _, reminder := range parsed.Reminders {
		query := strings.TrimSpace(reminder.SearchQuery)
		if query == "" {
			continue
		}
		searchResult, err := search.GlobalSearch(ctx, dto.SearchRequest{
			Query:   query,
			Limit:   defaultReplenishmentSearch,
			Offset:  0,
			InStock: &inStock,
		})
		if err != nil {
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
			response.Recommendations = append(response.Recommendations, dto.AiRecommendedProduct{
				Product: product,
				Reason:  reminder.Message,
			})
			if len(response.Recommendations) >= 4 {
				break
			}
		}
		if len(response.Recommendations) >= 4 {
			break
		}
	}
	if len(response.Recommendations) > 0 {
		response.Sources = append(response.Sources, "Catalog search")
	}

	utils.Log.WithFields(map[string]any{
		"user_id": userID,
		"task":    TaskReplenishmentReminders,
		"orders":  len(recentOrders),
	}).Info("ai replenishment reminders completed")

	return response, nil
}

type purchaseLine struct {
	ProductID   uint
	ProductName string
	Category    string
	Quantity    int
	OrderDate   time.Time
	OrderStatus string
}

func flattenPurchaseLines(orders []models.Order) []purchaseLine {
	lines := make([]purchaseLine, 0)
	for _, order := range orders {
		if order.Status == constants.OrderStatusCancelled || order.Status == constants.OrderStatusRefunded {
			continue
		}
		for _, item := range order.Items {
			if item.ProductID == 0 {
				continue
			}
			name := "product"
			category := "unknown"
			if item.Product.ID != 0 {
				name = item.Product.Name
				category = productCategoryLabel(&item.Product)
			}
			lines = append(lines, purchaseLine{
				ProductID:   item.ProductID,
				ProductName: name,
				Category:    category,
				Quantity:    item.Quantity,
				OrderDate:   order.CreatedAt,
				OrderStatus: order.Status,
			})
		}
	}
	return lines
}

func replenishmentFacts(lines []purchaseLine) (string, []string) {
	var b strings.Builder
	sources := []string{"Order history"}
	now := time.Now()

	fmt.Fprintf(&b, "%d purchase lines (most recent first):\n", len(lines))
	count := 0
	for _, line := range lines {
		if count >= maxReplenishmentOrders*3 {
			break
		}
		days := int(now.Sub(line.OrderDate).Hours() / 24)
		fmt.Fprintf(&b, "- id %d | %s | category %s | qty %d | ordered %d days ago | status %s\n",
			line.ProductID,
			sanitizeAIText(line.ProductName),
			sanitizeAIText(line.Category),
			line.Quantity,
			days,
			line.OrderStatus,
		)
		count++
	}
	return b.String(), sources
}

func parseReplenishmentJSON(raw string, limit int) (*replenishmentPayload, error) {
	raw = extractJSONObject(raw)
	var parsed replenishmentPayload
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, err
	}
	if strings.TrimSpace(parsed.Summary) == "" {
		return nil, fmt.Errorf("missing summary")
	}

	allowedUrgency := map[string]struct{}{
		"watch":   {},
		"due":     {},
		"soon":    {},
		"overdue": {},
	}

	clean := make([]dto.AiReplenishmentReminder, 0, len(parsed.Reminders))
	for _, item := range parsed.Reminders {
		name := strings.TrimSpace(item.ProductName)
		message := strings.TrimSpace(item.Message)
		if name == "" || message == "" {
			continue
		}
		urgency := strings.TrimSpace(strings.ToLower(item.Urgency))
		if _, ok := allowedUrgency[urgency]; !ok {
			urgency = "watch"
		}
		clean = append(clean, dto.AiReplenishmentReminder{
			ProductID:      item.ProductID,
			ProductName:    name,
			Category:       strings.TrimSpace(item.Category),
			DaysSinceOrder: item.DaysSinceOrder,
			Urgency:        urgency,
			Message:        message,
			SearchQuery:    strings.TrimSpace(item.SearchQuery),
		})
		if len(clean) >= limit {
			break
		}
	}

	if len(clean) == 0 {
		return nil, fmt.Errorf("no valid reminders")
	}

	parsed.Reminders = clean
	return &parsed, nil
}

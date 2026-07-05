package pdp

import (
	"sort"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

const (
	defaultTimelineDays = 365
	maxTimelineDays     = 730
	maxTimelineEvents   = 24
)

var reviewMilestones = []int{10, 25, 50, 100}

func normalizeTimelineDays(days int) int {
	if days <= 0 {
		return defaultTimelineDays
	}
	if days > maxTimelineDays {
		return maxTimelineDays
	}
	return days
}

type timelineEvent struct {
	occurredAt time.Time
	typ        string
	meta       dto.ProductTimelineMeta
}

// buildProductTimeline merges catalog, pricing, inventory, workflow, and review signals.
func buildProductTimeline(
	product models.Product,
	priceRows []models.ProductPriceHistory,
	adjustments []models.InventoryAdjustment,
	workflowLogs []models.WorkflowTransitionLog,
	reviews []models.Review,
	days int,
) dto.ProductTimelineData {
	days = normalizeTimelineDays(days)
	since := time.Now().UTC().AddDate(0, 0, -days)

	events := make([]timelineEvent, 0, 16)

	events = append(events, timelineEvent{
		occurredAt: product.CreatedAt,
		typ:        "listed",
	})

	if product.PublishedAt != nil {
		events = append(events, timelineEvent{
			occurredAt: *product.PublishedAt,
			typ:        "published",
		})
	}

	var prevPrice float64
	for i, row := range priceRows {
		if i == 0 {
			prevPrice = row.Price
			continue
		}
		if row.Price == prevPrice {
			continue
		}
		typ := "price_increase"
		if row.Price < prevPrice {
			typ = "price_drop"
		}
		from := prevPrice
		to := row.Price
		events = append(events, timelineEvent{
			occurredAt: row.RecordedAt,
			typ:        typ,
			meta: dto.ProductTimelineMeta{
				PriceFrom: &from,
				PriceTo:   &to,
			},
		})
		prevPrice = row.Price
	}

	for _, adj := range adjustments {
		if adj.QuantityAfter == 0 && adj.QuantityBefore > 0 {
			qty := adj.QuantityAfter
			events = append(events, timelineEvent{
				occurredAt: adj.CreatedAt,
				typ:        "sold_out",
				meta:       dto.ProductTimelineMeta{StockQuantity: &qty},
			})
			continue
		}
		if adj.QuantityAfter > 0 && adj.QuantityBefore == 0 {
			qty := adj.QuantityAfter
			events = append(events, timelineEvent{
				occurredAt: adj.CreatedAt,
				typ:        "restocked",
				meta:       dto.ProductTimelineMeta{StockQuantity: &qty},
			})
		}
	}

	for _, log := range workflowLogs {
		stateCode := ""
		if log.ToState != nil {
			stateCode = log.ToState.Code
		}
		events = append(events, timelineEvent{
			occurredAt: log.CreatedAt,
			typ:        "status_change",
			meta: dto.ProductTimelineMeta{
				WorkflowState: stateCode,
				WorkflowEvent: log.Event,
			},
		})
	}

	for i, review := range reviews {
		count := i + 1
		if count == 1 {
			rating := review.Rating
			events = append(events, timelineEvent{
				occurredAt: review.CreatedAt,
				typ:        "first_review",
				meta: dto.ProductTimelineMeta{
					ReviewRating: &rating,
					ReviewCount:  &count,
				},
			})
			continue
		}
		if isReviewMilestone(count) {
			events = append(events, timelineEvent{
				occurredAt: review.CreatedAt,
				typ:        "reviews_milestone",
				meta:       dto.ProductTimelineMeta{ReviewCount: &count},
			})
		}
	}

	filtered := make([]timelineEvent, 0, len(events))
	for _, event := range events {
		if event.typ == "listed" || event.typ == "published" || !event.occurredAt.Before(since) {
			filtered = append(filtered, event)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].occurredAt.After(filtered[j].occurredAt)
	})

	if len(filtered) > maxTimelineEvents {
		filtered = filtered[:maxTimelineEvents]
	}

	points := make([]dto.ProductTimelineEvent, len(filtered))
	for i, event := range filtered {
		points[i] = dto.ProductTimelineEvent{
			Type:       event.typ,
			OccurredAt: event.occurredAt,
			Meta:       event.meta,
		}
	}

	return dto.ProductTimelineData{
		Days:   days,
		Events: points,
	}
}

func isReviewMilestone(count int) bool {
	for _, milestone := range reviewMilestones {
		if count == milestone {
			return true
		}
	}
	return false
}

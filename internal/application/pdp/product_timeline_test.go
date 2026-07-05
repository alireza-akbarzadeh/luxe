package pdp

import (
	"testing"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

func TestBuildProductTimelineOrdersNewestFirst(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	created := now.AddDate(0, -6, 0)
	published := now.AddDate(0, -5, 0)
	product := models.Product{
		CreatedAt:   created,
		PublishedAt: &published,
	}

	priceFrom := 100.0
	priceTo := 80.0
	priceRows := []models.ProductPriceHistory{
		{RecordedAt: now.AddDate(0, -2, 0), Price: 100},
		{RecordedAt: now.AddDate(0, -1, 0), Price: 80},
	}

	data := buildProductTimeline(product, priceRows, nil, nil, nil, 365)
	if len(data.Events) < 3 {
		t.Fatalf("expected at least 3 events, got %d", len(data.Events))
	}
	if data.Events[0].Type != "price_drop" {
		t.Fatalf("newest event = %q, want price_drop", data.Events[0].Type)
	}
	if data.Events[0].Meta.PriceFrom == nil || *data.Events[0].Meta.PriceFrom != priceFrom {
		t.Fatalf("price_from = %v, want %v", data.Events[0].Meta.PriceFrom, priceFrom)
	}
	if data.Events[0].Meta.PriceTo == nil || *data.Events[0].Meta.PriceTo != priceTo {
		t.Fatalf("price_to = %v, want %v", data.Events[0].Meta.PriceTo, priceTo)
	}
}

func TestBuildProductTimelineStockEvents(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	product := models.Product{CreatedAt: now.AddDate(0, -1, 0)}
	adjustments := []models.InventoryAdjustment{
		{CreatedAt: now.Add(-48 * time.Hour), QuantityBefore: 5, QuantityAfter: 0},
		{CreatedAt: now.Add(-24 * time.Hour), QuantityBefore: 0, QuantityAfter: 12},
	}

	data := buildProductTimeline(product, nil, adjustments, nil, nil, 90)

	var hasSoldOut, hasRestocked bool
	for _, event := range data.Events {
		switch event.Type {
		case "sold_out":
			hasSoldOut = true
		case "restocked":
			hasRestocked = true
		}
	}
	if !hasSoldOut || !hasRestocked {
		t.Fatalf("expected sold_out and restocked events, got %+v", data.Events)
	}
}

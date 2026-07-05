package pdp

import (
	"testing"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

func TestStockAvailabilityLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		stock     int
		threshold int
		want      string
	}{
		{stock: 0, threshold: 5, want: "out"},
		{stock: 3, threshold: 5, want: "low"},
		{stock: 8, threshold: 5, want: "medium"},
		{stock: 20, threshold: 5, want: "high"},
		{stock: 2, threshold: 0, want: "low"},
	}

	for _, tc := range tests {
		if got := stockAvailabilityLevel(tc.stock, tc.threshold); got != tc.want {
			t.Fatalf("stockAvailabilityLevel(%d, %d) = %q, want %q", tc.stock, tc.threshold, got, tc.want)
		}
	}
}

func TestBuildStockHeatmapDigitalSkipsPoints(t *testing.T) {
	t.Parallel()

	product := models.Product{IsDigital: true, TrackInventory: true, Stock: 99}
	data := buildStockHeatmap(product, nil, 90)

	if len(data.Points) != 0 {
		t.Fatalf("expected no points for digital product, got %d", len(data.Points))
	}
	if !data.IsDigital {
		t.Fatal("expected is_digital true")
	}
}

func TestBuildStockHeatmapAppliesLedgerAdjustment(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, time.UTC)

	product := models.Product{
		TrackInventory:    true,
		Stock:             4,
		LowStockThreshold: 5,
	}
	adjustments := []models.InventoryAdjustment{
		{
			CreatedAt:      today,
			QuantityBefore: 20,
			QuantityAfter:  4,
		},
	}

	data := buildStockHeatmap(product, adjustments, 3)
	if len(data.Points) != 3 {
		t.Fatalf("expected 3 points, got %d", len(data.Points))
	}

	last := data.Points[len(data.Points)-1]
	if last.Stock != 4 {
		t.Fatalf("last day stock = %d, want current product stock 4", last.Stock)
	}
	if last.Level != "low" {
		t.Fatalf("last day level = %q, want low", last.Level)
	}
	if data.LowStockDays < 1 {
		t.Fatal("expected at least one low-stock day")
	}
}

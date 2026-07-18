package pdp

import (
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

const defaultStockHeatmapDays = 90

func normalizeStockHeatmapDays(days int) int {
	if days <= 0 {
		return defaultStockHeatmapDays
	}
	if days > 365 {
		return 365
	}
	return days
}

func stockAvailabilityLevel(stock, lowThreshold int) string {
	if stock <= 0 {
		return "out"
	}
	if lowThreshold <= 0 {
		lowThreshold = 5
	}
	if stock <= lowThreshold {
		return "low"
	}
	if stock <= lowThreshold*2 {
		return "medium"
	}
	return "high"
}

// buildStockHeatmap reconstructs daily stock levels from adjustment ledger rows.
func buildStockHeatmap(product models.Product, adjustments []models.InventoryAdjustment, days int) dto.StockHeatmapData {
	days = normalizeStockHeatmapDays(days)
	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -(days - 1))

	data := dto.StockHeatmapData{
		TrackInventory:    product.TrackInventory,
		IsDigital:         product.IsDigital,
		CurrentStock:      product.Stock,
		LowStockThreshold: product.LowStockThreshold,
		Days:              days,
		Points:            make([]dto.StockHeatmapPoint, 0, days),
	}

	if product.IsDigital {
		return data
	}
	if !product.TrackInventory {
		return data
	}

	runningStock := product.Stock
	if prior, ok := seedStockBeforeWindow(product, adjustments, start); ok {
		runningStock = prior
	}

	adjIdx := 0
	for offset := 0; offset < days; offset++ {
		day := start.AddDate(0, 0, offset)
		dayEnd := day.Add(24 * time.Hour).Add(-time.Nanosecond)

		for adjIdx < len(adjustments) && !adjustments[adjIdx].CreatedAt.After(dayEnd) {
			runningStock = adjustments[adjIdx].QuantityAfter
			adjIdx++
		}

		stock := runningStock
		if offset == days-1 {
			stock = product.Stock
		}

		level := stockAvailabilityLevel(stock, product.LowStockThreshold)
		data.Points = append(data.Points, dto.StockHeatmapPoint{
			Date:  day,
			Stock: stock,
			Level: level,
		})

		switch level {
		case "out":
			data.OutOfStockDays++
		case "low":
			data.LowStockDays++
		default:
			data.InStockDays++
		}
	}

	return data
}

func seedStockBeforeWindow(product models.Product, adjustments []models.InventoryAdjustment, start time.Time) (int, bool) {
	if len(adjustments) == 0 {
		return product.Stock, true
	}
	if adjustments[0].CreatedAt.After(start) {
		return adjustments[0].QuantityBefore, true
	}
	return product.Stock, true
}

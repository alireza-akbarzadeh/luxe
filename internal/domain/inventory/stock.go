package inventory

import "errors"

var (
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrNegativeStock     = errors.New("stock cannot be negative")
)

// ProductStock describes inventory fields used by stock rules.
type ProductStock struct {
	TrackInventory bool
	AllowBackorder bool
	Stock          int
}

// StockAvailable reports whether quantity can be fulfilled without going negative.
func StockAvailable(p ProductStock, quantity int) bool {
	if !p.TrackInventory {
		return true
	}
	if p.AllowBackorder {
		return true
	}
	return p.Stock >= quantity
}

// ShouldAdjustStock reports whether inventory ledger changes apply to a product.
func ShouldAdjustStock(p ProductStock) bool {
	return p.TrackInventory
}

// CanApplyDelta validates a stock change before persistence.
func CanApplyDelta(p ProductStock, before, delta int, skipAvailabilityCheck bool) error {
	if !p.TrackInventory {
		return nil
	}
	after := before + delta
	if after < 0 {
		return ErrNegativeStock
	}
	if delta < 0 && !skipAvailabilityCheck && !p.AllowBackorder {
		if !StockAvailable(ProductStock{
			TrackInventory: p.TrackInventory,
			AllowBackorder: p.AllowBackorder,
			Stock:          before,
		}, -delta) {
			return ErrInsufficientStock
		}
	}
	return nil
}

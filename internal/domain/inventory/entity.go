package inventory

// Adjustment describes a stock ledger entry.
type Adjustment struct {
	ProductID      uint
	QuantityDelta  int
	QuantityBefore int
	QuantityAfter  int
	AdjustmentType string
}

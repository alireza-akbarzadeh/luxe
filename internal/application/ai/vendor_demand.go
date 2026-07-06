package ai

// VendorProductDemand captures per-SKU sales velocity and catalog facts for vendor AI tools.
type VendorProductDemand struct {
	ProductID         uint
	Name              string
	Price             float64
	CompareAtPrice    *float64
	Cost              *float64
	Stock             int
	LowStockThreshold int
	UnitsSold         int64
	Revenue           float64
}

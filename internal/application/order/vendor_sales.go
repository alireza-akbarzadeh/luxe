package order

// VendorSalesSummary aggregates revenue metrics for a vendor store in a date range.
type VendorSalesSummary struct {
	Revenue       float64
	OrderCount    int64
	UnitsSold     int64
	AvgOrderValue float64
}

// VendorTopProduct ranks a product by store-scoped sales in a period.
type VendorTopProduct struct {
	ProductID uint
	Name      string
	Revenue   float64
	Units     int64
}

// VendorDailySales is one day of store revenue and order volume.
type VendorDailySales struct {
	Date    string
	Revenue float64
	Orders  int64
}

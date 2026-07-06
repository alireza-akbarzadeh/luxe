package order

import "time"

// VendorStoreCustomer aggregates purchase behavior for a buyer at a vendor store.
type VendorStoreCustomer struct {
	UserID      uint
	FirstName   string
	LastName    string
	Email       string
	OrderCount  int64
	TotalSpend  float64
	LastOrderAt time.Time
}

package pdp

// StockNotification represents a back-in-stock subscription.
type StockNotification struct {
	UserID    uint
	ProductID uint
	Status    string
}

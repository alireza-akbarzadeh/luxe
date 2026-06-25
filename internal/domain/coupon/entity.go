package coupon

// Coupon is the discount aggregate root (stub).
type Coupon struct {
	ID           uint
	Code         string
	DiscountType string
	IsActive     bool
}

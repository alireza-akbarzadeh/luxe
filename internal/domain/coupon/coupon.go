package coupon

import (
	"errors"
	"time"
)

var (
	ErrUsageLimitExceeded = errors.New("coupon usage limit exceeded")
	ErrBelowMinimumOrder  = errors.New("order total below minimum amount")
	ErrAlreadyUsedByUser  = errors.New("coupon already used by this user")
	ErrInactiveOrExpired  = errors.New("invalid or expired coupon")
)

const (
	DiscountTypePercentage = "percentage"
	DiscountTypeFixed      = "fixed"
)

// Coupon holds fields used by pure validation and discount rules.
type Coupon struct {
	DiscountType       string
	DiscountValue      float64
	MinimumOrderAmount float64
	MaxDiscountAmount  *float64
	UsageLimit         int
	UsedCount          int
	IsActive           bool
	StartDate          time.Time
	EndDate            time.Time
}

// ValidateEligibility checks whether a coupon can apply to an order for a user.
func ValidateEligibility(c Coupon, orderTotal float64, userUsageCount int, now time.Time) error {
	if !c.IsActive || now.Before(c.StartDate) || now.After(c.EndDate) {
		return ErrInactiveOrExpired
	}
	if c.UsageLimit > 0 && c.UsedCount >= c.UsageLimit {
		return ErrUsageLimitExceeded
	}
	if orderTotal < c.MinimumOrderAmount {
		return ErrBelowMinimumOrder
	}
	if userUsageCount > 0 {
		return ErrAlreadyUsedByUser
	}
	return nil
}

// CalculateDiscount returns the discount amount for an order total.
func CalculateDiscount(c Coupon, orderTotal float64) float64 {
	if c.DiscountType == DiscountTypePercentage {
		discount := orderTotal * (c.DiscountValue / 100)
		if c.MaxDiscountAmount != nil && discount > *c.MaxDiscountAmount {
			return *c.MaxDiscountAmount
		}
		return discount
	}
	discount := c.DiscountValue
	if discount > orderTotal {
		return orderTotal
	}
	return discount
}

// IsExhaustedAfterUse reports whether applying one more use hits the usage limit.
func IsExhaustedAfterUse(c Coupon) bool {
	return c.UsageLimit > 0 && c.UsedCount+1 >= c.UsageLimit
}

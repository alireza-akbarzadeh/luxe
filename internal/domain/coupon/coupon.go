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
	ErrInsufficientItems  = errors.New("cart does not meet item quantity requirement")
)

const (
	DiscountTypePercentage = "percentage"
	DiscountTypeFixed      = "fixed"

	ApplicationTypeCode      = "code"
	ApplicationTypeAutomatic = "automatic"
	ApplicationTypeBOGO      = "bogo"
)

// Coupon holds fields used by pure validation and discount rules.
type Coupon struct {
	ApplicationType        string
	DiscountType           string
	DiscountValue          float64
	MinimumOrderAmount     float64
	MaxDiscountAmount      *float64
	UsageLimit             int
	UsedCount              int
	IsActive               bool
	StartDate              time.Time
	EndDate                time.Time
	BogoBuyQuantity        int
	BogoGetQuantity        int
	BogoGetDiscountPercent float64
	MinItemQuantity        int
}

// ValidateEligibility checks whether a coupon can apply to an order for a user.
func ValidateEligibility(c Coupon, orderTotal float64, userUsageCount int, itemCount int, now time.Time) error {
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
	if c.ApplicationType == ApplicationTypeBOGO {
		required := c.BogoBuyQuantity + c.BogoGetQuantity
		if c.MinItemQuantity > required {
			required = c.MinItemQuantity
		}
		if itemCount > 0 && itemCount < required {
			return ErrInsufficientItems
		}
	} else if c.MinItemQuantity > 0 && itemCount > 0 && itemCount < c.MinItemQuantity {
		return ErrInsufficientItems
	}
	return nil
}

// CalculateDiscount returns the discount amount for an order total.
func CalculateDiscount(c Coupon, orderTotal float64, itemCount int) float64 {
	if c.ApplicationType == ApplicationTypeBOGO {
		return CalculateBOGODiscount(c, orderTotal, itemCount)
	}

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

// CalculateBOGODiscount applies buy-X-get-Y style savings using average unit price.
func CalculateBOGODiscount(c Coupon, orderTotal float64, itemCount int) float64 {
	if itemCount <= 0 || orderTotal <= 0 {
		return 0
	}
	buy := c.BogoBuyQuantity
	get := c.BogoGetQuantity
	if buy <= 0 {
		buy = 1
	}
	if get <= 0 {
		get = 1
	}
	groupSize := buy + get
	if itemCount < groupSize {
		return 0
	}

	unitPrice := orderTotal / float64(itemCount)
	sets := itemCount / groupSize
	getUnits := sets * get
	percent := c.BogoGetDiscountPercent
	if percent <= 0 {
		percent = 100
	}
	return unitPrice * float64(getUnits) * (percent / 100)
}

// IsExhaustedAfterUse reports whether applying one more use hits the usage limit.
func IsExhaustedAfterUse(c Coupon) bool {
	return c.UsageLimit > 0 && c.UsedCount+1 >= c.UsageLimit
}

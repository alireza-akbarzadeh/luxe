package membership

import (
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

// IsPlusActive reports whether the user currently has an active Luxe Plus membership.
func IsPlusActive(user *models.User) bool {
	if user == nil || user.MembershipTier != constants.MembershipTierPlus {
		return false
	}
	if user.PlusExpiresAt != nil && user.PlusExpiresAt.Before(time.Now()) {
		return false
	}
	return true
}

// ReturnWindowDays returns the return window for the user's membership tier.
func ReturnWindowDays(user *models.User) int {
	if IsPlusActive(user) {
		return constants.PlusReturnWindowDays
	}
	return constants.FreeReturnWindowDays
}

// OrderDiscountAmount calculates the Plus member discount on a subtotal (after coupons).
func OrderDiscountAmount(user *models.User, subtotalAfterCoupon float64) float64 {
	if !IsPlusActive(user) || subtotalAfterCoupon <= 0 {
		return 0
	}
	discount := subtotalAfterCoupon * float64(constants.PlusCheckoutDiscountPct) / 100
	return discount
}

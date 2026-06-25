package coupon

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateEligibility(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	base := Coupon{
		DiscountType:       DiscountTypePercentage,
		DiscountValue:      10,
		MinimumOrderAmount: 50,
		UsageLimit:         100,
		UsedCount:          0,
		IsActive:           true,
		StartDate:          now.Add(-24 * time.Hour),
		EndDate:            now.Add(24 * time.Hour),
	}

	t.Run("valid", func(t *testing.T) {
		require.NoError(t, ValidateEligibility(base, 100, 0, now))
	})

	t.Run("below minimum", func(t *testing.T) {
		err := ValidateEligibility(base, 40, 0, now)
		require.ErrorIs(t, err, ErrBelowMinimumOrder)
	})

	t.Run("already used", func(t *testing.T) {
		err := ValidateEligibility(base, 100, 1, now)
		require.ErrorIs(t, err, ErrAlreadyUsedByUser)
	})

	t.Run("expired", func(t *testing.T) {
		err := ValidateEligibility(base, 100, 0, now.Add(48*time.Hour))
		require.ErrorIs(t, err, ErrInactiveOrExpired)
	})
}

func TestCalculateDiscount(t *testing.T) {
	t.Run("percentage capped", func(t *testing.T) {
		max := 20.0
		c := Coupon{DiscountType: DiscountTypePercentage, DiscountValue: 50, MaxDiscountAmount: &max}
		assert.Equal(t, 20.0, CalculateDiscount(c, 100))
	})

	t.Run("fixed capped by order total", func(t *testing.T) {
		c := Coupon{DiscountType: DiscountTypeFixed, DiscountValue: 80}
		assert.Equal(t, 50.0, CalculateDiscount(c, 50))
	})
}

func TestIsExhaustedAfterUse(t *testing.T) {
	c := Coupon{UsageLimit: 5, UsedCount: 4}
	assert.True(t, IsExhaustedAfterUse(c))
}

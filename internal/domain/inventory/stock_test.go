package inventory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStockAvailable(t *testing.T) {
	t.Run("untracked inventory", func(t *testing.T) {
		p := ProductStock{TrackInventory: false, Stock: 0}
		assert.True(t, StockAvailable(p, 99))
	})

	t.Run("backorder allowed", func(t *testing.T) {
		p := ProductStock{TrackInventory: true, AllowBackorder: true, Stock: 0}
		assert.True(t, StockAvailable(p, 1))
	})

	t.Run("insufficient stock", func(t *testing.T) {
		p := ProductStock{TrackInventory: true, Stock: 2}
		assert.False(t, StockAvailable(p, 3))
	})
}

func TestCanApplyDelta(t *testing.T) {
	p := ProductStock{TrackInventory: true, Stock: 5}

	t.Run("sale within stock", func(t *testing.T) {
		require.NoError(t, CanApplyDelta(p, 5, -2, false))
	})

	t.Run("sale exceeds stock", func(t *testing.T) {
		err := CanApplyDelta(p, 5, -6, false)
		require.ErrorIs(t, err, ErrNegativeStock)
	})

	t.Run("skip availability for cancel restore", func(t *testing.T) {
		require.NoError(t, CanApplyDelta(p, 5, 3, true))
	})

	t.Run("negative stock rejected", func(t *testing.T) {
		err := CanApplyDelta(p, 1, -5, true)
		require.ErrorIs(t, err, ErrNegativeStock)
	})
}

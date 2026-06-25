package cart

import (
	"testing"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestProductStockAvailable(t *testing.T) {
	t.Run("untracked inventory", func(t *testing.T) {
		p := models.Product{TrackInventory: false, Stock: 0}
		assert.True(t, ProductStockAvailable(p, 99))
	})

	t.Run("backorder allowed", func(t *testing.T) {
		p := models.Product{TrackInventory: true, AllowBackorder: true, Stock: 0}
		assert.True(t, ProductStockAvailable(p, 1))
	})

	t.Run("insufficient stock", func(t *testing.T) {
		p := models.Product{TrackInventory: true, Stock: 2}
		assert.False(t, ProductStockAvailable(p, 3))
	})
}

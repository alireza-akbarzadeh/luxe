package cart

import "github.com/alireza-akbarzadeh/luxe/internal/models"

// ProductStockAvailable reports whether quantity can be fulfilled for a product row.
func ProductStockAvailable(product models.Product, quantity int) bool {
	if !product.TrackInventory {
		return true
	}
	if product.AllowBackorder {
		return true
	}
	return product.Stock >= quantity
}

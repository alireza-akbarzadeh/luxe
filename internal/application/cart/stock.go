package cart

import (
	"github.com/alireza-akbarzadeh/luxe/internal/domain/inventory"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

// ProductStockAvailable reports whether quantity can be fulfilled for a product row.
func ProductStockAvailable(product models.Product, quantity int) bool {
	return inventory.StockAvailable(productStockFromModel(product), quantity)
}

func productStockFromModel(product models.Product) inventory.ProductStock {
	return inventory.ProductStock{
		TrackInventory: product.TrackInventory,
		AllowBackorder: product.AllowBackorder,
		Stock:          product.Stock,
	}
}

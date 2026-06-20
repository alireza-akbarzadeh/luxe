package integration

import (
	"testing"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/stretchr/testify/require"
)

func seedProduct(t *testing.T, suffix string) *models.Product {
	store := models.Store{
		Name:   "Integration Store " + suffix,
		Slug:   "integration-store-" + suffix,
		Status: "active",
	}
	require.NoError(t, testDB.Create(&store).Error)

	product := models.Product{
		Name:           "Integration Product " + suffix,
		Price:          29.99,
		Stock:          100,
		SKU:            "SKU-" + suffix,
		Slug:           "product-" + suffix,
		Status:         "active",
		StoreID:        store.ID,
		TrackInventory: true,
		AllowBackorder: false,
	}
	require.NoError(t, testDB.Create(&product).Error)
	return &product
}

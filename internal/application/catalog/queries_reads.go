package catalog

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

// GetDetailedByID loads a product with relations for HTTP responses.
func (q *Queries) GetDetailedByID(ctx context.Context, id uint) (*models.Product, error) {
	return q.reader.GetDetailedByID(ctx, id)
}

// GetDetailedBySlug loads a product by slug with relations.
func (q *Queries) GetDetailedBySlug(ctx context.Context, slug string) (*models.Product, error) {
	return q.reader.GetDetailedBySlug(ctx, slug)
}

// ListDetailed returns a paginated product list with admin filters.
func (q *Queries) ListDetailed(ctx context.Context, limit, offset int, filters dto.ProductListFilters) ([]*models.Product, int64, error) {
	return q.reader.ListDetailed(ctx, limit, offset, filters)
}

// ExistsByID reports whether a product row exists.
func (q *Queries) ExistsByID(ctx context.Context, id uint) (bool, error) {
	return q.reader.ExistsByID(ctx, id)
}

// SKUTaken checks SKU uniqueness excluding an optional product id.
func (q *Queries) SKUTaken(ctx context.Context, sku string, excludeID uint) (bool, error) {
	return q.reader.SKUTaken(ctx, sku, excludeID)
}

// GetRelated returns products in the same category as the given product.
func (q *Queries) GetRelated(ctx context.Context, productID uint, limit int) ([]*models.Product, error) {
	return q.reader.GetRelated(ctx, productID, limit)
}

// GetSuggestions returns product suggestions based on cart product categories.
func (q *Queries) GetSuggestions(ctx context.Context, productIDs []uint, limit int) ([]*models.Product, error) {
	return q.reader.GetSuggestions(ctx, productIDs, limit)
}

// FindLowStockActive returns active products at or below their threshold.
func (q *Queries) FindLowStockActive(ctx context.Context) ([]models.Product, error) {
	return q.reader.FindLowStockActive(ctx, constants.ProductStatusActive)
}

// GetVendorStoreProductStats returns product count summaries for a vendor store.
func (q *Queries) GetVendorStoreProductStats(ctx context.Context, storeID uint) (dto.VendorProductStats, error) {
	return q.reader.CountByStoreStatus(ctx, storeID)
}

// BuildSearchDocument builds the search index document for a product.
func (q *Queries) BuildSearchDocument(ctx context.Context, product *models.Product) string {
	var category *models.Category
	if product.CategoryID != nil {
		cat, err := q.reader.FindCategoryByID(ctx, *product.CategoryID)
		if err == nil {
			category = cat
		}
	}
	return dto.BuildProductSearchDocument(product, category)
}

package catalog

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

// ProductReader loads product models with relations for HTTP handlers and application use cases.
type ProductReader interface {
	GetDetailedByID(ctx context.Context, id uint) (*models.Product, error)
	GetDetailedBySlug(ctx context.Context, slug string) (*models.Product, error)
	ListDetailed(ctx context.Context, limit, offset int, filters dto.ProductListFilters) ([]*models.Product, int64, error)
	ExistsByID(ctx context.Context, id uint) (bool, error)
	SKUTaken(ctx context.Context, sku string, excludeID uint) (bool, error)
	GetRelated(ctx context.Context, productID uint, limit int) ([]*models.Product, error)
	GetSuggestions(ctx context.Context, productIDs []uint, limit int) ([]*models.Product, error)
	FindLowStockActive(ctx context.Context, activeStatus string) ([]models.Product, error)
	FindCategoryByID(ctx context.Context, id uint) (*models.Category, error)
	CountByStoreStatus(ctx context.Context, storeID uint) (dto.VendorProductStats, error)
}

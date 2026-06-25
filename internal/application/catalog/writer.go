package catalog

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

// ProductWriter persists product models (implemented by postgres.ProductRepository).
type ProductWriter interface {
	CreateModel(ctx context.Context, product *models.Product) error
	SaveModel(ctx context.Context, product *models.Product) error
	UpdateStockColumn(ctx context.Context, productID uint, stock int) error
	ReplaceAttributes(ctx context.Context, productID uint, attrs []models.ProductAttribute) error
	BulkCreateModels(ctx context.Context, products []*models.Product) error
	BulkDeleteByIDs(ctx context.Context, ids []uint) (int64, error)
	DeleteByID(ctx context.Context, id uint) (int64, error)
}

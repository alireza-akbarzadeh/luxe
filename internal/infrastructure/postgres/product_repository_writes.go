package postgres

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// CreateModel inserts a new product row.
func (r *ProductRepository) CreateModel(ctx context.Context, product *models.Product) error {
	return r.db.WithContext(ctx).Create(product).Error
}

// SaveModel persists product field changes.
func (r *ProductRepository) SaveModel(ctx context.Context, product *models.Product) error {
	return r.db.WithContext(ctx).Save(product).Error
}

// UpdateStockColumn sets stock on a product without loading the full row.
func (r *ProductRepository) UpdateStockColumn(ctx context.Context, productID uint, stock int) error {
	return r.db.WithContext(ctx).Model(&models.Product{}).Where("id = ?", productID).Update("stock", stock).Error
}

// ReplaceAttributes deletes existing attributes and inserts new ones for a product.
func (r *ProductRepository) ReplaceAttributes(ctx context.Context, productID uint, attrs []models.ProductAttribute) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("product_id = ?", productID).Delete(&models.ProductAttribute{}).Error; err != nil {
			return err
		}
		if len(attrs) == 0 {
			return nil
		}
		return tx.Create(&attrs).Error
	})
}

// BulkCreateModels inserts multiple products in one transaction.
func (r *ProductRepository) BulkCreateModels(ctx context.Context, products []*models.Product) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, p := range products {
			if err := tx.Create(p).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// BulkDeleteByIDs soft-deletes products by primary keys.
func (r *ProductRepository) BulkDeleteByIDs(ctx context.Context, ids []uint) (int64, error) {
	result := r.db.WithContext(ctx).Where("id IN ?", ids).Delete(&models.Product{})
	return result.RowsAffected, result.Error
}

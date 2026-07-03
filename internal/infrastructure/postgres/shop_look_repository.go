package postgres

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// ShopLookRepository loads shoppable look scenes and tags.
type ShopLookRepository struct {
	db *gorm.DB
}

// NewShopLookRepository creates a GORM-backed shop look repository.
func NewShopLookRepository(db *gorm.DB) *ShopLookRepository {
	return &ShopLookRepository{db: db}
}

// ListActive returns active shop looks ordered for storefront display.
func (r *ShopLookRepository) ListActive(ctx context.Context, limit int) ([]models.ShopLook, error) {
	if limit <= 0 {
		limit = 12
	}
	if limit > 24 {
		limit = 24
	}

	var looks []models.ShopLook
	err := r.db.WithContext(ctx).
		Model(&models.ShopLook{}).
		Where("is_active = ?", true).
		Preload("Tags", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, id ASC")
		}).
		Order("sort_order ASC, created_at DESC").
		Limit(limit).
		Find(&looks).Error
	return looks, err
}

// GetBySlug loads an active shop look with tags and products.
func (r *ShopLookRepository) GetBySlug(ctx context.Context, slug string) (*models.ShopLook, error) {
	var look models.ShopLook
	err := r.db.WithContext(ctx).
		Where("slug = ? AND is_active = ?", slug, true).
		Preload("Tags", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, id ASC")
		}).
		Preload("Tags.Product").
		Preload("Tags.Product.Category").
		Preload("Tags.Product.Brand").
		Preload("Tags.Product.Attributes").
		First(&look).Error
	if err != nil {
		return nil, err
	}
	return &look, nil
}

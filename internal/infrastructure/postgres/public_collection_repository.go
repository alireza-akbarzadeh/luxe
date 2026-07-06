package postgres

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// PublicCollectionRepository loads community public collections.
type PublicCollectionRepository struct {
	db *gorm.DB
}

// NewPublicCollectionRepository creates a GORM-backed public collection repository.
func NewPublicCollectionRepository(db *gorm.DB) *PublicCollectionRepository {
	return &PublicCollectionRepository{db: db}
}

// ListActive returns active public collections ordered for discovery.
func (r *PublicCollectionRepository) ListActive(ctx context.Context, limit int) ([]models.PublicCollection, error) {
	if limit <= 0 {
		limit = 12
	}
	if limit > 24 {
		limit = 24
	}

	var collections []models.PublicCollection
	err := r.db.WithContext(ctx).
		Model(&models.PublicCollection{}).
		Where("is_active = ?", true).
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, id ASC")
		}).
		Order("sort_order ASC, created_at DESC").
		Limit(limit).
		Find(&collections).Error
	return collections, err
}

// GetBySlug loads an active public collection with items and products.
func (r *PublicCollectionRepository) GetBySlug(ctx context.Context, slug string) (*models.PublicCollection, error) {
	var collection models.PublicCollection
	err := r.db.WithContext(ctx).
		Where("slug = ? AND is_active = ?", slug, true).
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, id ASC")
		}).
		Preload("Items.Product").
		Preload("Items.Product.Category").
		Preload("Items.Product.Brand").
		Preload("Items.Product.Attributes").
		First(&collection).Error
	if err != nil {
		return nil, err
	}
	return &collection, nil
}

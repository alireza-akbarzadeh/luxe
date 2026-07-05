package postgres

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// CreatorStorefrontRepository loads creator profiles and curated picks.
type CreatorStorefrontRepository struct {
	db *gorm.DB
}

// NewCreatorStorefrontRepository creates a GORM-backed creator repository.
func NewCreatorStorefrontRepository(db *gorm.DB) *CreatorStorefrontRepository {
	return &CreatorStorefrontRepository{db: db}
}

// ListActive returns active creators ordered for storefront discovery.
func (r *CreatorStorefrontRepository) ListActive(ctx context.Context, limit int) ([]models.Creator, error) {
	if limit <= 0 {
		limit = 12
	}
	if limit > 24 {
		limit = 24
	}

	var creators []models.Creator
	err := r.db.WithContext(ctx).
		Model(&models.Creator{}).
		Where("is_active = ?", true).
		Preload("Picks", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, id ASC")
		}).
		Order("sort_order ASC, created_at DESC").
		Limit(limit).
		Find(&creators).Error
	return creators, err
}

// GetBySlug loads an active creator with picks and products.
func (r *CreatorStorefrontRepository) GetBySlug(ctx context.Context, slug string) (*models.Creator, error) {
	var creator models.Creator
	err := r.db.WithContext(ctx).
		Where("slug = ? AND is_active = ?", slug, true).
		Preload("Picks", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, id ASC")
		}).
		Preload("Picks.Product").
		Preload("Picks.Product.Category").
		Preload("Picks.Product.Brand").
		Preload("Picks.Product.Attributes").
		First(&creator).Error
	if err != nil {
		return nil, err
	}
	return &creator, nil
}

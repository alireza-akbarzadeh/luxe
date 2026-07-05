package postgres

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// CommunityShoppingListRepository loads public community shopping lists.
type CommunityShoppingListRepository struct {
	db *gorm.DB
}

// NewCommunityShoppingListRepository creates a GORM-backed community list repository.
func NewCommunityShoppingListRepository(db *gorm.DB) *CommunityShoppingListRepository {
	return &CommunityShoppingListRepository{db: db}
}

// ListActive returns active community lists ordered for discovery.
func (r *CommunityShoppingListRepository) ListActive(ctx context.Context, limit int) ([]models.CommunityShoppingList, error) {
	if limit <= 0 {
		limit = 12
	}
	if limit > 24 {
		limit = 24
	}

	var lists []models.CommunityShoppingList
	err := r.db.WithContext(ctx).
		Model(&models.CommunityShoppingList{}).
		Where("is_active = ?", true).
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, id ASC")
		}).
		Order("sort_order ASC, created_at DESC").
		Limit(limit).
		Find(&lists).Error
	return lists, err
}

// GetBySlug loads an active community list with items and products.
func (r *CommunityShoppingListRepository) GetBySlug(ctx context.Context, slug string) (*models.CommunityShoppingList, error) {
	var list models.CommunityShoppingList
	err := r.db.WithContext(ctx).
		Where("slug = ? AND is_active = ?", slug, true).
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, id ASC")
		}).
		Preload("Items.Product").
		Preload("Items.Product.Category").
		Preload("Items.Product.Brand").
		Preload("Items.Product.Attributes").
		First(&list).Error
	if err != nil {
		return nil, err
	}
	return &list, nil
}

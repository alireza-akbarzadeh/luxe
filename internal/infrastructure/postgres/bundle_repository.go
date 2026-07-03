package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// BundleRepository loads products for smart bundle generation.
type BundleRepository struct {
	db *gorm.DB
}

// NewBundleRepository creates a GORM-backed bundle repository.
func NewBundleRepository(db *gorm.DB) *BundleRepository {
	return &BundleRepository{db: db}
}

// GetProductsWithDetails loads products with category, brand, store, and attributes.
func (r *BundleRepository) GetProductsWithDetails(ctx context.Context, productIDs []uint) ([]models.Product, error) {
	if len(productIDs) == 0 {
		return nil, nil
	}
	var products []models.Product
	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Brand").
		Preload("Store").
		Preload("Attributes").
		Where("id IN ?", productIDs).
		Find(&products).Error
	return products, err
}

// FindComplementCandidates returns active in-stock products compatible with the anchor.
func (r *BundleRepository) FindComplementCandidates(
	ctx context.Context,
	anchor models.Product,
	excludeIDs []uint,
	keywords []string,
	limit int,
) ([]*models.Product, error) {
	if limit <= 0 {
		limit = 12
	}

	exclude := append([]uint{anchor.ID}, excludeIDs...)
	q := r.db.WithContext(ctx).
		Model(&models.Product{}).
		Preload("Category").
		Preload("Brand").
		Preload("Store").
		Where("id NOT IN ?", exclude).
		Where("status = ?", constants.ProductStatusActive).
		Where("(stock > 0 OR allow_backorder = ?)", true)

	if len(keywords) > 0 {
		var parts []string
		var args []any
		for _, kw := range keywords {
			pattern := "%" + kw + "%"
			parts = append(parts, "(name ILIKE ? OR description ILIKE ?)")
			args = append(args, pattern, pattern)
		}
		q = q.Where(strings.Join(parts, " OR "), args...)
	}

	if anchor.StoreID > 0 {
		q = q.Order(fmt.Sprintf("CASE WHEN store_id = %d THEN 0 ELSE 1 END", anchor.StoreID))
	}
	if anchor.CategoryID != nil {
		q = q.Order(fmt.Sprintf("CASE WHEN category_id = %d THEN 0 ELSE 1 END", *anchor.CategoryID))
	}

	var products []*models.Product
	err := q.
		Order("rating DESC, reviews_count DESC, created_at DESC").
		Limit(limit).
		Find(&products).Error
	return products, err
}

package postgres

import (
	"context"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// StorefrontRepository serves curated landing-page catalog queries.
type StorefrontRepository struct {
	db *gorm.DB
}

// NewStorefrontRepository creates a GORM-backed storefront repository.
func NewStorefrontRepository(db *gorm.DB) *StorefrontRepository {
	return &StorefrontRepository{db: db}
}

var storefrontProductPreloads = []string{"Category", "Brand", "Attributes", "WorkflowState"}

func (r *StorefrontRepository) preloadProducts(q *gorm.DB) *gorm.DB {
	for _, name := range storefrontProductPreloads {
		q = q.Preload(name)
	}
	return q
}

// ListDiscountedProducts returns active products on sale ordered by discount depth.
func (r *StorefrontRepository) ListDiscountedProducts(ctx context.Context, limit, offset int) ([]*models.Product, int64, error) {
	base := r.db.WithContext(ctx).Model(&models.Product{}).
		Where("status = ?", constants.ProductStatusActive).
		Where("compare_at_price IS NOT NULL AND compare_at_price > price")

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var products []*models.Product
	q := r.preloadProducts(base).
		Order("(compare_at_price - price) / NULLIF(compare_at_price, 0) DESC").
		Order("reviews_count DESC").
		Limit(limit).Offset(offset)
	if err := q.Find(&products).Error; err != nil {
		return nil, 0, err
	}
	return products, total, nil
}

type bestSellerRow struct {
	ProductID uint
	UnitsSold int64
}

// ListBestSellerIDs returns product ids ranked by units sold in paid orders.
func (r *StorefrontRepository) ListBestSellerIDs(ctx context.Context, limit int) ([]bestSellerRow, error) {
	since := time.Now().AddDate(0, -6, 0)
	var rows []bestSellerRow
	err := r.db.WithContext(ctx).Model(&models.OrderItem{}).
		Select(`
			order_items.product_id AS product_id,
			COALESCE(SUM(order_items.quantity), 0) AS units_sold`).
		Joins("INNER JOIN orders o ON o.id = order_items.order_id AND o.deleted_at IS NULL").
		Joins("INNER JOIN products p ON p.id = order_items.product_id AND p.deleted_at IS NULL").
		Where("o.created_at >= ? AND o.status IN ?", since, revenueOrderStatuses).
		Where("p.status = ?", constants.ProductStatusActive).
		Group("order_items.product_id").
		Order("units_sold DESC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

// ListProductsByIDs loads products preserving the given id order.
func (r *StorefrontRepository) ListProductsByIDs(ctx context.Context, ids []uint) ([]*models.Product, error) {
	if len(ids) == 0 {
		return []*models.Product{}, nil
	}
	var products []*models.Product
	q := r.preloadProducts(r.db.WithContext(ctx))
	if err := q.Where("id IN ?", ids).Find(&products).Error; err != nil {
		return nil, err
	}
	byID := make(map[uint]*models.Product, len(products))
	for _, p := range products {
		byID[p.ID] = p
	}
	ordered := make([]*models.Product, 0, len(ids))
	for _, id := range ids {
		if p, ok := byID[id]; ok {
			ordered = append(ordered, p)
		}
	}
	return ordered, nil
}

type popularBrandRow struct {
	models.Brand
	ProductCount int64 `gorm:"column:product_count"`
}

// ListPopularBrands returns brands with the most active catalog products.
func (r *StorefrontRepository) ListPopularBrands(ctx context.Context, limit int) ([]popularBrandRow, error) {
	var rows []popularBrandRow
	err := r.db.WithContext(ctx).Table("brands").
		Select("brands.*, COUNT(products.id) AS product_count").
		Joins("INNER JOIN products ON products.brand_id = brands.id AND products.deleted_at IS NULL").
		Where("products.status = ?", constants.ProductStatusActive).
		Where("brands.status IN ?", []string{"active", "published"}).
		Group("brands.id").
		Order("product_count DESC").
		Order("brands.name ASC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

type homeCategoryRow struct {
	models.Category
	ProductCount int64 `gorm:"column:product_count"`
}

// ListHomeCategories returns active top-level categories ranked by catalog size.
func (r *StorefrontRepository) ListHomeCategories(ctx context.Context, limit int) ([]homeCategoryRow, error) {
	var rows []homeCategoryRow
	err := r.db.WithContext(ctx).Table("categories").
		Select("categories.*, COUNT(products.id) AS product_count").
		Joins("LEFT JOIN products ON products.category_id = categories.id AND products.deleted_at IS NULL AND products.status = ?", constants.ProductStatusActive).
		Where("categories.is_active = ?", true).
		Where("categories.deleted_at IS NULL").
		Where("categories.parent_id IS NULL").
		Group("categories.id").
		Order("product_count DESC").
		Order("categories.name ASC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

// FindCategoryCoverImages returns the first product image per category id.
func (r *StorefrontRepository) FindCategoryCoverImages(ctx context.Context, categoryIDs []uint) (map[uint]string, error) {
	out := make(map[uint]string, len(categoryIDs))
	if len(categoryIDs) == 0 {
		return out, nil
	}

	var products []models.Product
	err := r.db.WithContext(ctx).
		Select("category_id", "images").
		Where("category_id IN ?", categoryIDs).
		Where("status = ?", constants.ProductStatusActive).
		Where("array_length(images, 1) > 0").
		Order("rating DESC, reviews_count DESC").
		Find(&products).Error
	if err != nil {
		return nil, err
	}

	for _, p := range products {
		if p.CategoryID == nil {
			continue
		}
		if _, exists := out[*p.CategoryID]; exists {
			continue
		}
		if len(p.Images) > 0 && p.Images[0] != "" {
			out[*p.CategoryID] = p.Images[0]
		}
	}
	return out, nil
}

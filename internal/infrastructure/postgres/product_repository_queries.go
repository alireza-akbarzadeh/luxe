package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

var productDetailPreloads = []string{"Category", "Store", "Brand", "Attributes", "WorkflowState"}

var productListPreloads = []string{"Category", "Brand", "Attributes", "WorkflowState"}

func (r *ProductRepository) preloadProduct(q *gorm.DB, names []string) *gorm.DB {
	for _, name := range names {
		q = q.Preload(name)
	}
	return q
}

// GetDetailedByID loads a product with relations for HTTP responses.
func (r *ProductRepository) GetDetailedByID(ctx context.Context, id uint) (*models.Product, error) {
	var product models.Product
	q := r.preloadProduct(r.db.WithContext(ctx), productDetailPreloads)
	if err := q.First(&product, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &product, nil
}

// GetDetailedBySlug loads a product by slug with relations.
func (r *ProductRepository) GetDetailedBySlug(ctx context.Context, slug string) (*models.Product, error) {
	var product models.Product
	q := r.preloadProduct(r.db.WithContext(ctx), productDetailPreloads)
	if err := q.Where("slug = ?", slug).First(&product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &product, nil
}

// ListDetailed returns paginated products with list filters and preloads.
func (r *ProductRepository) ListDetailed(ctx context.Context, limit, offset int, filters dto.ProductListFilters) ([]*models.Product, int64, error) {
	const maxLimit = 100
	if limit <= 0 {
		limit = 10
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	if offset < 0 {
		offset = 0
	}

	query := r.db.WithContext(ctx).Model(&models.Product{}).Order("id DESC")

	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.Name != "" {
		query = query.Where("LOWER(name) LIKE LOWER(?)", "%"+filters.Name+"%")
	}
	if filters.StoreID != nil && *filters.StoreID != 0 {
		query = query.Where("store_id = ?", *filters.StoreID)
	}
	if filters.SKU != "" {
		query = query.Where("sku LIKE ?", "%"+filters.SKU+"%")
	}
	if filters.CategoryID != 0 {
		query = query.Where("category_id = ?", filters.CategoryID)
	}
	if filters.BrandID != nil && *filters.BrandID != 0 {
		query = query.Where("brand_id = ?", *filters.BrandID)
	}
	if filters.MinPrice != 0 {
		query = query.Where("price >= ?", filters.MinPrice)
	}
	if filters.MaxPrice != 0 {
		query = query.Where("price <= ?", filters.MaxPrice)
	}
	if filters.MinRating != 0 {
		query = query.Where("rating >= ?", filters.MinRating)
	}
	if filters.MaxRating != 0 {
		query = query.Where("rating <= ?", filters.MaxRating)
	}
	if filters.MinReviews != 0 {
		query = query.Where("reviews_count >= ?", filters.MinReviews)
	}
	if filters.MaxReviews != 0 {
		query = query.Where("reviews_count <= ?", filters.MaxReviews)
	}
	if filters.IsDigital != nil {
		query = query.Where("is_digital = ?", *filters.IsDigital)
	}
	if filters.IsNew != nil {
		query = query.Where("is_new = ?", *filters.IsNew)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count products: %w", err)
	}

	listQ := r.preloadProduct(query, productListPreloads)
	var products []*models.Product
	if err := listQ.Limit(limit).Offset(offset).Find(&products).Error; err != nil {
		return nil, 0, fmt.Errorf("find products: %w", err)
	}
	return products, total, nil
}

// DeleteByID soft-deletes a product by primary key.
func (r *ProductRepository) DeleteByID(ctx context.Context, id uint) (int64, error) {
	result := r.db.WithContext(ctx).Delete(&models.Product{}, id)
	return result.RowsAffected, result.Error
}

// ExistsByID reports whether a product row exists.
func (r *ProductRepository) ExistsByID(ctx context.Context, id uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Product{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// SKUTaken checks SKU uniqueness excluding an optional product id.
func (r *ProductRepository) SKUTaken(ctx context.Context, sku string, excludeID uint) (bool, error) {
	q := r.db.WithContext(ctx).Model(&models.Product{}).Where("sku = ?", sku)
	if excludeID > 0 {
		q = q.Where("id != ?", excludeID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

var productSuggestionPreloads = []string{"Category", "Brand"}

// FindLowStockActive returns active products at or below their low-stock threshold.
func (r *ProductRepository) FindLowStockActive(ctx context.Context, activeStatus string) ([]models.Product, error) {
	var products []models.Product
	err := r.db.WithContext(ctx).
		Where("stock <= low_stock_threshold AND status = ?", activeStatus).
		Find(&products).Error
	return products, err
}

// GetRelated returns products in the same category as the given product.
func (r *ProductRepository) GetRelated(ctx context.Context, productID uint, limit int) ([]*models.Product, error) {
	var product models.Product
	if err := r.db.WithContext(ctx).First(&product, productID).Error; err != nil {
		return nil, err
	}
	if product.CategoryID == nil {
		return []*models.Product{}, nil
	}

	var related []*models.Product
	q := r.preloadProduct(r.db.WithContext(ctx), productSuggestionPreloads)
	err := q.
		Where("category_id = ? AND id != ?", *product.CategoryID, productID).
		Order("rating DESC, reviews_count DESC").
		Limit(limit).
		Find(&related).Error
	return related, err
}

// GetSuggestions returns products from categories of the given cart product ids.
func (r *ProductRepository) GetSuggestions(ctx context.Context, productIDs []uint, limit int) ([]*models.Product, error) {
	if limit <= 0 {
		limit = 4
	}
	if len(productIDs) == 0 {
		return []*models.Product{}, nil
	}

	var categoryIDs []uint
	if err := r.db.WithContext(ctx).Model(&models.Product{}).
		Where("id IN ?", productIDs).
		Distinct("category_id").
		Pluck("category_id", &categoryIDs).Error; err != nil {
		return nil, err
	}
	if len(categoryIDs) == 0 {
		return []*models.Product{}, nil
	}

	var suggestions []*models.Product
	q := r.preloadProduct(r.db.WithContext(ctx), productSuggestionPreloads)
	err := q.
		Where("category_id IN ? AND id NOT IN ?", categoryIDs, productIDs).
		Order("rating DESC, reviews_count DESC, created_at DESC").
		Limit(limit).
		Find(&suggestions).Error
	return suggestions, err
}

// FindCategoryByID loads a category row for search document building.
func (r *ProductRepository) FindCategoryByID(ctx context.Context, id uint) (*models.Category, error) {
	var cat models.Category
	if err := r.db.WithContext(ctx).Select("id", "name", "name_i18n", "slug").First(&cat, id).Error; err != nil {
		return nil, err
	}
	return &cat, nil
}

package postgres

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// BrandListItem is a brand row with aggregated product count for admin lists.
type BrandListItem struct {
	Brand        models.Brand
	ProductCount int64
}

// BrandRepository persists brands with GORM.
type BrandRepository struct {
	db *gorm.DB
}

// NewBrandRepository creates a GORM-backed brand repository.
func NewBrandRepository(db *gorm.DB) *BrandRepository {
	return &BrandRepository{db: db}
}

// GetByID loads a brand with workflow state.
func (r *BrandRepository) GetByID(ctx context.Context, id uint) (*models.Brand, error) {
	var brand models.Brand
	if err := r.db.WithContext(ctx).Preload("WorkflowState").First(&brand, id).Error; err != nil {
		return nil, err
	}
	return &brand, nil
}

// GetBySlug loads a brand by URL slug with workflow state.
func (r *BrandRepository) GetBySlug(ctx context.Context, slug string) (*models.Brand, error) {
	var brand models.Brand
	if err := r.db.WithContext(ctx).Preload("WorkflowState").Where("slug = ?", slug).First(&brand).Error; err != nil {
		return nil, err
	}
	return &brand, nil
}

// CountProducts returns the number of non-deleted products for a brand.
func (r *BrandRepository) CountProducts(ctx context.Context, brandID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Product{}).
		Where("brand_id = ? AND deleted_at IS NULL", brandID).
		Count(&count).Error
	return count, err
}

// Create inserts a brand row.
func (r *BrandRepository) Create(ctx context.Context, brand *models.Brand) error {
	return r.db.WithContext(ctx).Create(brand).Error
}

// Save persists brand field changes.
func (r *BrandRepository) Save(ctx context.Context, brand *models.Brand) error {
	return r.db.WithContext(ctx).Save(brand).Error
}

// DeleteByID removes a brand.
func (r *BrandRepository) DeleteByID(ctx context.Context, id uint) (int64, error) {
	result := r.db.WithContext(ctx).Delete(&models.Brand{}, id)
	return result.RowsAffected, result.Error
}

// List returns paginated brands matching filters with product counts.
func (r *BrandRepository) List(ctx context.Context, req *dto.ListBrandsRequest) ([]BrandListItem, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.Brand{})

	if req.Search != "" {
		search := "%" + req.Search + "%"
		query = query.Where("brands.name ILIKE ? OR brands.slug ILIKE ?", search, search)
	}
	if req.Status != "" {
		query = query.Where("brands.status = ?", req.Status)
	}
	if req.Featured != nil {
		query = query.Where("brands.is_featured = ?", *req.Featured)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (req.Page - 1) * req.Limit
	type listRow struct {
		models.Brand
		ProductCount int64 `gorm:"column:product_count"`
	}

	orderBy := "brands.created_at DESC"
	switch req.Sort {
	case "popular":
		orderBy = "product_count DESC, brands.name ASC"
	case "name_asc":
		orderBy = "brands.name ASC"
	case "featured":
		orderBy = "brands.is_featured DESC, brands.featured_sort_order ASC, brands.name ASC"
	}

	var rows []listRow
	if err := query.
		Select(`brands.*, COUNT(products.id) AS product_count`).
		Joins("LEFT JOIN products ON products.brand_id = brands.id AND products.deleted_at IS NULL").
		Group("brands.id").
		Offset(offset).Limit(req.Limit).
		Order(orderBy).
		Preload("WorkflowState").
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}

	items := make([]BrandListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, BrandListItem{
			Brand:        row.Brand,
			ProductCount: row.ProductCount,
		})
	}
	return items, total, nil
}

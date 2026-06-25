package postgres

import (
	"context"
	"errors"

	"github.com/alireza-akbarzadeh/luxe/internal/domain/catalog"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// ProductRepository implements catalog.ProductRepository with GORM.
type ProductRepository struct {
	db *gorm.DB
}

// NewProductRepository creates a GORM-backed product repository.
func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) GetByID(ctx context.Context, id uint) (*catalog.Product, error) {
	var m models.Product
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, catalog.ErrProductNotFound
		}
		return nil, err
	}
	return toDomainProduct(&m), nil
}

func (r *ProductRepository) GetBySlug(ctx context.Context, slug string) (*catalog.Product, error) {
	var m models.Product
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, catalog.ErrProductNotFound
		}
		return nil, err
	}
	return toDomainProduct(&m), nil
}

func (r *ProductRepository) List(ctx context.Context, filter catalog.ListFilter, limit, offset int) ([]catalog.Product, int64, error) {
	q := r.db.WithContext(ctx).Model(&models.Product{})
	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		q = q.Where("name ILIKE ? OR sku ILIKE ?", like, like)
	}
	if filter.CategoryID != nil {
		q = q.Where("category_id = ?", *filter.CategoryID)
	}
	if filter.BrandID != nil {
		q = q.Where("brand_id = ?", *filter.BrandID)
	}
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []models.Product
	if err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]catalog.Product, 0, len(rows))
	for i := range rows {
		out = append(out, *toDomainProduct(&rows[i]))
	}
	return out, total, nil
}

func (r *ProductRepository) SlugExists(ctx context.Context, slug string, excludeID uint) (bool, error) {
	q := r.db.WithContext(ctx).Model(&models.Product{}).Where("slug = ?", slug)
	if excludeID > 0 {
		q = q.Where("id != ?", excludeID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func toDomainProduct(m *models.Product) *catalog.Product {
	return &catalog.Product{
		ID:          m.ID,
		Name:        m.Name,
		Slug:        m.Slug,
		Description: m.Description,
		PriceCents:  int64(m.Price * 100),
		Currency:    "USD",
		SKU:         m.SKU,
		Stock:       m.Stock,
		Status:      m.Status,
		CategoryID:  m.CategoryID,
		BrandID:     m.BrandID,
	}
}

package postgres

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

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

// List returns paginated brands matching filters.
func (r *BrandRepository) List(ctx context.Context, req *dto.ListBrandsRequest) ([]models.Brand, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.Brand{})

	if req.Search != "" {
		search := "%" + req.Search + "%"
		query = query.Where("name ILIKE ? OR slug ILIKE ?", search, search)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (req.Page - 1) * req.Limit
	var brands []models.Brand
	if err := query.Offset(offset).Limit(req.Limit).
		Order("created_at DESC").
		Preload("WorkflowState").
		Find(&brands).Error; err != nil {
		return nil, 0, err
	}
	return brands, total, nil
}

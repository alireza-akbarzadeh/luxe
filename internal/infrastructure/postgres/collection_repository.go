package postgres

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// CollectionRepository persists collections with GORM.
type CollectionRepository struct {
	db *gorm.DB
}

// NewCollectionRepository creates a GORM-backed collection repository.
func NewCollectionRepository(db *gorm.DB) *CollectionRepository {
	return &CollectionRepository{db: db}
}

// SlugExists reports whether a slug is taken, optionally excluding an id.
func (r *CollectionRepository) SlugExists(ctx context.Context, slug string, excludeID uint) (bool, error) {
	q := r.db.WithContext(ctx).Model(&models.Collection{}).Where("slug = ?", slug)
	if excludeID > 0 {
		q = q.Where("id != ?", excludeID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetByID loads a collection with workflow state.
func (r *CollectionRepository) GetByID(ctx context.Context, id uint) (*models.Collection, error) {
	var collection models.Collection
	if err := r.db.WithContext(ctx).Preload("WorkflowState").First(&collection, id).Error; err != nil {
		return nil, err
	}
	return &collection, nil
}

// Create inserts a collection row.
func (r *CollectionRepository) Create(ctx context.Context, collection *models.Collection) error {
	return r.db.WithContext(ctx).Create(collection).Error
}

// Save persists collection field changes.
func (r *CollectionRepository) Save(ctx context.Context, collection *models.Collection) error {
	return r.db.WithContext(ctx).Save(collection).Error
}

// DeleteByID removes a collection.
func (r *CollectionRepository) DeleteByID(ctx context.Context, id uint) (int64, error) {
	result := r.db.WithContext(ctx).Delete(&models.Collection{}, id)
	return result.RowsAffected, result.Error
}

// List returns paginated collections matching filters.
func (r *CollectionRepository) List(ctx context.Context, req *dto.ListCollectionsRequest, page, limit int) ([]models.Collection, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.Collection{})

	if req.Search != "" {
		search := "%" + req.Search + "%"
		query = query.Where("title ILIKE ? OR slug ILIKE ? OR eyebrow ILIKE ?", search, search, search)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	var collections []models.Collection
	if err := query.Order("sort_order ASC, created_at DESC").
		Offset(offset).Limit(limit).
		Preload("WorkflowState").
		Find(&collections).Error; err != nil {
		return nil, 0, err
	}
	return collections, total, nil
}

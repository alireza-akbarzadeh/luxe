package postgres

import (
	"context"
	"errors"
	"time"

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

func (r *CollectionRepository) preloadCollectionProducts(db *gorm.DB) *gorm.DB {
	return db.Preload("Products", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("position ASC, sort_order ASC, id ASC").Preload("Product").Preload("Product.Category").
			Preload("Product.Brand").Preload("Product.Attributes").Preload("Product.WorkflowState")
	})
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

// GetByID loads a collection with workflow state and manual products.
func (r *CollectionRepository) GetByID(ctx context.Context, id uint) (*models.Collection, error) {
	var collection models.Collection
	q := r.preloadCollectionProducts(r.db.WithContext(ctx).Preload("WorkflowState"))
	if err := q.First(&collection, id).Error; err != nil {
		return nil, err
	}
	return &collection, nil
}

// GetBySlug loads a collection by current slug.
func (r *CollectionRepository) GetBySlug(ctx context.Context, slug string) (*models.Collection, error) {
	var collection models.Collection
	q := r.preloadCollectionProducts(r.db.WithContext(ctx).Preload("WorkflowState"))
	if err := q.Where("slug = ?", slug).First(&collection).Error; err != nil {
		return nil, err
	}
	return &collection, nil
}

// GetByRedirectSlug resolves a collection by an old slug redirect.
func (r *CollectionRepository) GetByRedirectSlug(ctx context.Context, slug string) (*models.Collection, error) {
	var redirect models.CollectionSlugRedirect
	err := r.db.WithContext(ctx).
		Preload("Collection").
		Where("old_slug = ?", slug).
		First(&redirect).Error
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, redirect.CollectionID)
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

// ReplaceProducts replaces manual collection membership in sort order.
func (r *CollectionRepository) ReplaceProducts(ctx context.Context, collectionID uint, productIDs []uint, overrides []dto.CollectionProductOverrideInput) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("collection_id = ?", collectionID).Delete(&models.CollectionProduct{}).Error; err != nil {
			return err
		}
		if len(productIDs) == 0 {
			return nil
		}
		overrideByProductID := make(map[uint]dto.CollectionProductOverrideInput, len(overrides))
		for _, override := range overrides {
			overrideByProductID[override.ProductID] = override
		}
		rows := make([]models.CollectionProduct, len(productIDs))
		for i, productID := range productIDs {
			override, hasOverride := overrideByProductID[productID]
			rows[i] = models.CollectionProduct{
				CollectionID: collectionID,
				ProductID:    productID,
				SortOrder:    i,
				Position:     i,
			}
			if hasOverride {
				rows[i].Position = override.Position
				rows[i].IsPinned = override.IsPinned
				rows[i].IsHidden = override.IsHidden
				rows[i].BoostScore = override.BoostScore
			}
		}
		return tx.Create(&rows).Error
	})
}

// ClearProducts removes all manual products from a collection.
func (r *CollectionRepository) ClearProducts(ctx context.Context, collectionID uint) error {
	return r.db.WithContext(ctx).Where("collection_id = ?", collectionID).Delete(&models.CollectionProduct{}).Error
}

// UpsertSlugRedirect creates a permanent redirect for a previous collection slug.
func (r *CollectionRepository) UpsertSlugRedirect(ctx context.Context, collectionID uint, oldSlug string) error {
	if oldSlug == "" {
		return nil
	}
	var existing models.CollectionSlugRedirect
	err := r.db.WithContext(ctx).Where("old_slug = ?", oldSlug).First(&existing).Error
	if err == nil {
		existing.CollectionID = collectionID
		return r.db.WithContext(ctx).Save(&existing).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return r.db.WithContext(ctx).Create(&models.CollectionSlugRedirect{
		CollectionID: collectionID,
		OldSlug:      oldSlug,
	}).Error
}

func applyCollectionListFilters(query *gorm.DB, req *dto.ListCollectionsRequest) *gorm.DB {
	if req.Search != "" {
		search := "%" + req.Search + "%"
		query = query.Where("title ILIKE ? OR slug ILIKE ? OR eyebrow ILIKE ?", search, search, search)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.Theme != "" {
		query = query.Where("theme = ?", req.Theme)
	}
	if req.CollectionType != "" {
		query = query.Where("collection_type = ?", req.CollectionType)
	}
	if req.Mode != "" {
		query = query.Where("mode = ?", req.Mode)
	}
	if req.LiveOnly {
		now := time.Now()
		query = query.Where("(starts_at IS NULL OR starts_at <= ?)", now).
			Where("(ends_at IS NULL OR ends_at >= ?)", now)
	}
	return query
}

// List returns paginated collections matching filters.
func (r *CollectionRepository) List(ctx context.Context, req *dto.ListCollectionsRequest, page, limit int) ([]models.Collection, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.Collection{})
	query = applyCollectionListFilters(query, req)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	var collections []models.Collection
	q := r.preloadCollectionProducts(query.Order("sort_order ASC, created_at DESC").
		Offset(offset).Limit(limit).
		Preload("WorkflowState"))
	if err := q.Find(&collections).Error; err != nil {
		return nil, 0, err
	}
	return collections, total, nil
}

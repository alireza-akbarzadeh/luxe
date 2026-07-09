package postgres

import (
	"context"
	"fmt"

	"github.com/alireza-akbarzadeh/luxe/internal/i18n"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

var categoryDetailPreloads = []string{"Parent", "WorkflowState"}

func childrenPreload(db *gorm.DB) *gorm.DB {
	return db.Order("sort_order ASC, id ASC")
}

// CategoryRepository persists categories with GORM.
type CategoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository creates a GORM-backed category repository.
func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) preload(q *gorm.DB, names []string) *gorm.DB {
	for _, name := range names {
		if name == "Children" {
			q = q.Preload("Children", childrenPreload)
			continue
		}
		q = q.Preload(name)
	}
	return q
}

// SlugExists reports whether a slug is taken, optionally excluding an id.
func (r *CategoryRepository) SlugExists(ctx context.Context, slug string, excludeID uint) (bool, error) {
	q := r.db.WithContext(ctx).Model(&models.Category{}).Where("slug = ?", slug)
	if excludeID > 0 {
		q = q.Where("id != ?", excludeID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// FindParent loads a parent category by id.
func (r *CategoryRepository) FindParent(ctx context.Context, parentID uint) (*models.Category, error) {
	var parent models.Category
	if err := r.db.WithContext(ctx).First(&parent, parentID).Error; err != nil {
		return nil, err
	}
	return &parent, nil
}

// Create inserts a category row.
func (r *CategoryRepository) Create(ctx context.Context, category *models.Category) error {
	return r.db.WithContext(ctx).Create(category).Error
}

// Save persists category field changes.
func (r *CategoryRepository) Save(ctx context.Context, category *models.Category) error {
	return r.db.WithContext(ctx).Save(category).Error
}

// GetByID loads a category with relations.
func (r *CategoryRepository) GetByID(ctx context.Context, id uint) (*models.Category, error) {
	var category models.Category
	q := r.preload(r.db.WithContext(ctx), append(categoryDetailPreloads, "Children"))
	if err := q.First(&category, id).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

// GetBySlug loads a category by slug with relations.
func (r *CategoryRepository) GetBySlug(ctx context.Context, slug string) (*models.Category, error) {
	var category models.Category
	q := r.preload(r.db.WithContext(ctx), append(categoryDetailPreloads, "Children"))
	if err := q.Where("slug = ?", slug).First(&category).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

// CountChildren counts direct child categories.
func (r *CategoryRepository) CountChildren(ctx context.Context, parentID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Category{}).Where("parent_id = ?", parentID).Count(&count).Error
	return count, err
}

// CountChildrenIn counts categories whose parent is in the given ids.
func (r *CategoryRepository) CountChildrenIn(ctx context.Context, ids []uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Category{}).Where("parent_id IN ?", ids).Count(&count).Error
	return count, err
}

// DeleteByID soft-deletes a category.
func (r *CategoryRepository) DeleteByID(ctx context.Context, id uint) (int64, error) {
	result := r.db.WithContext(ctx).Delete(&models.Category{}, id)
	return result.RowsAffected, result.Error
}

// BulkDeleteByIDs soft-deletes categories by id.
func (r *CategoryRepository) BulkDeleteByIDs(ctx context.Context, ids []uint) (int64, error) {
	result := r.db.WithContext(ctx).Where("id IN ?", ids).Delete(&models.Category{})
	return result.RowsAffected, result.Error
}

// List returns filtered categories with total count.
func (r *CategoryRepository) List(ctx context.Context, filters dto.CategoryListFilters) ([]models.Category, int64, error) {
	baseQuery := r.db.WithContext(ctx).Model(&models.Category{})

	if filters.IsActive != nil {
		baseQuery = baseQuery.Where("is_active = ?", *filters.IsActive)
	}
	if filters.ParentID != nil {
		baseQuery = baseQuery.Where("parent_id = ?", *filters.ParentID)
	}
	if filters.Search != "" {
		normalized := i18n.NormalizeSearchQuery(filters.Search)
		like := "%" + normalized + "%"
		baseQuery = baseQuery.Where("search_document ILIKE ? OR LOWER(name) LIKE LOWER(?)", like, like)
	}

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	fetchQuery := baseQuery
	switch filters.Sort {
	case "popular":
		fetchQuery = fetchQuery.Select("categories.*, COUNT(products.id) as product_count").
			Joins("LEFT JOIN products ON products.category_id = categories.id").
			Group("categories.id").
			Order("product_count DESC")
	case "name":
		fetchQuery = fetchQuery.Order("name ASC")
	default:
		fetchQuery = fetchQuery.Order("sort_order ASC, id ASC")
	}

	var categories []models.Category
	err := r.preload(fetchQuery, append(categoryDetailPreloads, "Children")).
		Limit(filters.Limit).Offset(filters.Offset).
		Find(&categories).Error
	return categories, total, err
}

// BulkCreate inserts categories in a transaction.
func (r *CategoryRepository) BulkCreate(ctx context.Context, categories []*models.Category) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, cat := range categories {
			if err := tx.Create(cat).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// FindProductsByCategoryID loads products in a category for search document refresh.
func (r *CategoryRepository) FindProductsByCategoryID(ctx context.Context, categoryID uint) ([]models.Product, error) {
	var products []models.Product
	err := r.db.WithContext(ctx).Where("category_id = ?", categoryID).Find(&products).Error
	return products, err
}

// UpdateProductSearchDocument updates a product search document column.
func (r *CategoryRepository) UpdateProductSearchDocument(ctx context.Context, productID uint, doc string) error {
	return r.db.WithContext(ctx).Model(&models.Product{}).Where("id = ?", productID).Update("search_document", doc).Error
}

// FindDescendantIDs returns the category id and all descendant ids.
func (r *CategoryRepository) FindDescendantIDs(ctx context.Context, cat *models.Category) ([]uint, error) {
	prefix := descendantPathPrefix(cat)
	var ids []uint
	err := r.db.WithContext(ctx).Model(&models.Category{}).
		Where("id = ? OR path = ? OR path LIKE ?", cat.ID, prefix, prefix+".%").
		Pluck("id", &ids).Error
	return ids, err
}

func descendantPathPrefix(cat *models.Category) string {
	if cat.Path == "" {
		return fmt.Sprintf("%d", cat.ID)
	}
	return fmt.Sprintf("%s.%d", cat.Path, cat.ID)
}

// FindByIDs loads categories by primary keys.
func (r *CategoryRepository) FindByIDs(ctx context.Context, ids []uint) ([]models.Category, error) {
	var categories []models.Category
	if len(ids) == 0 {
		return categories, nil
	}
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&categories).Error
	return categories, err
}

// ListChildren loads direct children of a parent ordered for display.
func (r *CategoryRepository) ListChildren(ctx context.Context, parentID *uint) ([]models.Category, error) {
	var categories []models.Category
	q := r.db.WithContext(ctx).Model(&models.Category{})
	if parentID == nil {
		q = q.Where("parent_id IS NULL")
	} else {
		q = q.Where("parent_id = ?", *parentID)
	}
	err := q.Order("sort_order ASC, id ASC").Find(&categories).Error
	return categories, err
}

// Reorder persists parent, sort order, and hierarchy fields in one transaction.
func (r *CategoryRepository) Reorder(ctx context.Context, categories []*models.Category) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, category := range categories {
			result := tx.Model(&models.Category{}).Where("id = ?", category.ID).Updates(map[string]any{
				"parent_id":  category.ParentID,
				"sort_order": category.SortOrder,
				"level":      category.Level,
				"path":       category.Path,
			})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return gorm.ErrRecordNotFound
			}
		}
		return nil
	})
}

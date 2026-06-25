package postgres

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// MenuRepository persists menu groups and items with GORM.
type MenuRepository struct {
	db *gorm.DB
}

// NewMenuRepository creates a GORM-backed menu repository.
func NewMenuRepository(db *gorm.DB) *MenuRepository {
	return &MenuRepository{db: db}
}

// ListGroups returns all menu groups ordered by display_order.
func (r *MenuRepository) ListGroups(ctx context.Context) ([]models.MenuGroup, error) {
	var groups []models.MenuGroup
	err := r.db.WithContext(ctx).Order("display_order ASC").Find(&groups).Error
	return groups, err
}

// FindGroupByID loads a menu group by id.
func (r *MenuRepository) FindGroupByID(ctx context.Context, id uint) (*models.MenuGroup, error) {
	var group models.MenuGroup
	if err := r.db.WithContext(ctx).First(&group, id).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

// CreateGroup inserts a menu group.
func (r *MenuRepository) CreateGroup(ctx context.Context, group *models.MenuGroup) error {
	return r.db.WithContext(ctx).Create(group).Error
}

// SaveGroup persists menu group changes.
func (r *MenuRepository) SaveGroup(ctx context.Context, group *models.MenuGroup) error {
	return r.db.WithContext(ctx).Save(group).Error
}

// DeleteGroup removes a menu group.
func (r *MenuRepository) DeleteGroup(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.MenuGroup{}, id).Error
}

// ListAllItems returns all menu items ordered by display_order.
func (r *MenuRepository) ListAllItems(ctx context.Context) ([]models.MenuItem, error) {
	var items []models.MenuItem
	err := r.db.WithContext(ctx).Order("display_order ASC").Find(&items).Error
	return items, err
}

// ListGroupsWithItems returns groups with preloaded items.
func (r *MenuRepository) ListGroupsWithItems(ctx context.Context) ([]models.MenuGroup, error) {
	var groups []models.MenuGroup
	err := r.db.WithContext(ctx).Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Order("display_order ASC")
	}).Order("display_order ASC").Find(&groups).Error
	return groups, err
}

// FindItemByID loads a menu item by id.
func (r *MenuRepository) FindItemByID(ctx context.Context, id uint) (*models.MenuItem, error) {
	var item models.MenuItem
	if err := r.db.WithContext(ctx).First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

// FindGroupExists checks that a menu group exists.
func (r *MenuRepository) FindGroupExists(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).First(&models.MenuGroup{}, id).Error
}

// FindItemExists checks that a menu item exists.
func (r *MenuRepository) FindItemExists(ctx context.Context, id uint) (*models.MenuItem, error) {
	var item models.MenuItem
	if err := r.db.WithContext(ctx).First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

// CreateItem inserts a menu item.
func (r *MenuRepository) CreateItem(ctx context.Context, item *models.MenuItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

// SaveItem persists menu item changes.
func (r *MenuRepository) SaveItem(ctx context.Context, item *models.MenuItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

// DeleteItem removes a menu item.
func (r *MenuRepository) DeleteItem(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.MenuItem{}, id).Error
}

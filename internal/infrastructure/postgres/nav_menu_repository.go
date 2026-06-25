package postgres

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// NavMenuRepository persists navigation menus with GORM.
type NavMenuRepository struct {
	db *gorm.DB
}

// NewNavMenuRepository creates a GORM-backed nav menu repository.
func NewNavMenuRepository(db *gorm.DB) *NavMenuRepository {
	return &NavMenuRepository{db: db}
}

// ListAll returns nav menus ordered by sort order.
func (r *NavMenuRepository) ListAll(ctx context.Context) ([]models.NavMenu, error) {
	var menus []models.NavMenu
	err := r.db.WithContext(ctx).Order("nav_menus.\"order\" ASC").Find(&menus).Error
	return menus, err
}

// FindByID loads a nav menu by id.
func (r *NavMenuRepository) FindByID(ctx context.Context, id uint) (*models.NavMenu, error) {
	var menu models.NavMenu
	err := r.db.WithContext(ctx).First(&menu, id).Error
	if err != nil {
		return nil, err
	}
	return &menu, nil
}

// Create inserts a nav menu row.
func (r *NavMenuRepository) Create(ctx context.Context, menu *models.NavMenu) error {
	return r.db.WithContext(ctx).Create(menu).Error
}

// Save persists nav menu changes.
func (r *NavMenuRepository) Save(ctx context.Context, menu *models.NavMenu) error {
	return r.db.WithContext(ctx).Save(menu).Error
}

// Reorder updates sort order for multiple menus in a transaction.
func (r *NavMenuRepository) Reorder(ctx context.Context, items []struct {
	ID    uint
	Order int
}) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			result := tx.Model(&models.NavMenu{}).
				Where("id = ?", item.ID).
				Update("order", item.Order)
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

// DeleteByID removes a nav menu by id.
func (r *NavMenuRepository) DeleteByID(ctx context.Context, id uint) (int64, error) {
	result := r.db.WithContext(ctx).Delete(&models.NavMenu{}, id)
	return result.RowsAffected, result.Error
}

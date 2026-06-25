package postgres

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// CompareRepository persists compare lists and product lookups with GORM.
type CompareRepository struct {
	db *gorm.DB
}

// NewCompareRepository creates a GORM-backed compare repository.
func NewCompareRepository(db *gorm.DB) *CompareRepository {
	return &CompareRepository{db: db}
}

// FindProductsByIDs loads products with category and store preloaded.
func (r *CompareRepository) FindProductsByIDs(ctx context.Context, productIDs []uint) ([]models.Product, error) {
	var products []models.Product
	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Store").
		Where("id IN ?", productIDs).
		Find(&products).Error
	return products, err
}

// FindCompareListByUser loads a user's compare list.
func (r *CompareRepository) FindCompareListByUser(userID uint) (*models.CompareList, error) {
	var list models.CompareList
	err := r.db.Where("user_id = ?", userID).First(&list).Error
	if err != nil {
		return nil, err
	}
	return &list, nil
}

// CreateCompareList inserts a new compare list.
func (r *CompareRepository) CreateCompareList(list *models.CompareList) error {
	return r.db.Create(list).Error
}

// SaveCompareList persists compare list changes.
func (r *CompareRepository) SaveCompareList(list *models.CompareList) error {
	return r.db.Save(list).Error
}

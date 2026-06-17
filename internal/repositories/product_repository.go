package repositories

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// ProductRepository handles product reads needed by cart and checkout flows.
type ProductRepository interface {
	GetByID(ctx context.Context, id uint) (*models.Product, error)
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) GetByID(ctx context.Context, id uint) (*models.Product, error) {
	var product models.Product
	err := r.db.WithContext(ctx).First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

package postgres

import (
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// UserLikeRepository persists product likes with GORM.
type UserLikeRepository struct {
	db *gorm.DB
}

// NewUserLikeRepository creates a GORM-backed user-like repository.
func NewUserLikeRepository(db *gorm.DB) *UserLikeRepository {
	return &UserLikeRepository{db: db}
}

// FindProductByID loads a product by id.
func (r *UserLikeRepository) FindProductByID(productID uint) (*models.Product, error) {
	var product models.Product
	err := r.db.First(&product, productID).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// CreateLike inserts a product like row.
func (r *UserLikeRepository) CreateLike(like *models.ProductLike) error {
	return r.db.Create(like).Error
}

// DeleteLike removes a product like row.
func (r *UserLikeRepository) DeleteLike(userID, productID uint) error {
	return r.db.Where("user_id = ? AND product_id = ?", userID, productID).Delete(&models.ProductLike{}).Error
}

// CountLike checks if a user liked a product.
func (r *UserLikeRepository) CountLike(userID, productID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.ProductLike{}).
		Where("user_id = ? AND product_id = ?", userID, productID).
		Count(&count).Error
	return count, err
}

// PluckLikedProductIDs returns product ids liked by a user.
func (r *UserLikeRepository) PluckLikedProductIDs(userID uint) ([]uint, error) {
	var ids []uint
	err := r.db.Model(&models.ProductLike{}).Where("user_id = ?", userID).Pluck("product_id", &ids).Error
	return ids, err
}

// CountWishlistProducts counts products in a user's wishlist.
func (r *UserLikeRepository) CountWishlistProducts(userID uint) (int64, error) {
	subQuery := r.db.Model(&models.ProductLike{}).
		Where("user_id = ?", userID).
		Select("product_id")

	var total int64
	err := r.db.Model(&models.Product{}).Where("id IN (?)", subQuery).Count(&total).Error
	return total, err
}

// FindWishlistProducts returns paginated wishlist products for a user.
func (r *UserLikeRepository) FindWishlistProducts(userID uint, limit, offset int, orderByClause string) ([]models.Product, error) {
	subQuery := r.db.Model(&models.ProductLike{}).
		Where("user_id = ?", userID).
		Select("product_id")

	var products []models.Product
	err := r.db.Where("id IN (?)", subQuery).
		Order(orderByClause).
		Limit(limit).
		Offset(offset).
		Find(&products).Error
	return products, err
}

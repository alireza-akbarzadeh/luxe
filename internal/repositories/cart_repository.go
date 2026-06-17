package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// CartRepository handles cart persistence.
type CartRepository interface {
	FindActiveByUserID(ctx context.Context, userID uint, preloadItems bool) (*models.Cart, error)
	Create(ctx context.Context, cart *models.Cart) error
	UpdateStatus(ctx context.Context, cartID uint, status string) error
	FindCartItemByCartAndProduct(ctx context.Context, cartID, productID uint) (*models.CartItem, error)
	FindCartItemForUser(ctx context.Context, userID, cartItemID uint) (*models.CartItem, error)
	CreateCartItem(ctx context.Context, item *models.CartItem) error
	SaveCartItem(ctx context.Context, item *models.CartItem) error
	DeleteCartItem(ctx context.Context, userID, cartItemID uint) (int64, error)
	DeleteItemsForActiveCart(ctx context.Context, userID uint) error
	FindStaleActiveCarts(ctx context.Context, cutoff time.Time) ([]models.Cart, error)
}

type cartRepository struct {
	db *gorm.DB
}

func NewCartRepository(db *gorm.DB) CartRepository {
	return &cartRepository{db: db}
}

func (r *cartRepository) FindActiveByUserID(ctx context.Context, userID uint, preloadItems bool) (*models.Cart, error) {
	q := r.db.WithContext(ctx).Where("user_id = ? AND status = ?", userID, constants.CartStatusActive)
	if preloadItems {
		q = q.Preload("Items.Product")
	}
	var cart models.Cart
	err := q.First(&cart).Error
	if err != nil {
		return nil, err
	}
	return &cart, nil
}

func (r *cartRepository) Create(ctx context.Context, cart *models.Cart) error {
	return r.db.WithContext(ctx).Create(cart).Error
}

func (r *cartRepository) UpdateStatus(ctx context.Context, cartID uint, status string) error {
	return r.db.WithContext(ctx).Model(&models.Cart{}).Where("id = ?", cartID).Update("status", status).Error
}

func (r *cartRepository) FindCartItemByCartAndProduct(ctx context.Context, cartID, productID uint) (*models.CartItem, error) {
	var item models.CartItem
	err := r.db.WithContext(ctx).Where("cart_id = ? AND product_id = ?", cartID, productID).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *cartRepository) FindCartItemForUser(ctx context.Context, userID, cartItemID uint) (*models.CartItem, error) {
	var item models.CartItem
	err := r.db.WithContext(ctx).
		Joins("JOIN carts ON carts.id = cart_items.cart_id").
		Where("cart_items.id = ? AND carts.user_id = ? AND carts.status = ?", cartItemID, userID, constants.CartStatusActive).
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *cartRepository) CreateCartItem(ctx context.Context, item *models.CartItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *cartRepository) SaveCartItem(ctx context.Context, item *models.CartItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *cartRepository) DeleteCartItem(ctx context.Context, userID, cartItemID uint) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("id = ? AND cart_id IN (SELECT id FROM carts WHERE user_id = ? AND status = ?)",
			cartItemID, userID, constants.CartStatusActive).
		Delete(&models.CartItem{})
	return result.RowsAffected, result.Error
}

func (r *cartRepository) DeleteItemsForActiveCart(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).
		Where("cart_id IN (SELECT id FROM carts WHERE user_id = ? AND status = ?)", userID, constants.CartStatusActive).
		Delete(&models.CartItem{}).Error
}

func (r *cartRepository) FindStaleActiveCarts(ctx context.Context, cutoff time.Time) ([]models.Cart, error) {
	var carts []models.Cart
	err := r.db.WithContext(ctx).
		Where("status = ? AND updated_at < ?", constants.CartStatusActive, cutoff).
		Find(&carts).Error
	return carts, err
}

// IsRecordNotFound reports whether err is a missing-row error from GORM.
func IsRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

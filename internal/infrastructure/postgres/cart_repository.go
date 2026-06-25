package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	domaincart "github.com/alireza-akbarzadeh/luxe/internal/domain/cart"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// CartRepository implements cart persistence with GORM.
type CartRepository struct {
	db *gorm.DB
}

// NewCartRepository creates a GORM-backed cart repository.
func NewCartRepository(db *gorm.DB) *CartRepository {
	return &CartRepository{db: db}
}

// GetActiveByUserID implements domain/cart.Repository for the active cart aggregate view.
func (r *CartRepository) GetActiveByUserID(ctx context.Context, userID uint) (*domaincart.Cart, error) {
	m, err := r.FindActiveCart(ctx, userID, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domaincart.ErrCartNotFound
		}
		return nil, err
	}
	return toDomainCart(m), nil
}

// FindActiveCart loads the user's active cart, optionally preloading items.
func (r *CartRepository) FindActiveCart(ctx context.Context, userID uint, preloadItems bool) (*models.Cart, error) {
	q := r.db.WithContext(ctx).Where("user_id = ? AND status = ?", userID, constants.CartStatusActive)
	if preloadItems {
		q = q.Preload("Items.Product")
	}
	var cart models.Cart
	if err := q.First(&cart).Error; err != nil {
		return nil, err
	}
	return &cart, nil
}

// CreateActiveCart creates a new active cart for a user.
func (r *CartRepository) CreateActiveCart(ctx context.Context, userID uint) (*models.Cart, error) {
	newCart := models.Cart{
		UserID:    userID,
		Status:    constants.CartStatusActive,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := r.db.WithContext(ctx).Create(&newCart).Error; err != nil {
		return nil, err
	}
	return &newCart, nil
}

// FindCartItemByCartAndProduct finds a line item for a cart/product pair.
func (r *CartRepository) FindCartItemByCartAndProduct(ctx context.Context, cartID, productID uint) (*models.CartItem, error) {
	var item models.CartItem
	if err := r.db.WithContext(ctx).Where("cart_id = ? AND product_id = ?", cartID, productID).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

// FindCartItemForUser finds a cart item owned by the user's active cart.
func (r *CartRepository) FindCartItemForUser(ctx context.Context, userID, cartItemID uint) (*models.CartItem, error) {
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

// GetProductByID loads a product row for cart validation.
func (r *CartRepository) GetProductByID(ctx context.Context, id uint) (*models.Product, error) {
	var product models.Product
	if err := r.db.WithContext(ctx).First(&product, id).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

// SaveCartItem persists cart item changes.
func (r *CartRepository) SaveCartItem(ctx context.Context, item *models.CartItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

// CreateCartItem inserts a new cart line item.
func (r *CartRepository) CreateCartItem(ctx context.Context, item *models.CartItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

// RemoveCartItemForUser deletes a cart item from the user's active cart.
func (r *CartRepository) RemoveCartItemForUser(ctx context.Context, userID, cartItemID uint) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("id = ? AND cart_id IN (SELECT id FROM carts WHERE user_id = ? AND status = ?)",
			cartItemID, userID, constants.CartStatusActive).
		Delete(&models.CartItem{})
	return result.RowsAffected, result.Error
}

// ClearActiveCartItems removes all items from the user's active cart.
func (r *CartRepository) ClearActiveCartItems(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).
		Where("cart_id IN (SELECT id FROM carts WHERE user_id = ? AND status = ?)", userID, constants.CartStatusActive).
		Delete(&models.CartItem{}).Error
}

// MarkAbandonedCarts sets status to abandoned for stale active carts.
func (r *CartRepository) MarkAbandonedCarts(ctx context.Context, cutoff time.Time) error {
	return r.db.WithContext(ctx).
		Model(&models.Cart{}).
		Where("status = ? AND updated_at < ?", constants.CartStatusActive, cutoff).
		Update("status", constants.CartStatusAbandoned).Error
}

// MarkConverted sets a cart status to converted after checkout.
func (r *CartRepository) MarkConverted(ctx context.Context, cartID uint) error {
	return r.db.WithContext(ctx).Model(&models.Cart{}).Where("id = ?", cartID).
		Update("status", "converted").Error
}

// MarkConvertedTx sets cart status inside an existing transaction.
func (r *CartRepository) MarkConvertedTx(tx *gorm.DB, cartID uint) error {
	return tx.Model(&models.Cart{}).Where("id = ?", cartID).
		Update("status", "converted").Error
}

func toDomainCart(m *models.Cart) *domaincart.Cart {
	items := make([]domaincart.Item, 0, len(m.Items))
	for _, it := range m.Items {
		items = append(items, domaincart.Item{
			ID:        it.ID,
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
			UnitCents: int64(it.Price * 100),
		})
	}
	userID := m.UserID
	return &domaincart.Cart{
		ID:     m.ID,
		UserID: &userID,
		Items:  items,
	}
}

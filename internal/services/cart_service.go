package services

import (
	"context"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"gorm.io/gorm"
)

type AddItemRequest struct {
	ProductID uint `json:"product_id" validate:"required,gt=0"`
	Quantity  int  `json:"quantity" validate:"required,gt=0"`

	Color string `json:"color"`
	Size  string `json:"size"`
}

type UpdateCartItemRequest struct {
	Quantity int    `json:"quantity" validate:"omitempty,gt=0"`
	Color    string `json:"color"`
	Size     string `json:"size"`
}

type CartServiceInterface interface {
	GetOrCreateCart(ctx context.Context, userID uint) (*models.Cart, error)
	AddItem(ctx context.Context, userID uint, req AddItemRequest) (*models.CartItem, error)
	UpdateCartItem(ctx context.Context, userID uint, cartItemID uint, req UpdateCartItemRequest) error
	RemoveItem(ctx context.Context, userID uint, cartItemID uint) error
	GetCart(ctx context.Context, userID uint) (*models.Cart, error)
	ClearCart(ctx context.Context, userID uint) error
	CleanAbandonedCarts(ctx context.Context) error
}

type cartService struct {
	db *gorm.DB
}

func NewCartService(db *gorm.DB) CartServiceInterface {
	return &cartService{db: db}
}

func (s *cartService) GetOrCreateCart(ctx context.Context, userID uint) (*models.Cart, error) {
	cart, err := s.findActiveCart(ctx, userID, true)
	if err == nil {
		return cart, nil
	}
	if !isRecordNotFound(err) {
		return nil, utils.ErrInternal(err)
	}

	newCart := models.Cart{
		UserID:    userID,
		Status:    constants.CartStatusActive,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.db.WithContext(ctx).Create(&newCart).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}
	return &newCart, nil
}

func (s *cartService) AddItem(ctx context.Context, userID uint, req AddItemRequest) (*models.CartItem, error) {
	if req.Quantity <= 0 {
		return nil, utils.ErrBadRequest("quantity must be positive")
	}

	product, err := s.getProductByID(ctx, req.ProductID)
	if err != nil {
		if isRecordNotFound(err) {
			return nil, utils.ErrNotFound("product not found")
		}
		return nil, utils.ErrInternal(err)
	}
	if product.Status != constants.ProductStatusActive {
		return nil, utils.ErrBadRequest("product is not available")
	}
	if !isProductStockAvailable(*product, req.Quantity) {
		return nil, utils.ErrBadRequest("insufficient stock")
	}

	cart, err := s.GetOrCreateCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	cartItem, err := s.findCartItemByCartAndProduct(ctx, cart.ID, req.ProductID)
	if err == nil {
		newQty := cartItem.Quantity + req.Quantity
		if !isProductStockAvailable(*product, newQty) {
			return nil, utils.ErrBadRequest("insufficient stock for updated quantity")
		}
		cartItem.Quantity = newQty
		if err := s.db.WithContext(ctx).Save(cartItem).Error; err != nil {
			return nil, utils.ErrInternal(err)
		}
		return cartItem, nil
	}
	if !isRecordNotFound(err) {
		return nil, utils.ErrInternal(err)
	}

	newItem := models.CartItem{
		CartID:    cart.ID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
		Price:     product.Price,
	}
	if err := s.db.WithContext(ctx).Create(&newItem).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}
	return &newItem, nil
}

func (s *cartService) UpdateCartItem(ctx context.Context, userID uint, cartItemID uint, req UpdateCartItemRequest) error {
	cartItem, err := s.findCartItemForUser(ctx, userID, cartItemID)
	if err != nil {
		if isRecordNotFound(err) {
			return utils.ErrNotFound("cart item not found")
		}
		return utils.ErrInternal(err)
	}

	if req.Quantity > 0 {
		product, err := s.getProductByID(ctx, cartItem.ProductID)
		if err != nil {
			return utils.ErrInternal(err)
		}
		if !isProductStockAvailable(*product, req.Quantity) {
			return utils.ErrBadRequest("insufficient stock")
		}
		cartItem.Quantity = req.Quantity
	}

	if req.Color != "" || req.Size != "" {
		cartItem.Color = req.Color
		cartItem.Size = req.Size
	}

	if err := s.db.WithContext(ctx).Save(cartItem).Error; err != nil {
		return utils.ErrInternal(err)
	}
	return nil
}

func (s *cartService) RemoveItem(ctx context.Context, userID uint, cartItemID uint) error {
	result := s.db.WithContext(ctx).
		Where("id = ? AND cart_id IN (SELECT id FROM carts WHERE user_id = ? AND status = ?)",
			cartItemID, userID, constants.CartStatusActive).
		Delete(&models.CartItem{})
	if result.Error != nil {
		return utils.ErrInternal(result.Error)
	}
	if result.RowsAffected == 0 {
		return utils.ErrNotFound("cart item not found")
	}
	return nil
}

func (s *cartService) GetCart(ctx context.Context, userID uint) (*models.Cart, error) {
	cart, err := s.findActiveCart(ctx, userID, true)
	if err != nil {
		if isRecordNotFound(err) {
			return &models.Cart{UserID: userID, Items: []models.CartItem{}}, nil
		}
		return nil, utils.ErrInternal(err)
	}
	return cart, nil
}

func (s *cartService) ClearCart(ctx context.Context, userID uint) error {
	err := s.db.WithContext(ctx).
		Where("cart_id IN (SELECT id FROM carts WHERE user_id = ? AND status = ?)", userID, constants.CartStatusActive).
		Delete(&models.CartItem{}).Error
	if err != nil {
		return utils.ErrInternal(err)
	}
	return nil
}

func (s *cartService) CleanAbandonedCarts(ctx context.Context) error {
	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	var carts []models.Cart
	err := s.db.WithContext(ctx).
		Where("status = ? AND updated_at < ?", constants.CartStatusActive, cutoff).
		Find(&carts).Error
	if err != nil {
		return err
	}
	for _, cart := range carts {
		if err := s.db.WithContext(ctx).Model(&models.Cart{}).Where("id = ?", cart.ID).Update("status", constants.CartStatusAbandoned).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *cartService) findActiveCart(ctx context.Context, userID uint, preloadItems bool) (*models.Cart, error) {
	q := s.db.WithContext(ctx).Where("user_id = ? AND status = ?", userID, constants.CartStatusActive)
	if preloadItems {
		q = q.Preload("Items.Product")
	}
	var cart models.Cart
	if err := q.First(&cart).Error; err != nil {
		return nil, err
	}
	return &cart, nil
}

func (s *cartService) findCartItemByCartAndProduct(ctx context.Context, cartID, productID uint) (*models.CartItem, error) {
	var item models.CartItem
	if err := s.db.WithContext(ctx).Where("cart_id = ? AND product_id = ?", cartID, productID).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *cartService) findCartItemForUser(ctx context.Context, userID, cartItemID uint) (*models.CartItem, error) {
	var item models.CartItem
	err := s.db.WithContext(ctx).
		Joins("JOIN carts ON carts.id = cart_items.cart_id").
		Where("cart_items.id = ? AND carts.user_id = ? AND carts.status = ?", cartItemID, userID, constants.CartStatusActive).
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *cartService) getProductByID(ctx context.Context, id uint) (*models.Product, error) {
	var product models.Product
	if err := s.db.WithContext(ctx).First(&product, id).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

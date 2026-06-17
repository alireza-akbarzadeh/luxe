package services

import (
	"errors"
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
	GetOrCreateCart(userID uint) (*models.Cart, error)
	AddItem(userID uint, req AddItemRequest) (*models.CartItem, error)
	UpdateCartItem(userID uint, cartItemID uint, req UpdateCartItemRequest) error
	RemoveItem(userID uint, cartItemID uint) error
	GetCart(userID uint) (*models.Cart, error)
	ClearCart(userID uint) error
	CleanAbandonedCarts() error
}

type cartService struct {
	db *gorm.DB
}

func NewCartService(db *gorm.DB) CartServiceInterface {
	return &cartService{db: db}
}

// GetOrCreateCart returns existing active cart or creates a new one.
func (s *cartService) GetOrCreateCart(userID uint) (*models.Cart, error) {
	var cart models.Cart
	err := s.db.Where("user_id = ? AND status = ?", userID, constants.CartStatusActive).
		Preload("Items.Product"). // optional: preload for view
		First(&cart).Error
	if err == nil {
		return &cart, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, utils.ErrInternal(err)
	}

	// Create new cart
	cart = models.Cart{
		UserID:    userID,
		Status:    constants.CartStatusActive,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.db.Create(&cart).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}
	return &cart, nil
}

// AddItem adds a product to the cart.
func (s *cartService) AddItem(userID uint, req AddItemRequest) (*models.CartItem, error) {
	if req.Quantity <= 0 {
		return nil, utils.ErrBadRequest("quantity must be positive")
	}

	// Get product and check stock
	var product models.Product
	if err := s.db.First(&product, req.ProductID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("product not found")
		}
		return nil, utils.ErrInternal(err)
	}
	if product.Status != constants.ProductStatusActive {
		return nil, utils.ErrBadRequest("product is not available")
	}
	if !isProductStockAvailable(product, req.Quantity) {
		return nil, utils.ErrBadRequest("insufficient stock")
	}

	// Get or create cart
	cart, err := s.GetOrCreateCart(userID)
	if err != nil {
		return nil, err
	}

	// Check if item already exists
	var cartItem models.CartItem
	err = s.db.Where("cart_id = ? AND product_id = ?", cart.ID, req.ProductID).First(&cartItem).Error
	if err == nil {
		// Update quantity
		newQty := cartItem.Quantity + req.Quantity
		if !isProductStockAvailable(product, newQty) {
			return nil, utils.ErrBadRequest("insufficient stock for updated quantity")
		}
		cartItem.Quantity = newQty
		if err := s.db.Save(&cartItem).Error; err != nil {
			return nil, utils.ErrInternal(err)
		}
		return &cartItem, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, utils.ErrInternal(err)
	}

	// Create new cart item with price snapshot
	cartItem = models.CartItem{
		CartID:    cart.ID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
		Price:     product.Price,
	}
	if err := s.db.Create(&cartItem).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}
	return &cartItem, nil
}

// UpdateItemQuantity modifies existing cart item quantity.
func (s *cartService) UpdateCartItem(userID uint, cartItemID uint, req UpdateCartItemRequest) error {
	var cartItem models.CartItem
	if err := s.db.Joins("JOIN carts ON carts.id = cart_items.cart_id").
		Where("cart_items.id = ? AND carts.user_id = ? AND carts.status = ?", cartItemID, userID, constants.CartStatusActive).
		First(&cartItem).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrNotFound("cart item not found")
		}
		return utils.ErrInternal(err)
	}

	// Update quantity if provided and positive
	if req.Quantity > 0 {
		// Validate stock
		var product models.Product
		if err := s.db.First(&product, cartItem.ProductID).Error; err != nil {
			return utils.ErrInternal(err)
		}
		if !isProductStockAvailable(product, req.Quantity) {
			return utils.ErrBadRequest("insufficient stock")
		}
		cartItem.Quantity = req.Quantity
	}

	// Update color/size if provided (even empty string is allowed to clear)
	if req.Color != "" || req.Size != "" {
		cartItem.Color = req.Color
		cartItem.Size = req.Size
	}

	if err := s.db.Save(&cartItem).Error; err != nil {
		return utils.ErrInternal(err)
	}
	return nil
}

// RemoveItem deletes a cart item.
func (s *cartService) RemoveItem(userID uint, cartItemID uint) error {
	result := s.db.Where("id = ? AND cart_id IN (SELECT id FROM carts WHERE user_id = ? AND status = ?)",
		cartItemID, userID, "active").Delete(&models.CartItem{})
	if result.Error != nil {
		return utils.ErrInternal(result.Error)
	}
	if result.RowsAffected == 0 {
		return utils.ErrNotFound("cart item not found")
	}
	return nil
}

// GetCart returns full cart with items for the user.
func (s *cartService) GetCart(userID uint) (*models.Cart, error) {
	var cart models.Cart
	err := s.db.Where("user_id = ? AND status = ?", userID, "active").
		Preload("Items.Product").
		First(&cart).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Return empty cart (not created yet)
			return &models.Cart{UserID: userID, Items: []models.CartItem{}}, nil
		}
		return nil, utils.ErrInternal(err)
	}
	return &cart, nil
}

// ClearCart removes all items from active cart.
func (s *cartService) ClearCart(userID uint) error {
	// Delete all cart items belonging to user's active cart
	err := s.db.Where("cart_id IN (SELECT id FROM carts WHERE user_id = ? AND status = ?)", userID, "active").
		Delete(&models.CartItem{}).Error
	return err
}

func (s *cartService) CleanAbandonedCarts() error {
	// Find carts with status 'active' and older than 7 days
	var carts []models.Cart
	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	if err := s.db.Where("status = ? AND updated_at < ?", "active", cutoff).Find(&carts).Error; err != nil {
		return err
	}
	for _, cart := range carts {
		// Delete cart items or mark cart as abandoned
		s.db.Model(&cart).Update("status", "abandoned")
	}
	return nil
}

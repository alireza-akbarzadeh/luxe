package services

import (
	"context"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/repositories"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
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
	carts    repositories.CartRepository
	products repositories.ProductRepository
}

func NewCartService(carts repositories.CartRepository, products repositories.ProductRepository) CartServiceInterface {
	return &cartService{carts: carts, products: products}
}

func (s *cartService) GetOrCreateCart(userID uint) (*models.Cart, error) {
	ctx := context.Background()
	cart, err := s.carts.FindActiveByUserID(ctx, userID, true)
	if err == nil {
		return cart, nil
	}
	if !repositories.IsRecordNotFound(err) {
		return nil, utils.ErrInternal(err)
	}

	newCart := models.Cart{
		UserID:    userID,
		Status:    constants.CartStatusActive,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.carts.Create(ctx, &newCart); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return &newCart, nil
}

func (s *cartService) AddItem(userID uint, req AddItemRequest) (*models.CartItem, error) {
	if req.Quantity <= 0 {
		return nil, utils.ErrBadRequest("quantity must be positive")
	}

	ctx := context.Background()
	product, err := s.products.GetByID(ctx, req.ProductID)
	if err != nil {
		if repositories.IsRecordNotFound(err) {
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

	cart, err := s.GetOrCreateCart(userID)
	if err != nil {
		return nil, err
	}

	cartItem, err := s.carts.FindCartItemByCartAndProduct(ctx, cart.ID, req.ProductID)
	if err == nil {
		newQty := cartItem.Quantity + req.Quantity
		if !isProductStockAvailable(*product, newQty) {
			return nil, utils.ErrBadRequest("insufficient stock for updated quantity")
		}
		cartItem.Quantity = newQty
		if err := s.carts.SaveCartItem(ctx, cartItem); err != nil {
			return nil, utils.ErrInternal(err)
		}
		return cartItem, nil
	}
	if !repositories.IsRecordNotFound(err) {
		return nil, utils.ErrInternal(err)
	}

	newItem := models.CartItem{
		CartID:    cart.ID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
		Price:     product.Price,
	}
	if err := s.carts.CreateCartItem(ctx, &newItem); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return &newItem, nil
}

func (s *cartService) UpdateCartItem(userID uint, cartItemID uint, req UpdateCartItemRequest) error {
	ctx := context.Background()
	cartItem, err := s.carts.FindCartItemForUser(ctx, userID, cartItemID)
	if err != nil {
		if repositories.IsRecordNotFound(err) {
			return utils.ErrNotFound("cart item not found")
		}
		return utils.ErrInternal(err)
	}

	if req.Quantity > 0 {
		product, err := s.products.GetByID(ctx, cartItem.ProductID)
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

	if err := s.carts.SaveCartItem(ctx, cartItem); err != nil {
		return utils.ErrInternal(err)
	}
	return nil
}

func (s *cartService) RemoveItem(userID uint, cartItemID uint) error {
	rows, err := s.carts.DeleteCartItem(context.Background(), userID, cartItemID)
	if err != nil {
		return utils.ErrInternal(err)
	}
	if rows == 0 {
		return utils.ErrNotFound("cart item not found")
	}
	return nil
}

func (s *cartService) GetCart(userID uint) (*models.Cart, error) {
	cart, err := s.carts.FindActiveByUserID(context.Background(), userID, true)
	if err != nil {
		if repositories.IsRecordNotFound(err) {
			return &models.Cart{UserID: userID, Items: []models.CartItem{}}, nil
		}
		return nil, utils.ErrInternal(err)
	}
	return cart, nil
}

func (s *cartService) ClearCart(userID uint) error {
	if err := s.carts.DeleteItemsForActiveCart(context.Background(), userID); err != nil {
		return utils.ErrInternal(err)
	}
	return nil
}

func (s *cartService) CleanAbandonedCarts() error {
	ctx := context.Background()
	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	carts, err := s.carts.FindStaleActiveCarts(ctx, cutoff)
	if err != nil {
		return err
	}
	for _, cart := range carts {
		if err := s.carts.UpdateStatus(ctx, cart.ID, constants.CartStatusAbandoned); err != nil {
			return err
		}
	}
	return nil
}

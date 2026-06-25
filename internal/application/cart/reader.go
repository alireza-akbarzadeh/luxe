package cart

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

// Reader loads cart models for HTTP and legacy services.
type Reader interface {
	FindActiveCart(ctx context.Context, userID uint, preloadItems bool) (*models.Cart, error)
	FindCartItemByCartAndProduct(ctx context.Context, cartID, productID uint) (*models.CartItem, error)
	FindCartItemForUser(ctx context.Context, userID, cartItemID uint) (*models.CartItem, error)
	GetProductByID(ctx context.Context, id uint) (*models.Product, error)
}

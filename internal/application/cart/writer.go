package cart

import (
	"context"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

// Writer persists cart changes.
type Writer interface {
	CreateActiveCart(ctx context.Context, userID uint) (*models.Cart, error)
	SaveCartItem(ctx context.Context, item *models.CartItem) error
	CreateCartItem(ctx context.Context, item *models.CartItem) error
	RemoveCartItemForUser(ctx context.Context, userID, cartItemID uint) (int64, error)
	ClearActiveCartItems(ctx context.Context, userID uint) error
	MarkAbandonedCarts(ctx context.Context, cutoff time.Time) error
	MarkConverted(ctx context.Context, cartID uint) error
}

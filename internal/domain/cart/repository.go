package cart

import "context"

// Repository loads and persists carts.
type Repository interface {
	GetActiveByUserID(ctx context.Context, userID uint) (*Cart, error)
}

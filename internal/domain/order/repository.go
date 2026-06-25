package order

import "context"

// Repository persists orders.
type Repository interface {
	GetByID(ctx context.Context, id uint) (*Order, error)
}

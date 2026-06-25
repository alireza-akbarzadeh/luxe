package payment

import "context"

// Repository loads and persists payments.
type Repository interface {
	GetByOrderID(ctx context.Context, orderID uint) (*Payment, error)
}

package address

import "context"

// Repository loads and persists user addresses.
type Repository interface {
	ListByUser(ctx context.Context, userID uint) ([]Address, error)
}

package order

import "errors"

var (
	ErrOrderNotFound = errors.New("order not found")
	ErrCannotCancel  = errors.New("order cannot be cancelled")
)

// Order is the commerce order aggregate (domain view).
type Order struct {
	ID          uint
	UserID      uint
	Status      string
	TotalCents  int64
	Currency    string
}

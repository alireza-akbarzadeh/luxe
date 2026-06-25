package cart

import "errors"

var (
	ErrCartNotFound   = errors.New("cart not found")
	ErrEmptyCart      = errors.New("cart is empty")
	ErrInsufficientStock = errors.New("insufficient stock")
)

// Item is a line in a shopping cart.
type Item struct {
	ID        uint
	ProductID uint
	Quantity  int
	UnitCents int64
}

// Cart is the cart aggregate.
type Cart struct {
	ID     uint
	UserID *uint
	Items  []Item
}

// TotalCents sums line totals.
func (c Cart) TotalCents() int64 {
	var total int64
	for _, item := range c.Items {
		total += int64(item.Quantity) * item.UnitCents
	}
	return total
}

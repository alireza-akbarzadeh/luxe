package order

import "errors"

var (
	ErrOrderNotFound = errors.New("order not found")
	ErrCannotCancel  = errors.New("order cannot be cancelled")
)

// Order is a slim domain view used for pure rules (not GORM models).
type Order struct {
	ID         uint
	UserID     uint
	Status     string
	TotalCents int64
	Currency   string
}

// Service holds order domain rules.
type Service struct{}

// NewService creates an order domain service.
func NewService() *Service { return &Service{} }

// CanCancel reports whether a customer may cancel an order.
func (s *Service) CanCancel(o Order) error {
	switch o.Status {
	case "delivered", "completed", "refunded", "cancelled":
		return ErrCannotCancel
	default:
		return nil
	}
}

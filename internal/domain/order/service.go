package order

// Service holds order domain rules.
type Service struct{}

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

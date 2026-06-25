// Package cart provides the domain service for the cart.
package cart

// Service holds cart domain rules.
type Service struct{}

func NewService() *Service { return &Service{} }

// ValidateCheckout ensures a cart can proceed to checkout.
func (s *Service) ValidateCheckout(c Cart) error {
	if len(c.Items) == 0 {
		return ErrEmptyCart
	}
	for _, item := range c.Items {
		if item.Quantity <= 0 {
			return ErrInsufficientStock
		}
	}
	return nil
}

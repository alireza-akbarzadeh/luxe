package checkout

import "errors"

var (
	ErrCheckoutFailed  = errors.New("checkout failed")
	ErrPaymentRequired = errors.New("payment required")
)

// CheckoutInput is the domain checkout command.
type CheckoutInput struct {
	UserID         uint
	CartTotalCents int64
	Currency       string
	PaymentMethod  string
}

// Service validates checkout invariants.
type Service struct{}

func NewService() *Service { return &Service{} }

func (s *Service) Validate(in CheckoutInput) error {
	if in.UserID == 0 {
		return ErrCheckoutFailed
	}
	if in.CartTotalCents <= 0 {
		return ErrCheckoutFailed
	}
	if in.PaymentMethod == "" {
		return ErrPaymentRequired
	}
	return nil
}

package payment

import "errors"

var ErrInvalidAmount = errors.New("payment amount must be positive")

// Payment represents a payment attempt.
type Payment struct {
	ID         uint
	OrderID    uint
	AmountCents int64
	Currency   string
	Provider   string
	Status     string
}

type Service struct{}

func NewService() *Service { return &Service{} }

func (s *Service) ValidateAmount(cents int64) error {
	if cents <= 0 {
		return ErrInvalidAmount
	}
	return nil
}

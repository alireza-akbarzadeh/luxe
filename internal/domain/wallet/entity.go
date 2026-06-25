package wallet

import "errors"

var ErrInsufficientBalance = errors.New("insufficient wallet balance")

type Wallet struct {
	UserID       uint
	BalanceCents int64
	Currency     string
}

type Service struct{}

func NewService() *Service { return &Service{} }

func (s *Service) CanDebit(w Wallet, amountCents int64) error {
	if amountCents <= 0 {
		return ErrInsufficientBalance
	}
	if w.BalanceCents < amountCents {
		return ErrInsufficientBalance
	}
	return nil
}

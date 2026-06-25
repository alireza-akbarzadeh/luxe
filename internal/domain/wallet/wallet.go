package wallet

import "errors"

var (
	ErrInsufficientBalance = errors.New("insufficient wallet balance")
	ErrInvalidAmount       = errors.New("amount must be positive")
)

// CanApplyDelta reports whether balance can change by delta without going negative.
func CanApplyDelta(balance, delta float64) error {
	newBalance := balance + delta
	if newBalance < 0 {
		return ErrInsufficientBalance
	}
	return nil
}

// ValidatePositiveAmount ensures deposit/withdraw amounts are valid.
func ValidatePositiveAmount(amount float64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	return nil
}

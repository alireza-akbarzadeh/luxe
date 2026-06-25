package payment

import (
	"errors"
	"time"
)

var (
	ErrCardDeclined = errors.New("card declined")
	ErrCardExpired  = errors.New("card expired")
	ErrInvalidCVV   = errors.New("invalid CVV")
)

// Card holds mock-gateway card fields for validation.
type Card struct {
	Number      string
	ExpiryMonth int
	ExpiryYear  int
	CVV         string
}

// ValidateCard checks card fields for the mock payment gateway.
func ValidateCard(c Card, now time.Time) error {
	if c.Number == "0000000000000000" {
		return ErrCardDeclined
	}
	year := now.Year()
	month := int(now.Month())
	if c.ExpiryYear < year || (c.ExpiryYear == year && c.ExpiryMonth < month) {
		return ErrCardExpired
	}
	if len(c.CVV) != 3 {
		return ErrInvalidCVV
	}
	return nil
}

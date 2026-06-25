// Package shared holds cross-domain value objects and primitives.
package shared

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

// Money stores amounts in minor units (e.g. cents) to avoid float drift.
type Money struct {
	amountCents int64
	currency    string
}

var (
	ErrInvalidMoneyAmount = errors.New("money amount must be non-negative")
	ErrInvalidCurrency    = errors.New("currency must be a 3-letter ISO code")
)

// NewMoney creates a Money value from major units (e.g. dollars).
func NewMoney(major float64, currency string) (Money, error) {
	if major < 0 {
		return Money{}, ErrInvalidMoneyAmount
	}
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if len(currency) != 3 {
		return Money{}, ErrInvalidCurrency
	}
	cents := int64(math.Round(major * 100))
	return Money{amountCents: cents, currency: currency}, nil
}

// NewMoneyFromCents creates Money from minor units.
func NewMoneyFromCents(cents int64, currency string) (Money, error) {
	if cents < 0 {
		return Money{}, ErrInvalidMoneyAmount
	}
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if len(currency) != 3 {
		return Money{}, ErrInvalidCurrency
	}
	return Money{amountCents: cents, currency: currency}, nil
}

// Cents returns the amount in minor units.
func (m Money) Cents() int64 { return m.amountCents }

// Currency returns the ISO 4217 code.
func (m Money) Currency() string { return m.currency }

// Major returns the amount in major units.
func (m Money) Major() float64 { return float64(m.amountCents) / 100 }

// Add returns the sum when currencies match.
func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, fmt.Errorf("currency mismatch: %s vs %s", m.currency, other.currency)
	}
	return Money{amountCents: m.amountCents + other.amountCents, currency: m.currency}, nil
}

// IsZero reports whether the amount is zero.
func (m Money) IsZero() bool { return m.amountCents == 0 }

// IsPositive reports whether the amount is greater than zero.
func (m Money) IsPositive() bool { return m.amountCents > 0 }

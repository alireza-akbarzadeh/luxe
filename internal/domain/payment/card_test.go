package payment

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestValidateCard(t *testing.T) {
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	valid := Card{Number: "4111111111111111", ExpiryMonth: 12, ExpiryYear: 2027, CVV: "123"}

	require.NoError(t, ValidateCard(valid, now))
	require.ErrorIs(t, ValidateCard(Card{Number: "0000000000000000", ExpiryMonth: 12, ExpiryYear: 2027, CVV: "123"}, now), ErrCardDeclined)
	require.ErrorIs(t, ValidateCard(Card{Number: "4111111111111111", ExpiryMonth: 1, ExpiryYear: 2025, CVV: "123"}, now), ErrCardExpired)
	require.ErrorIs(t, ValidateCard(Card{Number: "4111111111111111", ExpiryMonth: 12, ExpiryYear: 2027, CVV: "12"}, now), ErrInvalidCVV)
}

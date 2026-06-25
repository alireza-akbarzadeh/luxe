package wallet

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCanApplyDelta(t *testing.T) {
	require.NoError(t, CanApplyDelta(100, -50))
	require.ErrorIs(t, CanApplyDelta(10, -20), ErrInsufficientBalance)
}

func TestValidatePositiveAmount(t *testing.T) {
	require.NoError(t, ValidatePositiveAmount(1))
	require.ErrorIs(t, ValidatePositiveAmount(0), ErrInvalidAmount)
}

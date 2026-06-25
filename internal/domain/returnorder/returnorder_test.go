package returnorder

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateCreateRequest(t *testing.T) {
	require.NoError(t, ValidateCreateRequest(StatusDelivered, 0))
	require.NoError(t, ValidateCreateRequest(StatusCompleted, 0))

	require.ErrorIs(t, ValidateCreateRequest("pending", 0), ErrOrderNotEligible)
	require.ErrorIs(t, ValidateCreateRequest(StatusDelivered, 1), ErrOpenReturnExists)
}

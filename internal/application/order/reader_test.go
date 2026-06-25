package order

import (
	"testing"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/stretchr/testify/assert"
)

func TestListFilterFromDTO_DefaultLimit(t *testing.T) {
	q := ListFilterFromDTO(42, dto.OrderListFilters{})
	assert.Equal(t, uint(42), *q.UserID)
	assert.Equal(t, 20, q.Limit)
}

func TestListFilterFromDTO_StatusFilter(t *testing.T) {
	q := ListFilterFromDTO(1, dto.OrderListFilters{Status: constants.OrderStatusPaid})
	assert.Equal(t, constants.OrderStatusPaid, q.Status)
}

func TestOrderStatusDelayedConstant(t *testing.T) {
	assert.Equal(t, "delayed", constants.OrderStatusDelayed)
}

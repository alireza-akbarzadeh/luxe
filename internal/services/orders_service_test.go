package services

import (
	"testing"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/stretchr/testify/assert"
)

func TestOrderFiltersFromDTO_DefaultLimit(t *testing.T) {
	q := orderFiltersFromDTO(42, dto.OrderListFilters{})
	assert.Equal(t, uint(42), *q.UserID)
	assert.Equal(t, 0, q.Limit)
}

func TestOrderListQuery_StatusFilter(t *testing.T) {
	q := orderFiltersFromDTO(1, dto.OrderListFilters{Status: constants.OrderStatusPaid})
	assert.Equal(t, constants.OrderStatusPaid, q.Status)
}

func TestOrderStatusDelayedConstant(t *testing.T) {
	assert.Equal(t, "delayed", constants.OrderStatusDelayed)
}

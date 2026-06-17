package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCartService_AddItem_InvalidQuantity(t *testing.T) {
	svc := NewCartService(nil)

	_, err := svc.AddItem(context.Background(), 1, AddItemRequest{
		ProductID: 1,
		Quantity:  0,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "quantity must be positive")
}

func TestCartService_AddItem_NegativeQuantity(t *testing.T) {
	svc := NewCartService(nil)

	_, err := svc.AddItem(context.Background(), 1, AddItemRequest{
		ProductID: 1,
		Quantity:  -1,
	})
	assert.Error(t, err)
}

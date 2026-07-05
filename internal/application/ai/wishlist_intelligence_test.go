package ai

import (
	"testing"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseWishlistIntelligenceJSON_Valid(t *testing.T) {
	raw := `{"summary":"Two items are on sale.","highlights":["Save on watch"],"items":[{"product_id":1,"product_name":"Watch","priority":"buy_now","reason":"In stock and discounted"}]}`
	products := []models.Product{{ID: 1, Name: "Watch"}}
	result, err := parseWishlistIntelligenceJSON(raw, products)
	require.NoError(t, err)
	assert.Equal(t, "Two items are on sale.", result.Summary)
	require.Len(t, result.Items, 1)
	assert.Equal(t, "buy_now", result.Items[0].Priority)
}

func TestParseWishlistIntelligenceJSON_EmptySummary(t *testing.T) {
	_, err := parseWishlistIntelligenceJSON(`{"summary":"","items":[]}`, nil)
	assert.Error(t, err)
}

package dto

import (
	"encoding/json"
	"testing"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToNavItemResponseIncludesIDAndOrder(t *testing.T) {
	href := "/shop"
	menu := &models.NavMenu{
		ID:        7,
		Label:     "Gift Cards",
		Type:      "link",
		Href:      &href,
		SortOrder: 7,
	}

	resp, err := ToNavItemResponse(menu)
	require.NoError(t, err)
	assert.Equal(t, uint(7), resp.ID)
	assert.Equal(t, 7, resp.Order)
	assert.Equal(t, "Gift Cards", resp.Label)

	raw, err := json.Marshal(resp)
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"id":7`)
	assert.Contains(t, string(raw), `"order":7`)
}

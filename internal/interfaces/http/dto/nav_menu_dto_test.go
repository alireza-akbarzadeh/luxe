package dto

import (
	"context"
	"testing"

	"github.com/alireza-akbarzadeh/luxe/internal/i18n"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

func TestToNavItemResponseIncludesIDAndOrder(t *testing.T) {
	href := "/shop"
	menu := &models.NavMenu{
		ID:        7,
		Label:     "Gift Cards",
		LabelI18n: datatypes.JSON(`{"en":"Gift Cards","fa":"کارت هدیه"}`),
		Type:      "link",
		Href:      &href,
		SortOrder: 7,
	}

	ctx := i18n.WithLocale(context.Background(), "fa")
	resp, err := ToNavItemResponse(ctx, menu)
	require.NoError(t, err)
	assert.Equal(t, uint(7), resp.ID)
	assert.Equal(t, 7, resp.Order)
	assert.Equal(t, "کارت هدیه", resp.Label)
	assert.Equal(t, "Gift Cards", resp.LabelI18n["en"])
}

func TestToNavItemResponseDefaultsToEnglish(t *testing.T) {
	menu := &models.NavMenu{
		ID:    1,
		Label: "Shop",
		Type:  "link",
	}

	ctx := i18n.WithLocale(context.Background(), "es")
	resp, err := ToNavItemResponse(ctx, menu)
	require.NoError(t, err)
	assert.Equal(t, "Shop", resp.Label)
}

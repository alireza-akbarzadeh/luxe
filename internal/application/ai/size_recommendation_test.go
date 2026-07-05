package ai

import (
	"testing"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSizeRecommendationJSON_Valid(t *testing.T) {
	raw := `{"recommended_size":"M","alternative_size":"L","fit_notes":"true_to_size","confidence":"high","summary":"Most reviewers say it fits as labeled.","tips":["Check the size chart"]}`
	result, err := parseSizeRecommendationJSON(raw)
	require.NoError(t, err)
	assert.Equal(t, "M", result.RecommendedSize)
	assert.Equal(t, "true_to_size", result.FitNotes)
	assert.Equal(t, "high", result.Confidence)
}

func TestProductSizeValues_FromAttributesAndLegacy(t *testing.T) {
	product := &models.Product{
		Sizes: []string{" S ", "M"},
		Attributes: []models.ProductAttribute{
			{Name: "size", Values: []string{"M", "L"}},
		},
	}
	sizes := productSizeValues(product)
	assert.Equal(t, []string{"M", "L", "S"}, sizes)
}

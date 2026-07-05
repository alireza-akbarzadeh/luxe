package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePurchaseAdvisorJSON_Valid(t *testing.T) {
	raw := `{"verdict":"buy","confidence":"high","summary":"Strong reviews and fair price.","pros":["Well reviewed"],"cons":["Limited stock"],"ideal_for":["Gift buyers"],"considerations":["Check return policy"]}`
	result, err := parsePurchaseAdvisorJSON(raw)
	require.NoError(t, err)
	assert.Equal(t, "buy", result.Verdict)
	assert.Equal(t, "high", result.Confidence)
	assert.Equal(t, "Strong reviews and fair price.", result.Summary)
	assert.Len(t, result.Pros, 1)
	assert.Len(t, result.Considerations, 1)
}

func TestParsePurchaseAdvisorJSON_NormalizesVerdict(t *testing.T) {
	raw := `{"verdict":"maybe","confidence":"unknown","summary":"Mixed signals."}`
	result, err := parsePurchaseAdvisorJSON(raw)
	require.NoError(t, err)
	assert.Equal(t, "consider", result.Verdict)
	assert.Equal(t, "medium", result.Confidence)
}

func TestParsePurchaseAdvisorJSON_EmptySummary(t *testing.T) {
	_, err := parsePurchaseAdvisorJSON(`{"verdict":"wait","confidence":"low","summary":""}`)
	assert.Error(t, err)
}

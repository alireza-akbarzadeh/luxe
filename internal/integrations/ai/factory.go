package ai

import (
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/config"
)

// NewFromConfig returns the configured LLM provider.
func NewFromConfig(cfg config.AIConfig) Provider {
	baseURL := strings.TrimSpace(cfg.BaseURL)
	if baseURL == "" {
		baseURL = "http://localhost:11434/v1"
	}
	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		model = "llama3.2"
	}
	return NewProvider(baseURL, cfg.APIKey, model, cfg.Enabled)
}

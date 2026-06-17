package services

import (
	"testing"

	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewWalletService_StripeConfiguration(t *testing.T) {
	disabled := NewWalletService(nil, &config.Config{Stripe: config.StripeConfig{Enabled: false}})
	assert.NotNil(t, disabled)

	enabled := NewWalletService(nil, &config.Config{
		Stripe: config.StripeConfig{
			Enabled:   true,
			SecretKey: "sk_test_fake",
		},
		Email: config.Email{FrontendURL: "http://localhost:3000"},
	})
	assert.NotNil(t, enabled)
}

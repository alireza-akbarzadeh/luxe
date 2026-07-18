package privacyrule

import (
	"errors"
	"regexp"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
)

var (
	ErrInvalidName     = errors.New("privacy rule name is required")
	ErrInvalidKey      = errors.New("privacy rule key must be lowercase letters, numbers, dots, and hyphens")
	ErrInvalidProvider = errors.New("invalid privacy rule provider")
	ErrInvalidContent  = errors.New("privacy rule markdown content is required")
)

var keyPattern = regexp.MustCompile(`^[a-z0-9]+(?:[._-][a-z0-9]+)*$`)

var allowedProviders = map[string]struct{}{
	constants.PrivacyRuleProviderPlatform: {},
	constants.PrivacyRuleProviderStripe:   {},
	constants.PrivacyRuleProviderPaypal:   {},
	constants.PrivacyRuleProviderWallet:   {},
	constants.PrivacyRuleProviderGiftCard: {},
	constants.PrivacyRuleProviderShipping: {},
	constants.PrivacyRuleProviderAI:       {},
	constants.PrivacyRuleProviderAll:      {},
}

// Service holds privacy rule domain rules.
type Service struct{}

// NewService creates a privacy rule domain service.
func NewService() *Service { return &Service{} }

// ValidateName ensures a display name is present.
func (s *Service) ValidateName(name string) error {
	if strings.TrimSpace(name) == "" {
		return ErrInvalidName
	}
	return nil
}

// ValidateKey ensures a stable lookup key for cross-app resolution.
func (s *Service) ValidateKey(key string) error {
	key = strings.TrimSpace(key)
	if key == "" || !keyPattern.MatchString(key) {
		return ErrInvalidKey
	}
	return nil
}

// ValidateProvider ensures the provider is a known integration surface.
func (s *Service) ValidateProvider(provider string) error {
	if _, ok := allowedProviders[provider]; !ok {
		return ErrInvalidProvider
	}
	return nil
}

// ValidateContent ensures markdown body is non-empty.
func (s *Service) ValidateContent(content string) error {
	if strings.TrimSpace(content) == "" {
		return ErrInvalidContent
	}
	return nil
}

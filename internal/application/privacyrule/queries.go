package privacyrule

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

// Queries orchestrates privacy rule read use cases.
type Queries struct {
	repo *postgres.PrivacyRuleRepository
}

// NewQueries creates privacy rule query use cases.
func NewQueries(repo *postgres.PrivacyRuleRepository) *Queries {
	return &Queries{repo: repo}
}

// GetByID loads a privacy rule by ID.
func (q *Queries) GetByID(ctx context.Context, id uint) (*models.PrivacyRule, error) {
	return q.repo.GetByID(ctx, id)
}

// GetActiveByKey loads an active privacy rule by key and locale.
func (q *Queries) GetActiveByKey(ctx context.Context, key, locale string) (*models.PrivacyRule, error) {
	if locale == "" {
		locale = "en"
	}
	return q.repo.GetActiveByKeyLocale(ctx, key, locale)
}

// List returns paginated privacy rules.
func (q *Queries) List(ctx context.Context, req *dto.ListPrivacyRulesRequest) ([]models.PrivacyRule, int64, error) {
	return q.repo.List(ctx, req)
}

// ListActiveByProvider returns active rules for a provider (+ "all").
func (q *Queries) ListActiveByProvider(ctx context.Context, provider, locale string) ([]models.PrivacyRule, error) {
	return q.repo.ListActiveByProvider(ctx, provider, locale)
}

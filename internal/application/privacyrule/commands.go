package privacyrule

import (
	"context"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	domain "github.com/alireza-akbarzadeh/luxe/internal/domain/privacyrule"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

// Commands orchestrates privacy rule write use cases.
type Commands struct {
	domain *domain.Service
	repo   *postgres.PrivacyRuleRepository
}

// NewCommands creates privacy rule command use cases.
func NewCommands(domainSvc *domain.Service, repo *postgres.PrivacyRuleRepository) *Commands {
	return &Commands{domain: domainSvc, repo: repo}
}

// Create inserts a new privacy rule after domain validation.
func (c *Commands) Create(ctx context.Context, req *dto.CreatePrivacyRuleRequest) (*models.PrivacyRule, error) {
	if err := c.domain.ValidateName(req.Name); err != nil {
		return nil, utils.ErrBadRequest(err.Error())
	}
	if err := c.domain.ValidateKey(req.Key); err != nil {
		return nil, utils.ErrBadRequest(err.Error())
	}
	if err := c.domain.ValidateProvider(req.Provider); err != nil {
		return nil, utils.ErrBadRequest(err.Error())
	}
	if err := c.domain.ValidateContent(req.ContentMarkdown); err != nil {
		return nil, utils.ErrBadRequest(err.Error())
	}

	rule, err := BuildCreateModel(req)
	if err != nil {
		return nil, utils.ErrBadRequest("invalid effective_at; use RFC3339")
	}

	if err := c.repo.Create(ctx, rule); err != nil {
		if isUniqueViolation(err) {
			return nil, utils.ErrBadRequest("a privacy rule with this key and locale already exists")
		}
		return nil, err
	}
	return rule, nil
}

// Update validates and persists privacy rule field changes.
func (c *Commands) Update(ctx context.Context, rule *models.PrivacyRule, req *dto.UpdatePrivacyRuleRequest) error {
	if req.Name != nil {
		if err := c.domain.ValidateName(*req.Name); err != nil {
			return utils.ErrBadRequest(err.Error())
		}
	}
	if req.Key != nil {
		if err := c.domain.ValidateKey(*req.Key); err != nil {
			return utils.ErrBadRequest(err.Error())
		}
	}
	if req.Provider != nil {
		if err := c.domain.ValidateProvider(*req.Provider); err != nil {
			return utils.ErrBadRequest(err.Error())
		}
	}
	if req.ContentMarkdown != nil {
		if err := c.domain.ValidateContent(*req.ContentMarkdown); err != nil {
			return utils.ErrBadRequest(err.Error())
		}
	}

	prevStatus := rule.Status
	contentChanged, err := ApplyUpdateDTO(rule, req)
	if err != nil {
		return utils.ErrBadRequest("invalid effective_at; use RFC3339")
	}

	shouldBump := false
	if req.BumpVersion != nil && *req.BumpVersion {
		shouldBump = true
	}
	if contentChanged && (rule.Status == constants.PrivacyRuleStatusActive || prevStatus == constants.PrivacyRuleStatusActive) {
		shouldBump = true
	}
	if prevStatus != constants.PrivacyRuleStatusActive && rule.Status == constants.PrivacyRuleStatusActive {
		shouldBump = true
	}
	if shouldBump {
		rule.Version++
	}

	if err := c.repo.Save(ctx, rule); err != nil {
		if isUniqueViolation(err) {
			return utils.ErrBadRequest("a privacy rule with this key and locale already exists")
		}
		return err
	}
	return nil
}

// Delete removes a privacy rule by id.
func (c *Commands) Delete(ctx context.Context, id uint) (int64, error) {
	return c.repo.DeleteByID(ctx, id)
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") ||
		strings.Contains(msg, "unique constraint") ||
		strings.Contains(msg, "uq_privacy_rules_key_locale")
}

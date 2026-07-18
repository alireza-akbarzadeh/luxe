package postgres

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// PrivacyRuleRepository persists privacy rules with GORM.
type PrivacyRuleRepository struct {
	db *gorm.DB
}

// NewPrivacyRuleRepository creates a GORM-backed privacy rule repository.
func NewPrivacyRuleRepository(db *gorm.DB) *PrivacyRuleRepository {
	return &PrivacyRuleRepository{db: db}
}

// GetByID loads a privacy rule with workflow state.
func (r *PrivacyRuleRepository) GetByID(ctx context.Context, id uint) (*models.PrivacyRule, error) {
	var rule models.PrivacyRule
	if err := r.db.WithContext(ctx).Preload("WorkflowState").First(&rule, id).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

// GetByKeyLocale loads a privacy rule by key and locale.
func (r *PrivacyRuleRepository) GetByKeyLocale(ctx context.Context, key, locale string) (*models.PrivacyRule, error) {
	var rule models.PrivacyRule
	err := r.db.WithContext(ctx).Preload("WorkflowState").
		Where("key = ? AND locale = ?", key, locale).
		First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// GetActiveByKeyLocale returns the active rule for key+locale (for app consumption).
func (r *PrivacyRuleRepository) GetActiveByKeyLocale(ctx context.Context, key, locale string) (*models.PrivacyRule, error) {
	var rule models.PrivacyRule
	err := r.db.WithContext(ctx).Preload("WorkflowState").
		Where("key = ? AND locale = ? AND status = ?", key, locale, "active").
		First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// Create inserts a privacy rule row.
func (r *PrivacyRuleRepository) Create(ctx context.Context, rule *models.PrivacyRule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

// Save persists privacy rule field changes.
func (r *PrivacyRuleRepository) Save(ctx context.Context, rule *models.PrivacyRule) error {
	return r.db.WithContext(ctx).Save(rule).Error
}

// DeleteByID removes a privacy rule.
func (r *PrivacyRuleRepository) DeleteByID(ctx context.Context, id uint) (int64, error) {
	result := r.db.WithContext(ctx).Delete(&models.PrivacyRule{}, id)
	return result.RowsAffected, result.Error
}

// List returns paginated privacy rules matching filters.
func (r *PrivacyRuleRepository) List(ctx context.Context, req *dto.ListPrivacyRulesRequest) ([]models.PrivacyRule, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.PrivacyRule{})

	if req.Search != "" {
		search := "%" + req.Search + "%"
		query = query.Where(
			"privacy_rules.name ILIKE ? OR privacy_rules.key ILIKE ? OR privacy_rules.summary ILIKE ?",
			search, search, search,
		)
	}
	if req.Status != "" {
		query = query.Where("privacy_rules.status = ?", req.Status)
	}
	if req.Provider != "" {
		query = query.Where("privacy_rules.provider = ?", req.Provider)
	}
	if req.Locale != "" {
		query = query.Where("privacy_rules.locale = ?", req.Locale)
	}
	if req.Key != "" {
		query = query.Where("privacy_rules.key = ?", req.Key)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	var rules []models.PrivacyRule
	err := query.Preload("WorkflowState").
		Order("privacy_rules.updated_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&rules).Error
	if err != nil {
		return nil, 0, err
	}
	return rules, total, nil
}

// ListActiveByProvider returns active rules for a provider (includes provider=all).
func (r *PrivacyRuleRepository) ListActiveByProvider(ctx context.Context, provider, locale string) ([]models.PrivacyRule, error) {
	query := r.db.WithContext(ctx).Preload("WorkflowState").
		Where("status = ?", "active").
		Where("provider IN ?", []string{provider, "all"})
	if locale != "" {
		query = query.Where("locale = ?", locale)
	}
	var rules []models.PrivacyRule
	err := query.Order("provider ASC, key ASC").Find(&rules).Error
	return rules, err
}

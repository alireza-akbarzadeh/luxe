package postgres

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// DeliveryCalendarRuleRepository persists delivery calendar rule toggles with GORM.
type DeliveryCalendarRuleRepository struct {
	db *gorm.DB
}

// NewDeliveryCalendarRuleRepository creates a GORM-backed delivery calendar rule repository.
func NewDeliveryCalendarRuleRepository(db *gorm.DB) *DeliveryCalendarRuleRepository {
	return &DeliveryCalendarRuleRepository{db: db}
}

// List returns all rules ordered by sort_order.
func (r *DeliveryCalendarRuleRepository) List(ctx context.Context) ([]models.DeliveryCalendarRule, error) {
	var rules []models.DeliveryCalendarRule
	err := r.db.WithContext(ctx).Order("sort_order ASC").Find(&rules).Error
	return rules, err
}

// EnabledMap returns a rule_key -> enabled lookup for the delivery calculator.
func (r *DeliveryCalendarRuleRepository) EnabledMap(ctx context.Context) (map[string]bool, error) {
	rules, err := r.List(ctx)
	if err != nil {
		return nil, err
	}
	enabled := make(map[string]bool, len(rules))
	for _, rule := range rules {
		enabled[rule.RuleKey] = rule.Enabled
	}
	return enabled, nil
}

// SetEnabled updates the enabled flag for a rule by key. Returns rows affected.
func (r *DeliveryCalendarRuleRepository) SetEnabled(ctx context.Context, ruleKey string, enabled bool) (int64, error) {
	result := r.db.WithContext(ctx).Model(&models.DeliveryCalendarRule{}).
		Where("rule_key = ?", ruleKey).
		Update("enabled", enabled)
	return result.RowsAffected, result.Error
}

// BulkSetEnabled updates enabled flags for multiple rules inside a transaction.
func (r *DeliveryCalendarRuleRepository) BulkSetEnabled(ctx context.Context, updates map[string]bool) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for key, enabled := range updates {
			if err := tx.Model(&models.DeliveryCalendarRule{}).
				Where("rule_key = ?", key).
				Update("enabled", enabled).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

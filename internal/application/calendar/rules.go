package calendar

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

// ListRules returns all delivery calendar rule toggles.
func (s *Service) ListRules(ctx context.Context) ([]dto.DeliveryCalendarRuleResponse, error) {
	rules, err := s.ruleRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	resp := make([]dto.DeliveryCalendarRuleResponse, 0, len(rules))
	for i := range rules {
		resp = append(resp, toRuleResponse(&rules[i]))
	}
	return resp, nil
}

// UpdateRules bulk-updates enabled flags for the given rule keys.
func (s *Service) UpdateRules(ctx context.Context, req *dto.UpdateDeliveryCalendarRulesRequest) ([]dto.DeliveryCalendarRuleResponse, error) {
	if len(req.Rules) == 0 {
		return nil, utils.ErrBadRequest("rules payload must not be empty")
	}
	updates := make(map[string]bool, len(req.Rules))
	for _, item := range req.Rules {
		updates[item.RuleKey] = item.Enabled
	}
	if err := s.ruleRepo.BulkSetEnabled(ctx, updates); err != nil {
		return nil, err
	}
	return s.ListRules(ctx)
}

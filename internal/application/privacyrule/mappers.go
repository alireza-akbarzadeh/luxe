package privacyrule

import (
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

// BuildCreateModel maps a create DTO to a privacy rule model.
func BuildCreateModel(req *dto.CreatePrivacyRuleRequest) (*models.PrivacyRule, error) {
	status := constants.PrivacyRuleStatusDraft
	if req.Status != nil && *req.Status != "" {
		status = *req.Status
	}
	locale := "en"
	if req.Locale != nil && *req.Locale != "" {
		locale = *req.Locale
	}

	rule := &models.PrivacyRule{
		Name:            req.Name,
		Key:             req.Key,
		Provider:        req.Provider,
		ContentMarkdown: req.ContentMarkdown,
		Summary:         req.Summary,
		Version:         1,
		Locale:          locale,
		Status:          status,
	}

	if req.EffectiveAt != nil && *req.EffectiveAt != "" {
		t, err := time.Parse(time.RFC3339, *req.EffectiveAt)
		if err != nil {
			return nil, err
		}
		rule.EffectiveAt = &t
	}

	return rule, nil
}

// ApplyUpdateDTO mutates a loaded privacy rule from an update DTO.
// Returns whether content changed (callers may bump version).
func ApplyUpdateDTO(rule *models.PrivacyRule, req *dto.UpdatePrivacyRuleRequest) (contentChanged bool, err error) {
	if req.Name != nil {
		rule.Name = *req.Name
	}
	if req.Key != nil {
		rule.Key = *req.Key
	}
	if req.Provider != nil {
		rule.Provider = *req.Provider
	}
	if req.ContentMarkdown != nil {
		if *req.ContentMarkdown != rule.ContentMarkdown {
			contentChanged = true
		}
		rule.ContentMarkdown = *req.ContentMarkdown
	}
	if req.Summary != nil {
		rule.Summary = req.Summary
	}
	if req.Locale != nil {
		rule.Locale = *req.Locale
	}
	if req.Status != nil {
		rule.Status = *req.Status
	}
	if req.EffectiveAt != nil {
		if *req.EffectiveAt == "" {
			rule.EffectiveAt = nil
		} else {
			t, parseErr := time.Parse(time.RFC3339, *req.EffectiveAt)
			if parseErr != nil {
				return contentChanged, parseErr
			}
			rule.EffectiveAt = &t
		}
	}
	return contentChanged, nil
}

// ToResponse maps a privacy rule model to an API response.
func ToResponse(r *models.PrivacyRule) *dto.PrivacyRuleResponse {
	resp := &dto.PrivacyRuleResponse{
		ID:              r.ID,
		Name:            r.Name,
		Key:             r.Key,
		Provider:        r.Provider,
		ContentMarkdown: r.ContentMarkdown,
		Summary:         r.Summary,
		Version:         r.Version,
		Locale:          r.Locale,
		Status:          r.Status,
		EffectiveAt:     r.EffectiveAt,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
	if r.WorkflowState != nil {
		resp.WorkflowState = dto.ToStateView(r.WorkflowState)
	}
	return resp
}

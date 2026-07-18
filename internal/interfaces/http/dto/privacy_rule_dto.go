package dto

import "time"

// CreatePrivacyRuleRequest creates a new privacy rule draft.
type CreatePrivacyRuleRequest struct {
	Name            string  `json:"name" binding:"required,min=2,max=255"`
	Key             string  `json:"key" binding:"required,min=2,max=100"`
	Provider        string  `json:"provider" binding:"required,oneof=platform stripe paypal wallet gift_card shipping ai all"`
	ContentMarkdown string  `json:"content_markdown" binding:"required"`
	Summary         *string `json:"summary" binding:"omitempty,max=2000"`
	Locale          *string `json:"locale" binding:"omitempty,max=10"`
	Status          *string `json:"status" binding:"omitempty,oneof=draft active inactive archived"`
	EffectiveAt     *string `json:"effective_at" binding:"omitempty"`
}

// UpdatePrivacyRuleRequest updates an existing privacy rule.
type UpdatePrivacyRuleRequest struct {
	Name            *string `json:"name" binding:"omitempty,min=2,max=255"`
	Key             *string `json:"key" binding:"omitempty,min=2,max=100"`
	Provider        *string `json:"provider" binding:"omitempty,oneof=platform stripe paypal wallet gift_card shipping ai all"`
	ContentMarkdown *string `json:"content_markdown"`
	Summary         *string `json:"summary" binding:"omitempty,max=2000"`
	Locale          *string `json:"locale" binding:"omitempty,max=10"`
	Status          *string `json:"status" binding:"omitempty,oneof=draft active inactive archived"`
	EffectiveAt     *string `json:"effective_at" binding:"omitempty"`
	BumpVersion     *bool   `json:"bump_version"`
}

// ListPrivacyRulesRequest filters admin or public privacy rule lists.
type ListPrivacyRulesRequest struct {
	Page     int    `form:"page,default=1"`
	Limit    int    `form:"limit,default=20"`
	Search   string `form:"search"`
	Status   string `form:"status"`
	Provider string `form:"provider"`
	Locale   string `form:"locale"`
	Key      string `form:"key"`
}

// PrivacyRuleListResponse wraps a paginated privacy rule list.
type PrivacyRuleListResponse struct {
	BaseResponse
	Data PrivacyRuleListData `json:"data"`
}

// PrivacyRuleListData is the paginated list payload.
type PrivacyRuleListData struct {
	Rules []PrivacyRuleResponse `json:"rules"`
	Total int64                 `json:"total"`
	Page  int                   `json:"page"`
	Limit int                   `json:"limit"`
}

// PrivacyRuleResponse is the API shape for a privacy rule.
type PrivacyRuleResponse struct {
	ID              uint       `json:"id"`
	Name            string     `json:"name"`
	Key             string     `json:"key"`
	Provider        string     `json:"provider"`
	ContentMarkdown string     `json:"content_markdown"`
	Summary         *string    `json:"summary,omitempty"`
	Version         int        `json:"version"`
	Locale          string     `json:"locale"`
	Status          string     `json:"status"`
	WorkflowState   *StateView `json:"workflow_state,omitempty"`
	EffectiveAt     *time.Time `json:"effective_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

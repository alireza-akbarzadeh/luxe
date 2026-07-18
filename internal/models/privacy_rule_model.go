package models

import "time"

// PrivacyRule is a versioned markdown privacy/legal rule scoped to a provider.
// Apps fetch active rules by key or provider and parse content_markdown client-side.
type PrivacyRule struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Name            string         `gorm:"size:255;not null" json:"name"`
	Key             string         `gorm:"size:100;not null" json:"key"`
	Provider        string         `gorm:"size:50;not null" json:"provider"`
	ContentMarkdown string         `gorm:"type:text;not null;default:''" json:"content_markdown"`
	Summary         *string        `gorm:"type:text" json:"summary,omitempty"`
	Version         int            `gorm:"not null;default:1" json:"version"`
	Locale          string         `gorm:"size:10;not null;default:'en'" json:"locale"`
	Status          string         `gorm:"size:50;not null;default:'draft'" json:"status"`
	WorkflowStateID *uint          `gorm:"index" json:"workflow_state_id,omitempty"`
	WorkflowState   *WorkflowState `gorm:"foreignKey:WorkflowStateID;references:ID" json:"workflow_state,omitempty"`
	EffectiveAt     *time.Time     `json:"effective_at,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

func (PrivacyRule) TableName() string {
	return "privacy_rules"
}

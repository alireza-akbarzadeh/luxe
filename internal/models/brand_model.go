package models

import "time"

type Brand struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Slug        string    `gorm:"size:255;unique;not null" json:"slug"`
	Description *string   `gorm:"type:text" json:"description"`
	LogoURL     *string   `gorm:"size:512" json:"logo_url"`
	Status      string    `gorm:"size:50;not null;default:'draft'" json:"status"`
	WorkflowStateID *uint          `gorm:"index" json:"workflow_state_id,omitempty"`
	WorkflowState   *WorkflowState `gorm:"foreignKey:WorkflowStateID;references:ID" json:"workflow_state,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Brand) TableName() string {
	return "brands"
}

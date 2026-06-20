package models

import (
	"time"

	"gorm.io/gorm"
)

// Collection is a curated merchandising group shown on the storefront collections page.
type Collection struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Slug        string `gorm:"uniqueIndex;not null" json:"slug"`
	Eyebrow     string `gorm:"not null;default:''" json:"eyebrow"`
	Title       string `gorm:"not null" json:"title"`
	Description string `gorm:"not null;default:''" json:"description"`
	Href        string `gorm:"not null;default:'/shop'" json:"href"`
	ImageURL    string `gorm:"column:image_url;not null;default:''" json:"image_url"`
	CTALabel    string `gorm:"column:cta_label;not null;default:'Shop collection'" json:"cta_label"`
	SortOrder   int    `gorm:"not null;default:0" json:"sort_order"`
	Status      string `gorm:"not null;default:'draft'" json:"status"`

	WorkflowStateID *uint          `gorm:"index" json:"workflow_state_id,omitempty"`
	WorkflowState   *WorkflowState `gorm:"foreignKey:WorkflowStateID;references:ID" json:"workflow_state,omitempty"`

	PreviewSort       string `gorm:"not null;default:''" json:"preview_sort"`
	PreviewIsNew      *bool  `json:"preview_is_new,omitempty"`
	PreviewCategoryID *uint  `gorm:"index" json:"preview_category_id,omitempty"`
}

func (Collection) TableName() string {
	return "collections"
}

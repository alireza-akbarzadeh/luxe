package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Category represents a product category (self‑referencing).
type Category struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Name            string         `gorm:"not null" json:"name" validate:"required,min=2,max=100"`
	NameI18n        datatypes.JSON `gorm:"column:name_i18n;type:jsonb" json:"nameI18n,omitempty"`
	Slug            string         `gorm:"uniqueIndex;not null" json:"slug" validate:"required,slug"`
	Description     string         `json:"description,omitempty"`
	DescriptionI18n datatypes.JSON `gorm:"column:description_i18n;type:jsonb" json:"descriptionI18n,omitempty"`
	SearchDocument  string         `gorm:"column:search_document;type:text" json:"-"`
	ParentID    *uint  `json:"parent_id,omitempty"`
	Level       int    `gorm:"default:0" json:"level"`
	Path        string `json:"path,omitempty"`
	IsActive    bool   `gorm:"default:true" json:"is_active"`
	Icon        string `gorm:"not null;default:''" json:"icon,omitempty"`
	ImageURL    string `gorm:"column:image_url;not null;default:''" json:"image_url,omitempty"`
	MetaTitle       string `gorm:"column:meta_title;not null;default:''" json:"meta_title,omitempty"`
	MetaDescription string `gorm:"column:meta_description;not null;default:''" json:"meta_description,omitempty"`
	SortOrder   int    `gorm:"column:sort_order;not null;default:0" json:"sort_order"`

	WorkflowStateID *uint          `gorm:"index" json:"workflow_state_id,omitempty"`
	WorkflowState   *WorkflowState `gorm:"foreignKey:WorkflowStateID;references:ID" json:"workflow_state,omitempty"`

	// Associations
	Parent   *Category  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Category `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Products []Product  `json:"products,omitempty"`
}

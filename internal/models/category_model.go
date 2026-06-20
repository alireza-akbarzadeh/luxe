package models

import (
	"time"

	"gorm.io/gorm"
)

// Category represents a product category (self‑referencing).
type Category struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Name        string `gorm:"not null" json:"name" validate:"required,min=2,max=100"`
	Slug        string `gorm:"uniqueIndex;not null" json:"slug" validate:"required,slug"`
	Description string `json:"description,omitempty"`
	ParentID    *uint  `json:"parent_id,omitempty"`
	Level       int    `gorm:"default:0" json:"level"`
	Path        string `json:"path,omitempty"`
	IsActive    bool   `gorm:"default:true" json:"is_active"`

	WorkflowStateID *uint          `gorm:"index" json:"workflow_state_id,omitempty"`
	WorkflowState   *WorkflowState `gorm:"foreignKey:WorkflowStateID;references:ID" json:"workflow_state,omitempty"`

	// Associations
	Parent   *Category  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Category `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Products []Product  `json:"products,omitempty"`
}

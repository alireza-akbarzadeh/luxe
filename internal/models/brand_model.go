package models

import "time"

type Brand struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	Name              string    `gorm:"size:255;not null" json:"name"`
	Slug              string    `gorm:"size:255;unique;not null" json:"slug"`
	Description       *string   `gorm:"type:text" json:"description"`
	LogoURL           *string   `gorm:"size:512" json:"logo_url"`
	Status            string    `gorm:"size:50;not null;default:'draft'" json:"status"`
	IsFeatured        bool      `gorm:"column:is_featured;not null;default:false" json:"is_featured"`
	FeaturedSortOrder int       `gorm:"column:featured_sort_order;not null;default:0" json:"featured_sort_order"`
	MetaTitle         string    `gorm:"column:meta_title;not null;default:''" json:"meta_title,omitempty"`
	MetaDescription   string    `gorm:"column:meta_description;not null;default:''" json:"meta_description,omitempty"`
	WorkflowStateID   *uint          `gorm:"index" json:"workflow_state_id,omitempty"`
	WorkflowState     *WorkflowState `gorm:"foreignKey:WorkflowStateID;references:ID" json:"workflow_state,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (Brand) TableName() string {
	return "brands"
}

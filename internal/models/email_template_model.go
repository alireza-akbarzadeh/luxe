package models

import "time"

// EmailTemplate is a reusable HTML email layout for campaigns.
type EmailTemplate struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name" validate:"required,min=2,max=255"`
	Slug      string    `gorm:"uniqueIndex;not null;size:128" json:"slug" validate:"required,min=2,max=128"`
	Subject   string    `gorm:"not null;size:512" json:"subject" validate:"required,min=2,max=512"`
	BodyHTML  string    `gorm:"type:text;not null;default:''" json:"body_html"`
	Status    string    `gorm:"size:32;not null;default:'draft'" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (EmailTemplate) TableName() string {
	return "email_templates"
}

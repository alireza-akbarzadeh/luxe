package models

import "time"

// EmailCampaign is a broadcast email send to a subscriber segment.
type EmailCampaign struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	Name           string     `gorm:"not null" json:"name" validate:"required,min=2,max=255"`
	Subject        string     `gorm:"not null;size:512" json:"subject" validate:"required,min=2,max=512"`
	BodyHTML       string     `gorm:"type:text;not null;default:''" json:"body_html"`
	TemplateID     *uint      `json:"template_id,omitempty"`
	Segment        string     `gorm:"size:64;not null;default:'all'" json:"segment"`
	Status         string     `gorm:"size:32;not null;default:'draft'" json:"status"`
	ScheduledAt    *time.Time `json:"scheduled_at,omitempty"`
	SentAt         *time.Time `json:"sent_at,omitempty"`
	RecipientCount int        `gorm:"not null;default:0" json:"recipient_count"`
	SentCount      int        `gorm:"not null;default:0" json:"sent_count"`
	FailedCount    int        `gorm:"not null;default:0" json:"failed_count"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	Template       *EmailTemplate `gorm:"foreignKey:TemplateID" json:"template,omitempty"`
}

func (EmailCampaign) TableName() string {
	return "email_campaigns"
}

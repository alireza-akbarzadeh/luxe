package models

import "time"

// NewsletterSubscriber stores marketing opt-ins from storefront, checkout, and admin.
type NewsletterSubscriber struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	Email            string     `gorm:"not null;size:255" json:"email" validate:"required,email,max=255"`
	UserID           *uint      `json:"user_id,omitempty"`
	Status           string     `gorm:"size:32;not null;default:'subscribed'" json:"status"`
	Source           string     `gorm:"size:64;not null;default:'manual'" json:"source"`
	UnsubscribeToken string     `gorm:"size:64;not null" json:"-"`
	SubscribedAt     time.Time  `json:"subscribed_at"`
	UnsubscribedAt   *time.Time `json:"unsubscribed_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	User             *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (NewsletterSubscriber) TableName() string {
	return "newsletter_subscribers"
}

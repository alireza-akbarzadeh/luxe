package models

import "time"

// PushSubscription stores a browser Web Push endpoint for a user.
type PushSubscription struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Endpoint  string    `gorm:"not null;uniqueIndex" json:"endpoint"`
	P256dh    string    `gorm:"not null" json:"p256dh"`
	Auth      string    `gorm:"not null" json:"auth"`
	UserAgent string    `gorm:"size:512" json:"user_agent,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (PushSubscription) TableName() string {
	return "push_subscriptions"
}

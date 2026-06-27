package models

import "time"

type GiftCard struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	Code            string     `gorm:"size:32;not null;uniqueIndex" json:"code"`
	SenderUserID    uint       `gorm:"not null;index" json:"sender_user_id"`
	RecipientUserID *uint      `gorm:"index" json:"recipient_user_id,omitempty"`
	RecipientEmail  string     `gorm:"size:255;not null;index" json:"recipient_email"`
	RecipientName   string     `gorm:"size:255" json:"recipient_name,omitempty"`
	SenderName      string     `gorm:"size:255" json:"sender_name,omitempty"`
	Message         string     `gorm:"type:text" json:"message,omitempty"`
	InitialAmount   float64    `gorm:"type:decimal(10,2);not null" json:"initial_amount"`
	Balance         float64    `gorm:"type:decimal(10,2);not null" json:"balance"`
	Currency        string     `gorm:"size:3;not null;default:USD" json:"currency"`
	Status          string     `gorm:"size:20;not null;default:active;index" json:"status"`
	DeliveryDate    *time.Time `json:"delivery_date,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	RedeemedAt      *time.Time `json:"redeemed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	Sender    User  `gorm:"foreignKey:SenderUserID" json:"-"`
	Recipient *User `gorm:"foreignKey:RecipientUserID" json:"-"`
}

func (GiftCard) TableName() string { return "gift_cards" }

package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Wallet struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID       uint                `gorm:"not null;uniqueIndex" json:"user_id"`
	Balance      float64             `gorm:"type:decimal(10,2);not null;default:0" json:"balance"`
	Currency     string              `gorm:"not null;default:'USD'" json:"currency"`
	User         User                `gorm:"foreignKey:UserID" json:"-"`
	Transactions []WalletTransaction `gorm:"foreignKey:UserID" json:"-"`
}

type WalletTransaction struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID        uint           `gorm:"not null;index" json:"user_id"`
	Amount        float64        `gorm:"type:decimal(10,2);not null" json:"amount"`
	Type          string         `gorm:"not null;index" json:"type"`
	ReferenceType string         `gorm:"type:text;index" json:"reference_type,omitempty"`
	ReferenceID   *uint          `json:"reference_id,omitempty"`
	Description   string         `json:"description,omitempty"`
	BalanceAfter  float64        `gorm:"type:decimal(10,2);not null" json:"balance_after"`
	Status        string         `gorm:"not null;default:'pending';index" json:"status"`
	Metadata      datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`
	User          User           `gorm:"foreignKey:UserID" json:"-"`
}

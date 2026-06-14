package models

import (
	"time"

	"gorm.io/gorm"
)

type RefreshToken struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Token      string    `gorm:"uniqueIndex;not null" json:"token"`
	UserID     uint      `gorm:"not null;index" json:"user_id"`
	User       User      `gorm:"foreignKey:UserID" json:"-"`
	ExpiresAt  time.Time `gorm:"not null" json:"expires_at"`
	Revoked    bool      `gorm:"default:false;index" json:"revoked"`
	UserAgent  string    `gorm:"type:varchar(512)" json:"user_agent"`
	IPAddress  string    `gorm:"type:varchar(45)" json:"ip_address"`
	LastUsedAt time.Time `json:"last_used_at"`
}

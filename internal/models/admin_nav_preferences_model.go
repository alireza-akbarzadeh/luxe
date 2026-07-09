package models

import (
	"time"

	"gorm.io/datatypes"
)

// AdminNavPreferences stores per-user admin navigation favorites and recent pages.
type AdminNavPreferences struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null;uniqueIndex" json:"user_id"`
	Favorites datatypes.JSON `gorm:"type:jsonb;not null;default:'[]'" json:"favorites"`
	Recent    datatypes.JSON `gorm:"type:jsonb;not null;default:'[]'" json:"recent"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

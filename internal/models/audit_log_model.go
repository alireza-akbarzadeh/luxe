package models

import (
	"time"

	"gorm.io/datatypes"
)

type AuditLog struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	UserID     uint           `gorm:"not null;index" json:"user_id"`
	Action     string         `gorm:"not null;size:32" json:"action"`
	Resource   string         `gorm:"not null;size:128;index" json:"resource"`
	ResourceID string         `gorm:"size:64" json:"resource_id,omitempty"`
	Path       string         `gorm:"not null;type:text" json:"path"`
	IPAddress  string         `gorm:"column:ip_address;size:64" json:"ip_address,omitempty"`
	RequestID  string         `gorm:"column:request_id;size:64" json:"request_id,omitempty"`
	Metadata   datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

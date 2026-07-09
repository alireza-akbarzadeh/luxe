package models

import (
	"time"

	"gorm.io/datatypes"
)

// Campaign groups merchandising placements (flash deals, banners, collections).
type Campaign struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"not null" json:"name" validate:"required,min=2,max=255"`
	Slug        string         `gorm:"uniqueIndex;not null;size:128" json:"slug" validate:"required,min=2,max=128"`
	Description string         `gorm:"not null;default:''" json:"description,omitempty"`
	StartsAt    *time.Time     `json:"starts_at,omitempty"`
	EndsAt      *time.Time     `json:"ends_at,omitempty"`
	Status      string         `gorm:"size:32;not null;default:'draft'" json:"status"`
	Placements  datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'" json:"placements,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

func (Campaign) TableName() string {
	return "campaigns"
}

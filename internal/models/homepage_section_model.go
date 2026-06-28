package models

import (
	"time"

	"gorm.io/datatypes"
)

// HomepageSection is admin-configurable merchandising block (seasonal picks, etc.).
type HomepageSection struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	SectionKey string         `gorm:"column:section_key;size:64;uniqueIndex;not null" json:"section_key"`
	Title      string         `gorm:"not null" json:"title"`
	Href       string         `gorm:"not null;default:'/shop'" json:"href"`
	ImageURL   string         `gorm:"column:image_url;not null;default:''" json:"image_url"`
	SortOrder  int            `gorm:"not null;default:0" json:"sort_order"`
	Status     string         `gorm:"not null;default:'draft'" json:"status"`
	Filters    datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'" json:"filters,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

func (HomepageSection) TableName() string {
	return "homepage_sections"
}

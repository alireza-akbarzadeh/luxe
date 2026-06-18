package models

import (
	"time"

	"gorm.io/datatypes"
)

type NavMenu struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Label     string         `gorm:"size:100;not null" json:"label"`
	Type      string         `gorm:"size:20;not null" json:"type"`
	Href      *string        `gorm:"size:500" json:"href,omitempty"`
	Badge     *string        `gorm:"size:50" json:"badge,omitempty"`
	ViewAll   datatypes.JSON `gorm:"column:view_all;type:jsonb" json:"viewAll,omitempty"`
	Columns   datatypes.JSON `gorm:"type:jsonb" json:"columns,omitempty"`
	Featured  datatypes.JSON `gorm:"type:jsonb" json:"featured,omitempty"`
	SortOrder int            `gorm:"column:order;not null;default:0" json:"order"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

func (NavMenu) TableName() string { return "nav_menus" }

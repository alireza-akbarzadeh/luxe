package models

import (
	"encoding/json"
	"time"
)

type Setting struct {
	ID          uint            `gorm:"primaryKey" json:"id"`
	Key         string          `gorm:"size:255;unique;not null" json:"key"`
	Value       json.RawMessage `gorm:"type:jsonb;not null;default:'{}'" json:"value"`
	Description *string         `gorm:"type:text" json:"description"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

func (Setting) TableName() string {
	return "settings"
}

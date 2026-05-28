package models

import "time"

type SearchLog struct {
	ID        uint   `gorm:"primarykey"`
	Query     string `gorm:"type:text;not null"`
	UserID    *uint  `gorm:"index"`
	CreatedAt time.Time
}

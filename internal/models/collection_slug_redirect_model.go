package models

import "time"

// CollectionSlugRedirect preserves old collection slugs for permanent redirects.
type CollectionSlugRedirect struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	CollectionID uint      `gorm:"not null;index" json:"collection_id"`
	OldSlug      string    `gorm:"not null;uniqueIndex" json:"old_slug"`

	Collection *Collection `gorm:"foreignKey:CollectionID;references:ID" json:"collection,omitempty"`
}

func (CollectionSlugRedirect) TableName() string {
	return "collection_slug_redirects"
}

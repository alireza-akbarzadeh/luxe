package models

import (
	"time"

	"gorm.io/gorm"
)

// Creator is a curated influencer/creator storefront profile.
type Creator struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Slug          string         `gorm:"uniqueIndex;not null" json:"slug"`
	DisplayName   string         `gorm:"column:display_name;not null" json:"display_name"`
	Handle        string         `gorm:"not null;default:''" json:"handle"`
	Bio           string         `gorm:"not null;default:''" json:"bio"`
	Specialty     string         `gorm:"not null;default:''" json:"specialty"`
	AvatarURL     string         `gorm:"column:avatar_url;not null;default:''" json:"avatar_url"`
	CoverImageURL string         `gorm:"column:cover_image_url;not null;default:''" json:"cover_image_url"`
	InstagramURL  string         `gorm:"column:instagram_url;not null;default:''" json:"instagram_url"`
	IsActive      bool           `gorm:"not null;default:true" json:"is_active"`
	SortOrder     int            `gorm:"not null;default:0" json:"sort_order"`
	Picks         []CreatorPick  `gorm:"foreignKey:CreatorID" json:"picks,omitempty"`
}

func (Creator) TableName() string {
	return "creators"
}

// CreatorPick links a product to a creator storefront with optional headline.
type CreatorPick struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatorID uint      `gorm:"not null;index" json:"creator_id"`
	ProductID uint      `gorm:"not null;index" json:"product_id"`
	Headline  string    `gorm:"not null;default:''" json:"headline"`
	SortOrder int       `gorm:"not null;default:0" json:"sort_order"`
	Product   *Product  `gorm:"foreignKey:ProductID;references:ID" json:"product,omitempty"`
}

func (CreatorPick) TableName() string {
	return "creator_picks"
}

package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Store struct {
	ID            uint           `gorm:"primarykey"`
	Name          string         `gorm:"size:255;not null"`
	Slug          string         `gorm:"size:255;uniqueIndex;not null"`
	Description   string         `gorm:"type:text"`
	LogoURL       string         `gorm:"type:text"`
	BannerURL     string         `gorm:"type:text"`
	IsVerified    bool           `gorm:"default:false"`
	Rating        float64        `gorm:"type:decimal(3,2);default:0"`
	ReviewCount   int            `gorm:"default:0"`
	FollowerCount int            `gorm:"default:0"`
	Location      string         `gorm:"size:255"`
	ShippingInfo  string         `gorm:"type:text"`
	ReturnPolicy  string         `gorm:"type:text"`
	JoinedAt      time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
	Status        string         `gorm:"size:20;default:active"`
	UserID        *uint          `gorm:"index"` // owner (admin user)
	Settings      datatypes.JSON `gorm:"type:jsonb;default:'{}'"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`

	// Relationships
	User       *User       `gorm:"foreignKey:UserID"`
	Categories []*Category `gorm:"many2many:store_categories"`
	Followers  []*User     `gorm:"many2many:store_followers"`
	Reviews    []*StoreReview
	Products   []*Product `gorm:"foreignKey:StoreID"`
}

type StoreReview struct {
	ID        uint   `gorm:"primarykey"`
	StoreID   uint   `gorm:"index;not null"`
	UserID    uint   `gorm:"index;not null"`
	Rating    int    `gorm:"not null;check:rating >= 1 AND rating <= 5"`
	Comment   string `gorm:"type:text"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Store *Store `gorm:"foreignKey:StoreID"`
	User  *User  `gorm:"foreignKey:UserID"`
}

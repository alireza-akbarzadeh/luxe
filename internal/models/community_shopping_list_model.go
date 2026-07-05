package models

import (
	"time"

	"gorm.io/gorm"
)

// CommunityShoppingList is a public community-curated shopping list.
type CommunityShoppingList struct {
	ID            uint                        `gorm:"primaryKey" json:"id"`
	CreatedAt     time.Time                   `json:"created_at"`
	UpdatedAt     time.Time                   `json:"updated_at"`
	DeletedAt     gorm.DeletedAt              `gorm:"index" json:"deleted_at,omitempty"`
	Slug          string                      `gorm:"uniqueIndex;not null" json:"slug"`
	Title         string                      `gorm:"not null" json:"title"`
	Description   string                      `gorm:"not null;default:''" json:"description"`
	Theme         string                      `gorm:"not null;default:''" json:"theme"`
	CoverImageURL string                      `gorm:"column:cover_image_url;not null;default:''" json:"cover_image_url"`
	AuthorName    string                      `gorm:"column:author_name;not null;default:''" json:"author_name"`
	AuthorHandle  string                      `gorm:"column:author_handle;not null;default:''" json:"author_handle"`
	IsActive      bool                        `gorm:"not null;default:true" json:"is_active"`
	SortOrder     int                         `gorm:"not null;default:0" json:"sort_order"`
	Items         []CommunityShoppingListItem `gorm:"foreignKey:ListID" json:"items,omitempty"`
}

func (CommunityShoppingList) TableName() string {
	return "community_shopping_lists"
}

// CommunityShoppingListItem links a product to a community shopping list.
type CommunityShoppingListItem struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ListID    uint      `gorm:"not null;index" json:"list_id"`
	ProductID uint      `gorm:"not null;index" json:"product_id"`
	Note      string    `gorm:"not null;default:''" json:"note"`
	SortOrder int       `gorm:"not null;default:0" json:"sort_order"`
	Product   *Product  `gorm:"foreignKey:ProductID;references:ID" json:"product,omitempty"`
}

func (CommunityShoppingListItem) TableName() string {
	return "community_shopping_list_items"
}

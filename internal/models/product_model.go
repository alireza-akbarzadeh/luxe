package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Product struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Name              string         `gorm:"type:text;not null" json:"name" validate:"required,min=2,max=255"`
	NameI18n          datatypes.JSON `gorm:"column:name_i18n;type:jsonb" json:"nameI18n,omitempty"`
	Description       string         `gorm:"type:text" json:"description"`
	DescriptionI18n   datatypes.JSON `gorm:"column:description_i18n;type:jsonb" json:"descriptionI18n,omitempty"`
	SearchAliases     datatypes.JSON `gorm:"column:search_aliases;type:jsonb;default:'[]'" json:"searchAliases,omitempty"`
	SearchDocument    string         `gorm:"column:search_document;type:text" json:"-"`
	Price             float64        `gorm:"type:decimal(10,2);not null;check:price >= 0" json:"price" validate:"required,gt=0"`
	Stock             int            `gorm:"not null;default:0;check:stock >= 0" json:"stock" validate:"gte=0"`
	SKU               string         `gorm:"type:text;not null;uniqueIndex" json:"sku" validate:"required"`
	CategoryID        *uint          `gorm:"index" json:"category_id,omitempty"`
	Category          *Category      `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Images            pq.StringArray `gorm:"type:text[]" json:"images"`
	Status            string         `gorm:"type:text;not null;default:'active'" json:"status"`
	WorkflowStateID   *uint          `gorm:"index" json:"workflow_state_id,omitempty"`
	WorkflowState     *WorkflowState `gorm:"foreignKey:WorkflowStateID;references:ID" json:"workflow_state,omitempty"`
	CreatedBy         *uint          `gorm:"index" json:"created_by,omitempty"`
	UpdatedBy         *uint          `gorm:"index" json:"updated_by,omitempty"`
	Rating            float64        `gorm:"type:decimal(3,2);default:0.0" json:"rating"`
	ReviewsCount      int            `gorm:"default:0" json:"reviews_count"`
	IsNew             bool           `gorm:"default:false" json:"is_new"`
	Slug              string         `gorm:"type:text;not null;uniqueIndex" json:"slug" validate:"required"`
	CompareAtPrice    *float64       `gorm:"type:decimal(10,2);check:compare_at_price >= 0" json:"compare_at_price,omitempty"`
	Cost              *float64       `gorm:"type:decimal(10,2);check:cost >= 0" json:"cost,omitempty"`
	Barcode           string         `gorm:"type:text" json:"barcode"`
	LowStockThreshold int            `gorm:"default:5" json:"low_stock_threshold"`
	Weight            *float64       `gorm:"type:decimal(8,2);check:weight >= 0" json:"weight,omitempty"`
	IsDigital         bool           `gorm:"not null;default:false;index" json:"is_digital"`
	MetaTitle         string         `gorm:"type:text" json:"meta_title"`
	MetaDescription   string         `gorm:"type:text" json:"meta_description"`
	Colors            []string       `gorm:"type:jsonb;default:'[]';serializer:json" json:"colors"`
	Sizes             []string       `gorm:"type:jsonb;default:'[]';serializer:json" json:"sizes"`

	StoreID uint   `gorm:"index"`
	Store   *Store `gorm:"foreignKey:StoreID;references:ID"`

	BrandID    *uint              `gorm:"index" json:"brand_id,omitempty"`
	Brand      *Brand             `gorm:"foreignKey:BrandID" json:"brand,omitempty"`
	Attributes []ProductAttribute `gorm:"foreignKey:ProductID" json:"attributes,omitempty"`

	// Inventory extras
	TrackInventory    bool   `gorm:"not null;default:true" json:"track_inventory"`
	WarehouseLocation string `gorm:"type:text" json:"warehouse_location,omitempty"`
	AllowBackorder    bool   `gorm:"not null;default:false" json:"allow_backorder"`

	// Publishing extras
	Visibility  string         `gorm:"type:text;not null;default:'public';check:visibility IN ('public','private')" json:"visibility"`
	Tags        pq.StringArray `gorm:"type:text[];default:'{}'" json:"tags"`
	Channels    pq.StringArray `gorm:"type:text[];default:'{}'" json:"channels"`
	PublishedAt *time.Time     `json:"published_at,omitempty"`
}

type ProductLike struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	UserID    uint `gorm:"not null;uniqueIndex:idx_user_product" json:"user_id"`
	ProductID uint `gorm:"not null;uniqueIndex:idx_user_product" json:"product_id"`

	User    User    `gorm:"foreignKey:UserID" json:"-"`
	Product Product `gorm:"foreignKey:ProductID" json:"-"`
}

type ProductAttribute struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	ProductID uint           `gorm:"not null;index" json:"product_id"`
	Name      string         `gorm:"type:text;not null" json:"name" validate:"required,min=1,max=100"`
	Values    pq.StringArray `gorm:"type:text[];not null" json:"values" validate:"required,min=1"`
}

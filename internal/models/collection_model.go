package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Collection is a curated merchandising group shown on the storefront collections page.
type Collection struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Slug        string `gorm:"uniqueIndex;not null" json:"slug"`
	Eyebrow     string `gorm:"not null;default:''" json:"eyebrow"`
	Title       string `gorm:"not null" json:"title"`
	Subtitle    string `gorm:"not null;default:''" json:"subtitle"`
	Description string `gorm:"not null;default:''" json:"description"`
	Href        string `gorm:"not null;default:'/shop'" json:"href"`
	ImageURL    string `gorm:"column:image_url;not null;default:''" json:"image_url"`
	CTALabel    string `gorm:"column:cta_label;not null;default:'Shop collection'" json:"cta_label"`
	SortOrder   int    `gorm:"not null;default:0" json:"sort_order"`
	Status      string `gorm:"not null;default:'draft'" json:"status"`

	WorkflowStateID *uint          `gorm:"index" json:"workflow_state_id,omitempty"`
	WorkflowState   *WorkflowState `gorm:"foreignKey:WorkflowStateID;references:ID" json:"workflow_state,omitempty"`

	PreviewSort       string `gorm:"not null;default:''" json:"preview_sort"`
	PreviewIsNew      *bool  `json:"preview_is_new,omitempty"`
	PreviewCategoryID *uint  `gorm:"index" json:"preview_category_id,omitempty"`
	Theme             string `gorm:"not null;default:'';index" json:"theme"`

	// CollectionType is a legacy compatibility field retained for existing clients.
	CollectionType string `gorm:"not null;default:'smart';index" json:"collection_type"`
	Mode           string `gorm:"not null;default:'dynamic';index" json:"mode"`
	StartsAt       *time.Time `json:"starts_at,omitempty"`
	EndsAt         *time.Time `json:"ends_at,omitempty"`
	SortKey        string `gorm:"not null;default:''" json:"sort_key"`
	RulesJSON      datatypes.JSON `gorm:"column:rules_json;type:jsonb;default:'{}'" json:"rules_json,omitempty"`

	SEOTitle         string `gorm:"column:seo_title;not null;default:''" json:"seo_title"`
	SEODescription   string `gorm:"column:seo_description;not null;default:''" json:"seo_description"`
	MetaKeywords     string `gorm:"column:meta_keywords;not null;default:''" json:"meta_keywords"`
	OGTitle          string `gorm:"column:og_title;not null;default:''" json:"og_title"`
	OGDescription    string `gorm:"column:og_description;not null;default:''" json:"og_description"`
	OGImageURL       string `gorm:"column:og_image_url;not null;default:''" json:"og_image_url"`
	TwitterTitle     string `gorm:"column:twitter_title;not null;default:''" json:"twitter_title"`
	TwitterDescription string `gorm:"column:twitter_description;not null;default:''" json:"twitter_description"`
	TwitterImageURL  string `gorm:"column:twitter_image_url;not null;default:''" json:"twitter_image_url"`
	CanonicalURL     string `gorm:"column:canonical_url;not null;default:''" json:"canonical_url"`
	RobotsDirectives string `gorm:"column:robots_directives;not null;default:''" json:"robots_directives"`
	IsIndexable      bool   `gorm:"column:is_indexable;not null;default:true" json:"is_indexable"`

	HeroTitle       string  `gorm:"column:hero_title;not null;default:''" json:"hero_title"`
	HeroDescription string  `gorm:"column:hero_description;not null;default:''" json:"hero_description"`
	DesktopImageURL string  `gorm:"column:desktop_image_url;not null;default:''" json:"desktop_image_url"`
	TabletImageURL  string  `gorm:"column:tablet_image_url;not null;default:''" json:"tablet_image_url"`
	MobileImageURL  string  `gorm:"column:mobile_image_url;not null;default:''" json:"mobile_image_url"`
	OverlayOpacity  float64 `gorm:"column:overlay_opacity;not null;default:0.25" json:"overlay_opacity"`
	ThemeVariant    string  `gorm:"column:theme_variant;not null;default:''" json:"theme_variant"`

	Products []CollectionProduct `gorm:"foreignKey:CollectionID" json:"products,omitempty"`
}

func (Collection) TableName() string {
	return "collections"
}

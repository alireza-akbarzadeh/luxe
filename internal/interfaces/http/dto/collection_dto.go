package dto

import "time"

type CollectionRuleCondition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"` // eq, neq, gte, lte, contains, in
	Value    any    `json:"value,omitempty"`
}

// CollectionRuleGroup is a nested AND/OR group (one nesting level preferred).
type CollectionRuleGroup struct {
	Operator   string                    `json:"operator"` // and | or
	Conditions []CollectionRuleCondition `json:"conditions,omitempty"`
	Groups     []CollectionRuleGroup     `json:"groups,omitempty"`
}

type CollectionRules struct {
	Operator   string                    `json:"operator"` // and | or
	Conditions []CollectionRuleCondition `json:"conditions,omitempty"`
	Groups     []CollectionRuleGroup     `json:"groups,omitempty"`
}

type CollectionProductOverrideInput struct {
	ProductID  uint `json:"product_id" validate:"required,gt=0"`
	Position   int  `json:"position"`
	IsPinned   bool `json:"is_pinned"`
	IsHidden   bool `json:"is_hidden"`
	BoostScore int  `json:"boost_score"`
}

type CreateCollectionRequest struct {
	Slug               string                           `json:"slug" validate:"required,min=2,max=128"`
	Eyebrow            string                           `json:"eyebrow" validate:"omitempty,max=128"`
	Title              string                           `json:"title" validate:"required,min=2,max=255"`
	Subtitle           string                           `json:"subtitle" validate:"omitempty,max=255"`
	Description        string                           `json:"description" validate:"omitempty,max=2000"`
	Href               string                           `json:"href" validate:"omitempty,max=512"`
	ImageURL           string                           `json:"image_url" validate:"omitempty,max=2048"`
	CTALabel           string                           `json:"cta_label" validate:"omitempty,max=128"`
	SortOrder          int                              `json:"sort_order"`
	Status             string                           `json:"status" validate:"omitempty,oneof=draft scheduled active inactive archived"`
	CollectionType     string                           `json:"collection_type" validate:"omitempty,oneof=manual smart"`
	Mode               string                           `json:"mode" validate:"omitempty,oneof=manual dynamic hybrid"`
	StartsAt           *string                          `json:"starts_at"`
	EndsAt             *string                          `json:"ends_at"`
	ProductIDs         []uint                           `json:"product_ids"`
	ProductOverrides   []CollectionProductOverrideInput `json:"product_overrides,omitempty"`
	PreviewSort        string                           `json:"preview_sort" validate:"omitempty,max=64"`
	PreviewIsNew       *bool                            `json:"preview_is_new"`
	PreviewCategoryID  *uint                            `json:"preview_category_id"`
	Theme              string                           `json:"theme" validate:"omitempty,max=64"`
	SortKey            string                           `json:"sort_key" validate:"omitempty,max=64"`
	Rules              *CollectionRules                 `json:"rules,omitempty"`
	SEOTitle           string                           `json:"seo_title" validate:"omitempty,max=255"`
	SEODescription     string                           `json:"seo_description" validate:"omitempty,max=500"`
	MetaKeywords       string                           `json:"meta_keywords" validate:"omitempty,max=500"`
	OGTitle            string                           `json:"og_title" validate:"omitempty,max=255"`
	OGDescription      string                           `json:"og_description" validate:"omitempty,max=500"`
	OGImageURL         string                           `json:"og_image_url" validate:"omitempty,max=2048"`
	TwitterTitle       string                           `json:"twitter_title" validate:"omitempty,max=255"`
	TwitterDescription string                           `json:"twitter_description" validate:"omitempty,max=500"`
	TwitterImageURL    string                           `json:"twitter_image_url" validate:"omitempty,max=2048"`
	CanonicalURL       string                           `json:"canonical_url" validate:"omitempty,max=2048"`
	RobotsDirectives   string                           `json:"robots_directives" validate:"omitempty,max=255"`
	IsIndexable        *bool                            `json:"is_indexable"`
	HeroTitle          string                           `json:"hero_title" validate:"omitempty,max=255"`
	HeroDescription    string                           `json:"hero_description" validate:"omitempty,max=2000"`
	DesktopImageURL    string                           `json:"desktop_image_url" validate:"omitempty,max=2048"`
	TabletImageURL     string                           `json:"tablet_image_url" validate:"omitempty,max=2048"`
	MobileImageURL     string                           `json:"mobile_image_url" validate:"omitempty,max=2048"`
	OverlayOpacity     *float64                         `json:"overlay_opacity"`
	ThemeVariant       string                           `json:"theme_variant" validate:"omitempty,max=64"`
}

type UpdateCollectionRequest struct {
	Slug               *string                           `json:"slug" validate:"omitempty,min=2,max=128"`
	Eyebrow            *string                           `json:"eyebrow" validate:"omitempty,max=128"`
	Title              *string                           `json:"title" validate:"omitempty,min=2,max=255"`
	Subtitle           *string                           `json:"subtitle" validate:"omitempty,max=255"`
	Description        *string                           `json:"description" validate:"omitempty,max=2000"`
	Href               *string                           `json:"href" validate:"omitempty,max=512"`
	ImageURL           *string                           `json:"image_url" validate:"omitempty,max=2048"`
	CTALabel           *string                           `json:"cta_label" validate:"omitempty,max=128"`
	SortOrder          *int                              `json:"sort_order"`
	Status             *string                           `json:"status" validate:"omitempty,oneof=draft scheduled active inactive archived"`
	CollectionType     *string                           `json:"collection_type" validate:"omitempty,oneof=manual smart"`
	Mode               *string                           `json:"mode" validate:"omitempty,oneof=manual dynamic hybrid"`
	StartsAt           *string                           `json:"starts_at"`
	EndsAt             *string                           `json:"ends_at"`
	ProductIDs         *[]uint                           `json:"product_ids"`
	ProductOverrides   *[]CollectionProductOverrideInput `json:"product_overrides"`
	PreviewSort        *string                           `json:"preview_sort" validate:"omitempty,max=64"`
	PreviewIsNew       *bool                             `json:"preview_is_new"`
	PreviewCategoryID  *uint                             `json:"preview_category_id"`
	Theme              *string                           `json:"theme" validate:"omitempty,max=64"`
	SortKey            *string                           `json:"sort_key" validate:"omitempty,max=64"`
	Rules              *CollectionRules                  `json:"rules"`
	SEOTitle           *string                           `json:"seo_title" validate:"omitempty,max=255"`
	SEODescription     *string                           `json:"seo_description" validate:"omitempty,max=500"`
	MetaKeywords       *string                           `json:"meta_keywords" validate:"omitempty,max=500"`
	OGTitle            *string                           `json:"og_title" validate:"omitempty,max=255"`
	OGDescription      *string                           `json:"og_description" validate:"omitempty,max=500"`
	OGImageURL         *string                           `json:"og_image_url" validate:"omitempty,max=2048"`
	TwitterTitle       *string                           `json:"twitter_title" validate:"omitempty,max=255"`
	TwitterDescription *string                           `json:"twitter_description" validate:"omitempty,max=500"`
	TwitterImageURL    *string                           `json:"twitter_image_url" validate:"omitempty,max=2048"`
	CanonicalURL       *string                           `json:"canonical_url" validate:"omitempty,max=2048"`
	RobotsDirectives   *string                           `json:"robots_directives" validate:"omitempty,max=255"`
	IsIndexable        *bool                             `json:"is_indexable"`
	HeroTitle          *string                           `json:"hero_title" validate:"omitempty,max=255"`
	HeroDescription    *string                           `json:"hero_description" validate:"omitempty,max=2000"`
	DesktopImageURL    *string                           `json:"desktop_image_url" validate:"omitempty,max=2048"`
	TabletImageURL     *string                           `json:"tablet_image_url" validate:"omitempty,max=2048"`
	MobileImageURL     *string                           `json:"mobile_image_url" validate:"omitempty,max=2048"`
	OverlayOpacity     *float64                          `json:"overlay_opacity"`
	ThemeVariant       *string                           `json:"theme_variant" validate:"omitempty,max=64"`
}

type ListCollectionsRequest struct {
	Page           int    `form:"page,default=1"`
	Limit          int    `form:"limit,default=20"`
	Search         string `form:"search"`
	Status         string `form:"status"`
	Theme          string `form:"theme"`
	CollectionType string `form:"collection_type"`
	Mode           string `form:"mode"`
	LiveOnly       bool   `form:"live_only"`
}

type CollectionProductsRequest struct {
	Limit      int     `form:"limit,default=12"`
	Offset     int     `form:"offset,default=0"`
	Search     string  `form:"search"`
	CategoryID uint    `form:"category_id"`
	MinPrice   float64 `form:"min_price"`
	MaxPrice   float64 `form:"max_price"`
	MinRating  float64 `form:"min_rating"`
	Sort       string  `form:"sort"`
	InStock    *bool   `form:"in_stock"`
	OnSale     *bool   `form:"on_sale"`
	IsNew      *bool   `form:"is_new"`
}

type CollectionRulesValidationRequest struct {
	Mode      string                           `json:"mode" validate:"omitempty,oneof=manual dynamic hybrid"`
	Rules     *CollectionRules                 `json:"rules,omitempty"`
	Overrides []CollectionProductOverrideInput `json:"overrides,omitempty"`
	Limit     int                              `json:"limit" validate:"omitempty,min=1,max=24"`
}

type CollectionResponse struct {
	ID                 uint                             `json:"id"`
	Slug               string                           `json:"slug"`
	Eyebrow            string                           `json:"eyebrow"`
	Title              string                           `json:"title"`
	Subtitle           string                           `json:"subtitle"`
	Description        string                           `json:"description"`
	Href               string                           `json:"href"`
	ImageURL           string                           `json:"image_url"`
	CTALabel           string                           `json:"cta_label"`
	SortOrder          int                              `json:"sort_order"`
	Status             string                           `json:"status"`
	CollectionType     string                           `json:"collection_type"`
	Mode               string                           `json:"mode"`
	StartsAt           *time.Time                       `json:"starts_at,omitempty"`
	EndsAt             *time.Time                       `json:"ends_at,omitempty"`
	ProductIDs         []uint                           `json:"product_ids,omitempty"`
	ProductOverrides   []CollectionProductOverrideInput `json:"product_overrides,omitempty"`
	WorkflowState      *StateView                       `json:"workflow_state,omitempty"`
	PreviewSort        string                           `json:"preview_sort"`
	PreviewIsNew       *bool                            `json:"preview_is_new,omitempty"`
	PreviewCategoryID  *uint                            `json:"preview_category_id,omitempty"`
	Theme              string                           `json:"theme,omitempty"`
	SortKey            string                           `json:"sort_key,omitempty"`
	Rules              *CollectionRules                 `json:"rules,omitempty"`
	SEOTitle           string                           `json:"seo_title,omitempty"`
	SEODescription     string                           `json:"seo_description,omitempty"`
	MetaKeywords       string                           `json:"meta_keywords,omitempty"`
	OGTitle            string                           `json:"og_title,omitempty"`
	OGDescription      string                           `json:"og_description,omitempty"`
	OGImageURL         string                           `json:"og_image_url,omitempty"`
	TwitterTitle       string                           `json:"twitter_title,omitempty"`
	TwitterDescription string                           `json:"twitter_description,omitempty"`
	TwitterImageURL    string                           `json:"twitter_image_url,omitempty"`
	CanonicalURL       string                           `json:"canonical_url,omitempty"`
	RobotsDirectives   string                           `json:"robots_directives,omitempty"`
	IsIndexable        bool                             `json:"is_indexable"`
	HeroTitle          string                           `json:"hero_title,omitempty"`
	HeroDescription    string                           `json:"hero_description,omitempty"`
	DesktopImageURL    string                           `json:"desktop_image_url,omitempty"`
	TabletImageURL     string                           `json:"tablet_image_url,omitempty"`
	MobileImageURL     string                           `json:"mobile_image_url,omitempty"`
	OverlayOpacity     float64                          `json:"overlay_opacity"`
	ThemeVariant       string                           `json:"theme_variant,omitempty"`
	CreatedAt          time.Time                        `json:"created_at"`
	UpdatedAt          time.Time                        `json:"updated_at"`
}

type CollectionListData struct {
	Collections []CollectionResponse `json:"collections"`
	Total       int64                `json:"total"`
	Page        int                  `json:"page"`
	Limit       int                  `json:"limit"`
}

type CollectionListResponse struct {
	BaseResponse
	Data CollectionListData `json:"data"`
}

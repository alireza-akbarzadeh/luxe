package dto

import "time"

type CreateCollectionRequest struct {
	Slug              string     `json:"slug" validate:"required,min=2,max=128"`
	Eyebrow           string     `json:"eyebrow" validate:"omitempty,max=128"`
	Title             string     `json:"title" validate:"required,min=2,max=255"`
	Description       string     `json:"description" validate:"omitempty,max=2000"`
	Href              string     `json:"href" validate:"omitempty,max=512"`
	ImageURL          string     `json:"image_url" validate:"omitempty,max=2048"`
	CTALabel          string     `json:"cta_label" validate:"omitempty,max=128"`
	SortOrder         int        `json:"sort_order"`
	Status            string     `json:"status" validate:"omitempty,oneof=draft active inactive archived"`
	CollectionType    string   `json:"collection_type" validate:"omitempty,oneof=manual smart"`
	StartsAt          *string  `json:"starts_at"`
	EndsAt            *string  `json:"ends_at"`
	ProductIDs        []uint     `json:"product_ids"`
	PreviewSort       string     `json:"preview_sort" validate:"omitempty,max=64"`
	PreviewIsNew      *bool      `json:"preview_is_new"`
	PreviewCategoryID *uint      `json:"preview_category_id"`
	Theme             string     `json:"theme" validate:"omitempty,max=64"`
}

type UpdateCollectionRequest struct {
	Slug              *string    `json:"slug" validate:"omitempty,min=2,max=128"`
	Eyebrow           *string    `json:"eyebrow" validate:"omitempty,max=128"`
	Title             *string    `json:"title" validate:"omitempty,min=2,max=255"`
	Description       *string    `json:"description" validate:"omitempty,max=2000"`
	Href              *string    `json:"href" validate:"omitempty,max=512"`
	ImageURL          *string    `json:"image_url" validate:"omitempty,max=2048"`
	CTALabel          *string    `json:"cta_label" validate:"omitempty,max=128"`
	SortOrder         *int       `json:"sort_order"`
	Status            *string    `json:"status" validate:"omitempty,oneof=draft active inactive archived"`
	CollectionType    *string  `json:"collection_type" validate:"omitempty,oneof=manual smart"`
	StartsAt          *string  `json:"starts_at"`
	EndsAt            *string  `json:"ends_at"`
	ProductIDs        *[]uint    `json:"product_ids"`
	PreviewSort       *string    `json:"preview_sort" validate:"omitempty,max=64"`
	PreviewIsNew      *bool      `json:"preview_is_new"`
	PreviewCategoryID *uint      `json:"preview_category_id"`
	Theme             *string    `json:"theme" validate:"omitempty,max=64"`
}

type ListCollectionsRequest struct {
	Page           int    `form:"page,default=1"`
	Limit          int    `form:"limit,default=20"`
	Search         string `form:"search"`
	Status         string `form:"status"`
	Theme          string `form:"theme"`
	CollectionType string `form:"collection_type"`
	LiveOnly       bool   `form:"live_only"`
}

type CollectionResponse struct {
	ID                uint       `json:"id"`
	Slug              string     `json:"slug"`
	Eyebrow           string     `json:"eyebrow"`
	Title             string     `json:"title"`
	Description       string     `json:"description"`
	Href              string     `json:"href"`
	ImageURL          string     `json:"image_url"`
	CTALabel          string     `json:"cta_label"`
	SortOrder         int        `json:"sort_order"`
	Status            string     `json:"status"`
	CollectionType    string     `json:"collection_type"`
	StartsAt          *time.Time `json:"starts_at,omitempty"`
	EndsAt            *time.Time `json:"ends_at,omitempty"`
	ProductIDs        []uint     `json:"product_ids,omitempty"`
	WorkflowState     *StateView `json:"workflow_state,omitempty"`
	PreviewSort       string     `json:"preview_sort"`
	PreviewIsNew      *bool      `json:"preview_is_new,omitempty"`
	PreviewCategoryID *uint      `json:"preview_category_id,omitempty"`
	Theme             string     `json:"theme,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
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

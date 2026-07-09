package dto

import (
	"github.com/alireza-akbarzadeh/luxe/internal/i18n"
)

type CategoryListFilters struct {
	Limit    int    `form:"limit" validate:"omitempty,min=1,max=100"`
	Offset   int    `form:"offset" validate:"omitempty,min=0"`
	IsActive *bool  `form:"is_active"`
	ParentID *uint  `form:"parent_id" validate:"omitempty,gt=0"`
	Sort     string `form:"sort"`
	Search   string `form:"search"`
}

type CreateCategoryRequest struct {
	Name            string            `json:"name" validate:"required,min=2,max=100"`
	NameI18n        i18n.LocalizedMap `json:"nameI18n,omitempty"`
	Slug            string            `json:"slug"`
	Description     string            `json:"description,omitempty"`
	DescriptionI18n i18n.LocalizedMap `json:"descriptionI18n,omitempty"`
	ParentID        *uint             `json:"parent_id,omitempty"`
	IsActive        bool              `json:"is_active"`
	Icon            string            `json:"icon,omitempty" validate:"omitempty,max=64"`
	ImageURL        string            `json:"image_url,omitempty" validate:"omitempty,max=512"`
	MetaTitle       string            `json:"meta_title,omitempty" validate:"omitempty,max=70"`
	MetaDescription string            `json:"meta_description,omitempty" validate:"omitempty,max=160"`
	SortOrder       *int              `json:"sort_order,omitempty" validate:"omitempty,min=0"`
}

type UpdateCategoryRequest struct {
	Name            *string           `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	NameI18n        i18n.LocalizedMap `json:"nameI18n,omitempty"`
	Slug            *string           `json:"slug,omitempty"`
	Description     *string           `json:"description,omitempty"`
	DescriptionI18n i18n.LocalizedMap `json:"descriptionI18n,omitempty"`
	ParentID        *uint             `json:"parent_id,omitempty"`
	IsActive        *bool             `json:"is_active,omitempty"`
	Icon            *string           `json:"icon,omitempty" validate:"omitempty,max=64"`
	ImageURL        *string           `json:"image_url,omitempty" validate:"omitempty,max=512"`
	MetaTitle       *string           `json:"meta_title,omitempty" validate:"omitempty,max=70"`
	MetaDescription *string           `json:"meta_description,omitempty" validate:"omitempty,max=160"`
	SortOrder       *int              `json:"sort_order,omitempty" validate:"omitempty,min=0"`
}

type ReorderCategoryItem struct {
	ID        uint  `json:"id" validate:"required,gt=0"`
	ParentID  *uint `json:"parent_id"`
	SortOrder int   `json:"sort_order" validate:"min=0"`
}

type ReorderCategoriesRequest struct {
	Items []ReorderCategoryItem `json:"items" validate:"required,min=1,dive"`
}

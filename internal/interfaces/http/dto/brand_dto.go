package dto

import "time"

type CreateBrandRequest struct {
	Name              string  `json:"name" binding:"required"`
	Slug              string  `json:"slug" binding:"required"`
	Description       *string `json:"description"`
	LogoURL           *string `json:"logo_url"`
	Status            *string `json:"status"`
	IsFeatured        *bool   `json:"is_featured"`
	FeaturedSortOrder *int    `json:"featured_sort_order" binding:"omitempty,min=0"`
	MetaTitle         *string `json:"meta_title" binding:"omitempty,max=70"`
	MetaDescription   *string `json:"meta_description" binding:"omitempty,max=160"`
}

type UpdateBrandRequest struct {
	Name              *string `json:"name"`
	Slug              *string `json:"slug"`
	Description       *string `json:"description"`
	LogoURL           *string `json:"logo_url"`
	Status            *string `json:"status"`
	IsFeatured        *bool   `json:"is_featured"`
	FeaturedSortOrder *int    `json:"featured_sort_order" binding:"omitempty,min=0"`
	MetaTitle         *string `json:"meta_title" binding:"omitempty,max=70"`
	MetaDescription   *string `json:"meta_description" binding:"omitempty,max=160"`
}

type ListBrandsRequest struct {
	Page     int    `form:"page,default=1"`
	Limit    int    `form:"limit,default=20"`
	Search   string `form:"search"`
	Status   string `form:"status"`
	Featured *bool  `form:"featured"`
}

type BrandListResponse struct {
	BaseResponse
	Data BrandListData `json:"data"`
}

type BrandListData struct {
	Brands []BrandResponse `json:"brands"`
	Total  int64           `json:"total"`
	Page   int             `json:"page"`
	Limit  int             `json:"limit"`
}

type BrandResponse struct {
	ID                uint       `json:"id"`
	Name              string     `json:"name"`
	Slug              string     `json:"slug"`
	Description       *string    `json:"description,omitempty"`
	LogoURL           *string    `json:"logo_url,omitempty"`
	Status            string     `json:"status"`
	IsFeatured        bool       `json:"is_featured"`
	FeaturedSortOrder int        `json:"featured_sort_order"`
	MetaTitle         string     `json:"meta_title,omitempty"`
	MetaDescription   string     `json:"meta_description,omitempty"`
	ProductCount      *int64     `json:"product_count,omitempty"`
	WorkflowState     *StateView `json:"workflow_state,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

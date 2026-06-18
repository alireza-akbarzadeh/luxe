package dto

import "time"

type CreateBrandRequest struct {
	Name        string  `json:"name" binding:"required"`
	Slug        string  `json:"slug" binding:"required"`
	Description *string `json:"description"`
	LogoURL     *string `json:"logo_url"`
	Status      *string `json:"status"`
}

type UpdateBrandRequest struct {
	Name        *string `json:"name"`
	Slug        *string `json:"slug"`
	Description *string `json:"description"`
	LogoURL     *string `json:"logo_url"`
	Status      *string `json:"status"`
}

type ListBrandsRequest struct {
	Page   int    `form:"page,default=1"`
	Limit  int    `form:"limit,default=20"`
	Search string `form:"search"`
	Status string `form:"status"`
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
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description *string   `json:"description,omitempty"`
	LogoURL     *string   `json:"logo_url,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

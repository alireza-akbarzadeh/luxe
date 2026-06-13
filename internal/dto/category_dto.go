package dto

type CategoryListFilters struct {
	Limit    int    `form:"limit" validate:"omitempty,min=1,max=100"`
	Offset   int    `form:"offset" validate:"omitempty,min=0"`
	IsActive *bool  `form:"is_active"`
	ParentID *uint  `form:"parent_id" validate:"omitempty,gt=0"`
	Sort     string `form:"sort"`
	Search   string `form:"search"`
}

type CreateCategoryRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Slug        string `json:"slug"`
	Description string `json:"description,omitempty"`
	ParentID    *uint  `json:"parent_id,omitempty"`
	IsActive    bool   `json:"is_active"`
}

type UpdateCategoryRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Slug        *string `json:"slug,omitempty"`
	Description *string `json:"description,omitempty"`
	ParentID    *uint   `json:"parent_id,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

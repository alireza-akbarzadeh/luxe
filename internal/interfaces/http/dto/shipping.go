package dto

type CreateShippingProviderRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"min=0"`
	IsActive    *bool   `json:"is_active"`
}

type UpdateShippingProviderRequest struct {
	Name        *string  `json:"name" validate:"omitempty"`
	Description *string  `json:"description"`
	Price       *float64 `json:"price" validate:"omitempty,min=0"`
	IsActive    *bool    `json:"is_active"`
}

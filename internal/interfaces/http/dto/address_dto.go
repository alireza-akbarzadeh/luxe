package dto

import "github.com/alireza-akbarzadeh/luxe/internal/models"

type CreateAddressRequest struct {
	AddressType   string `json:"address_type" validate:"required,oneof=shipping billing both"`
	IsDefault     bool   `json:"is_default"`
	RecipientName string `json:"recipient_name" validate:"required"`
	Phone         string `json:"phone" validate:"required,e164"`
	AddressLine1  string `json:"address_line1" validate:"required"`
	AddressLine2  string `json:"address_line2"`
	City          string `json:"city" validate:"required"`
	State         string `json:"state"`
	PostalCode    string `json:"postal_code" validate:"required"`
	Country       string `json:"country" validate:"required"`
	Instructions  string `json:"instructions"`
}

type UpdateAddressRequest struct {
	AddressType   *string `json:"address_type,omitempty" validate:"omitempty,oneof=shipping billing both"`
	IsDefault     *bool   `json:"is_default,omitempty"`
	RecipientName *string `json:"recipient_name,omitempty" validate:"omitempty,required"`
	Phone         *string `json:"phone,omitempty" validate:"omitempty,e164"`
	AddressLine1  *string `json:"address_line1,omitempty" validate:"omitempty,required"`
	AddressLine2  *string `json:"address_line2,omitempty"`
	City          *string `json:"city,omitempty" validate:"omitempty,required"`
	State         *string `json:"state,omitempty"`
	PostalCode    *string `json:"postal_code,omitempty" validate:"omitempty,required"`
	Country       *string `json:"country,omitempty" validate:"omitempty,required"`
	Instructions  *string `json:"instructions,omitempty"`
}

// AddressSingleResponse is the response for single address operations
type AddressSingleResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Code    int            `json:"code"`
	Address models.Address `json:"address"`
}

// AddressListResponse is the response for listing addresses
type AddressListResponse struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Code    int              `json:"code"`
	Address []models.Address `json:"address"`
}

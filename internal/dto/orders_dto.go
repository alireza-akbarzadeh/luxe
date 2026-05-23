package dto

import (
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

type CheckoutRequest struct {
	CouponCode     string `json:"coupon_code,omitempty"`
	Email          string `json:"email" validate:"required,email"`
	FirstName      string `json:"first_name" validate:"required"`
	LastName       string `json:"last_name" validate:"required"`
	AddressLine1   string `json:"address_line1" validate:"required"`
	AddressLine2   string `json:"address_line2"`
	City           string `json:"city" validate:"required"`
	State          string `json:"state" validate:"required"`
	Zip            string `json:"zip" validate:"required"`
	Country        string `json:"country" validate:"required"`
	Phone          string `json:"phone" validate:"required"`
	ShippingMethod string `json:"shipping_method"`
	PaymentMethod  string `json:"payment_method"`
	SaveInfo       bool   `json:"save_info"`
	Newsletter     bool   `json:"newsletter"`
	CardLast4      string `json:"card_last4,omitempty"`

	ShippingProviderID *uint  `json:"shipping_provider_id,omitempty"`
	CardNumber         string `json:"card_number" validate:"required,len=16"`
	ExpiryMonth        int    `json:"expiry_month" validate:"required,min=1,max=12"`
	ExpiryYear         int    `json:"expiry_year" validate:"required,min=2025"`
	CVV                string `json:"cvv" validate:"required,len=3"`
}

// CardInfo used internally (no JSON tags needed)
type CardInfo struct {
	CardNumber  string
	ExpiryMonth int
	ExpiryYear  int
	CVV         string
}

func MapAddress(userID uint, req CheckoutRequest) models.Address {
	recipientName := req.FirstName + " " + req.LastName
	return models.Address{
		UserID:        userID,
		AddressType:   "shipping",
		IsDefault:     req.SaveInfo,
		RecipientName: recipientName,
		Phone:         req.Phone,
		AddressLine1:  req.AddressLine1,
		AddressLine2:  req.AddressLine2,
		City:          req.City,
		State:         req.State,
		PostalCode:    req.Zip,
		Country:       req.Country,
		Instructions:  "",
	}
}

type OrderResponse struct {
	ID          uint      `json:"id"`
	OrderNumber string    `json:"order_number"`
	TotalAmount float64   `json:"total_amount"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

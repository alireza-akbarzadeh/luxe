package dto

import (
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

type CreateGiftCardRequest struct {
	Amount         float64 `json:"amount" validate:"required,gt=0"`
	RecipientEmail string  `json:"recipient_email" validate:"required,email"`
	RecipientName  string  `json:"recipient_name" validate:"required,min=1,max=255"`
	SenderName     string  `json:"sender_name" validate:"required,min=1,max=255"`
	Message        string  `json:"message" validate:"omitempty,max=500"`
	DeliveryDate   string  `json:"delivery_date" validate:"omitempty"`
}

type GiftCardResponse struct {
	ID              uint       `json:"id"`
	Code            string     `json:"code"`
	SenderUserID    uint       `json:"sender_user_id"`
	RecipientUserID *uint      `json:"recipient_user_id,omitempty"`
	RecipientEmail  string     `json:"recipient_email"`
	RecipientName   string     `json:"recipient_name,omitempty"`
	SenderName      string     `json:"sender_name,omitempty"`
	Message         string     `json:"message,omitempty"`
	InitialAmount   float64    `json:"initial_amount"`
	Balance         float64    `json:"balance"`
	Currency        string     `json:"currency"`
	Status          string     `json:"status"`
	DeliveryDate    *time.Time `json:"delivery_date,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	RedeemedAt      *time.Time `json:"redeemed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

func ToGiftCardResponse(card *models.GiftCard) GiftCardResponse {
	return GiftCardResponse{
		ID:              card.ID,
		Code:            card.Code,
		SenderUserID:    card.SenderUserID,
		RecipientUserID: card.RecipientUserID,
		RecipientEmail:  card.RecipientEmail,
		RecipientName:   card.RecipientName,
		SenderName:      card.SenderName,
		Message:         card.Message,
		InitialAmount:   card.InitialAmount,
		Balance:         card.Balance,
		Currency:        card.Currency,
		Status:          card.Status,
		DeliveryDate:    card.DeliveryDate,
		ExpiresAt:       card.ExpiresAt,
		RedeemedAt:      card.RedeemedAt,
		CreatedAt:       card.CreatedAt,
	}
}

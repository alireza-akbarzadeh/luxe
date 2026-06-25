package dto

import (
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

type CreateStoreReviewRequest struct {
	Rating  int    `json:"rating" validate:"required,min=1,max=5"`
	Comment string `json:"comment" validate:"required,min=3,max=2000"`
}

type UpdateStoreReviewRequest struct {
	Rating  *int    `json:"rating,omitempty" validate:"omitempty,min=1,max=5"`
	Comment *string `json:"comment,omitempty" validate:"omitempty,min=3,max=2000"`
}

type StoreReviewResponse struct {
	ID        uint      `json:"id"`
	StoreID   uint      `json:"store_id"`
	UserID    uint      `json:"user_id"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsOwner   bool      `json:"is_owner,omitempty"`
}

type StoreReviewSummary struct {
	Average float64        `json:"average"`
	Total   int64          `json:"total"`
	Counts  map[string]int `json:"counts"`
}

func ToStoreReviewResponse(review *models.StoreReview, viewerUserID uint) StoreReviewResponse {
	author := "Anonymous"
	if review.User != nil && review.User.ID != 0 {
		name := review.User.FirstName + " " + review.User.LastName
		if name != " " {
			author = name
		} else if review.User.Email != "" {
			author = review.User.Email
		}
	}

	return StoreReviewResponse{
		ID:        review.ID,
		StoreID:   review.StoreID,
		UserID:    review.UserID,
		Rating:    review.Rating,
		Comment:   review.Comment,
		Author:    author,
		CreatedAt: review.CreatedAt,
		UpdatedAt: review.UpdatedAt,
		IsOwner:   viewerUserID != 0 && review.UserID == viewerUserID,
	}
}

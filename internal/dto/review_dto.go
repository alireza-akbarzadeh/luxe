package dto

import (
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

type CreateReviewRequest struct {
	ProductID uint   `json:"product_id" validate:"required"`
	Rating    int    `json:"rating" validate:"required,min=1,max=5"`
	Comment   string `json:"comment,omitempty"`
	Title     string `json:"title,omitempty"`
}

type UpdateReviewRequest struct {
	Rating  *int    `json:"rating,omitempty" validate:"omitempty,min=1,max=5"`
	Title   *string `json:"title,omitempty"`
	Comment *string `json:"comment,omitempty"`
}

type ReviewResponse struct {
	ID         uint      `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	ProductID  uint      `json:"product_id"`
	UserID     uint      `json:"user_id"`
	Rating     int       `json:"rating"`
	Comment    string    `json:"comment,omitempty"`
	IsVerified bool      `json:"is_verified"`
	Title      string    `json:"title"`
	Status     string    `json:"status"`
	Author     string    `json:"author"`
	IsOwner    bool      `json:"is_owner,omitempty"`
}

type AdminReviewResponse struct {
	ReviewResponse
	ProductName string `json:"product_name,omitempty"`
}

type AdminReviewListFilters struct {
	Status    string `form:"status"`
	ProductID uint   `form:"product_id"`
	Limit     int    `form:"limit"`
	Offset    int    `form:"offset"`
}

type ModerateReviewRequest struct {
	Status string `json:"status" validate:"required,oneof=approved rejected"`
}

type ReviewSummary struct {
	Average float64        `json:"average"`
	Total   int64          `json:"total"`
	Counts  map[string]int `json:"counts"`
}

func ToReviewResponse(review *models.Review, viewerUserID uint) ReviewResponse {
	author := "Anonymous"
	if review.User.ID != 0 {
		name := review.User.FirstName + " " + review.User.LastName
		if name != " " {
			author = name
		}
	}

	return ReviewResponse{
		ID:         review.ID,
		CreatedAt:  review.CreatedAt,
		UpdatedAt:  review.UpdatedAt,
		ProductID:  review.ProductID,
		UserID:     review.UserID,
		Rating:     review.Rating,
		Comment:    review.Comment,
		IsVerified: review.IsVerified,
		Title:      review.Title,
		Status:     review.Status,
		Author:     author,
		IsOwner:    viewerUserID != 0 && review.UserID == viewerUserID,
	}
}

func ToAdminReviewResponse(review *models.Review) AdminReviewResponse {
	resp := AdminReviewResponse{
		ReviewResponse: ToReviewResponse(review, 0),
	}
	if review.Product.ID != 0 {
		resp.ProductName = review.Product.Name
	}
	return resp
}

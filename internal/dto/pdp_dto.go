package dto

import (
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

type ProductStoreSummary struct {
	ID            uint    `json:"id"`
	Name          string  `json:"name"`
	Slug          string  `json:"slug"`
	LogoURL       string  `json:"logo_url,omitempty"`
	Rating        float64 `json:"rating"`
	ReviewCount   int     `json:"review_count"`
	IsVerified    bool    `json:"is_verified"`
	ShippingInfo  string  `json:"shipping_info,omitempty"`
	ReturnPolicy  string  `json:"return_policy,omitempty"`
	Location      string  `json:"location,omitempty"`
	FollowerCount int     `json:"follower_count,omitempty"`
}

func ToProductStoreSummary(store *models.Store) *ProductStoreSummary {
	if store == nil || store.ID == 0 {
		return nil
	}
	return &ProductStoreSummary{
		ID:            store.ID,
		Name:          store.Name,
		Slug:          store.Slug,
		LogoURL:       store.LogoURL,
		Rating:        store.Rating,
		ReviewCount:   store.ReviewCount,
		IsVerified:    store.IsVerified,
		ShippingInfo:  store.ShippingInfo,
		ReturnPolicy:  store.ReturnPolicy,
		Location:      store.Location,
		FollowerCount: store.FollowerCount,
	}
}

type PriceHistoryPoint struct {
	RecordedAt     time.Time `json:"recorded_at"`
	Price          float64   `json:"price"`
	CompareAtPrice *float64  `json:"compare_at_price,omitempty"`
}

type ProductAlternativeResponse struct {
	ProductResponse
	StoreName   string  `json:"store_name"`
	StoreSlug   string  `json:"store_slug"`
	StoreLogo   string  `json:"store_logo,omitempty"`
	StoreRating float64 `json:"store_rating"`
}

type CreateProductQuestionRequest struct {
	Body string `json:"body" validate:"required,min=5,max=1000"`
}

type CreateProductAnswerRequest struct {
	Body string `json:"body" validate:"required,min=2,max=2000"`
}

type ProductAnswerResponse struct {
	ID           uint      `json:"id"`
	QuestionID   uint      `json:"question_id"`
	Author       string    `json:"author"`
	Body         string    `json:"body"`
	IsStoreReply bool      `json:"is_store_reply"`
	IsAIReply    bool      `json:"is_ai_reply"`
	CreatedAt    time.Time `json:"created_at"`
}

type ProductQuestionResponse struct {
	ID        uint                    `json:"id"`
	ProductID uint                    `json:"product_id"`
	Author    string                  `json:"author"`
	Body      string                  `json:"body"`
	CreatedAt time.Time               `json:"created_at"`
	IsOwner   bool                    `json:"is_owner,omitempty"`
	Answers   []ProductAnswerResponse `json:"answers"`
}

func ToProductAnswerResponse(answer *models.ProductAnswer) ProductAnswerResponse {
	author := "Anonymous"
	if answer.User.ID != 0 {
		name := answer.User.FirstName + " " + answer.User.LastName
		if name != " " {
			author = name
		}
		if answer.IsStoreReply {
			author = "Store"
		}
		if answer.IsAIReply {
			author = "Store assistant"
		}
	}
	return ProductAnswerResponse{
		ID:           answer.ID,
		QuestionID:   answer.QuestionID,
		Author:       author,
		Body:         answer.Body,
		IsStoreReply: answer.IsStoreReply,
		IsAIReply:    answer.IsAIReply,
		CreatedAt:    answer.CreatedAt,
	}
}

func ToProductQuestionResponse(question *models.ProductQuestion, viewerUserID uint) ProductQuestionResponse {
	author := "Anonymous"
	if question.User.ID != 0 {
		name := question.User.FirstName + " " + question.User.LastName
		if name != " " {
			author = name
		}
	}
	answers := make([]ProductAnswerResponse, len(question.Answers))
	for i := range question.Answers {
		answers[i] = ToProductAnswerResponse(&question.Answers[i])
	}
	return ProductQuestionResponse{
		ID:        question.ID,
		ProductID: question.ProductID,
		Author:    author,
		Body:      question.Body,
		CreatedAt: question.CreatedAt,
		IsOwner:   viewerUserID != 0 && question.UserID == viewerUserID,
		Answers:   answers,
	}
}

type StockNotificationStatusResponse struct {
	Subscribed bool `json:"subscribed"`
}

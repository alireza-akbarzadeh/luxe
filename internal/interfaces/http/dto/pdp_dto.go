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
	ID              uint      `json:"id"`
	QuestionID      uint      `json:"question_id"`
	Author          string    `json:"author"`
	Body            string    `json:"body"`
	IsStoreReply    bool      `json:"is_store_reply"`
	IsAIReply       bool      `json:"is_ai_reply"`
	IsVerifiedBuyer bool      `json:"is_verified_buyer"`
	CreatedAt       time.Time `json:"created_at"`
}

type ProductQuestionResponse struct {
	ID              uint                    `json:"id"`
	ProductID       uint                    `json:"product_id"`
	ProductName     string                  `json:"product_name,omitempty"`
	ProductSlug     string                  `json:"product_slug,omitempty"`
	Author          string                  `json:"author"`
	Body            string                  `json:"body"`
	CreatedAt       time.Time               `json:"created_at"`
	IsOwner         bool                    `json:"is_owner,omitempty"`
	IsVerifiedBuyer bool                    `json:"is_verified_buyer"`
	Answers         []ProductAnswerResponse `json:"answers"`
}

func ToProductAnswerResponse(answer *models.ProductAnswer, verifiedBuyers map[uint]bool) ProductAnswerResponse {
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
	isVerifiedBuyer := false
	if !answer.IsStoreReply && !answer.IsAIReply && verifiedBuyers != nil {
		isVerifiedBuyer = verifiedBuyers[answer.UserID]
	}
	return ProductAnswerResponse{
		ID:              answer.ID,
		QuestionID:      answer.QuestionID,
		Author:          author,
		Body:            answer.Body,
		IsStoreReply:    answer.IsStoreReply,
		IsAIReply:       answer.IsAIReply,
		IsVerifiedBuyer: isVerifiedBuyer,
		CreatedAt:       answer.CreatedAt,
	}
}

func ToProductQuestionResponse(
	question *models.ProductQuestion,
	viewerUserID uint,
	verifiedBuyers map[uint]bool,
) ProductQuestionResponse {
	author := "Anonymous"
	if question.User.ID != 0 {
		name := question.User.FirstName + " " + question.User.LastName
		if name != " " {
			author = name
		}
	}
	answers := make([]ProductAnswerResponse, len(question.Answers))
	for i := range question.Answers {
		answers[i] = ToProductAnswerResponse(&question.Answers[i], verifiedBuyers)
	}
	isVerifiedBuyer := verifiedBuyers != nil && verifiedBuyers[question.UserID]
	return ProductQuestionResponse{
		ID:              question.ID,
		ProductID:       question.ProductID,
		Author:          author,
		Body:            question.Body,
		CreatedAt:       question.CreatedAt,
		IsOwner:         viewerUserID != 0 && question.UserID == viewerUserID,
		IsVerifiedBuyer: isVerifiedBuyer,
		Answers:         answers,
	}
}

func ToUserProductQuestionResponse(
	question *models.ProductQuestion,
	viewerUserID uint,
	verifiedBuyers map[uint]bool,
) ProductQuestionResponse {
	resp := ToProductQuestionResponse(question, viewerUserID, verifiedBuyers)
	if question.Product.ID != 0 {
		resp.ProductName = question.Product.Name
		resp.ProductSlug = question.Product.Slug
	}
	return resp
}

type CreateProductDiscussionRequest struct {
	Title string `json:"title" validate:"required,min=5,max=200"`
	Body  string `json:"body" validate:"required,min=10,max=2000"`
}

type CreateProductDiscussionReplyRequest struct {
	Body string `json:"body" validate:"required,min=2,max=2000"`
}

type ProductDiscussionReplyResponse struct {
	ID           uint      `json:"id"`
	DiscussionID uint      `json:"discussion_id"`
	Author       string    `json:"author"`
	Body         string    `json:"body"`
	CreatedAt    time.Time `json:"created_at"`
	IsOwner      bool      `json:"is_owner,omitempty"`
}

type ProductDiscussionResponse struct {
	ID        uint                             `json:"id"`
	ProductID uint                             `json:"product_id"`
	Author    string                           `json:"author"`
	Title     string                           `json:"title"`
	Body      string                           `json:"body"`
	CreatedAt time.Time                        `json:"created_at"`
	IsOwner   bool                             `json:"is_owner,omitempty"`
	Replies   []ProductDiscussionReplyResponse `json:"replies"`
}

func ToProductDiscussionReplyResponse(reply *models.ProductDiscussionReply, viewerUserID uint) ProductDiscussionReplyResponse {
	author := "Anonymous"
	if reply.User.ID != 0 {
		name := reply.User.FirstName + " " + reply.User.LastName
		if name != " " {
			author = name
		}
	}
	return ProductDiscussionReplyResponse{
		ID:           reply.ID,
		DiscussionID: reply.DiscussionID,
		Author:       author,
		Body:         reply.Body,
		CreatedAt:    reply.CreatedAt,
		IsOwner:      viewerUserID != 0 && reply.UserID == viewerUserID,
	}
}

func ToProductDiscussionResponse(discussion *models.ProductDiscussion, viewerUserID uint) ProductDiscussionResponse {
	author := "Anonymous"
	if discussion.User.ID != 0 {
		name := discussion.User.FirstName + " " + discussion.User.LastName
		if name != " " {
			author = name
		}
	}
	replies := make([]ProductDiscussionReplyResponse, len(discussion.Replies))
	for i := range discussion.Replies {
		replies[i] = ToProductDiscussionReplyResponse(&discussion.Replies[i], viewerUserID)
	}
	return ProductDiscussionResponse{
		ID:        discussion.ID,
		ProductID: discussion.ProductID,
		Author:    author,
		Title:     discussion.Title,
		Body:      discussion.Body,
		CreatedAt: discussion.CreatedAt,
		IsOwner:   viewerUserID != 0 && discussion.UserID == viewerUserID,
		Replies:   replies,
	}
}

type StockNotificationStatusResponse struct {
	Subscribed bool `json:"subscribed"`
}

// StockHeatmapPoint is a single day of availability for PDP heatmap charts.
type StockHeatmapPoint struct {
	Date  time.Time `json:"date"`
	Stock int       `json:"stock"`
	Level string    `json:"level"`
}

// StockHeatmapData summarizes stock availability over a rolling window.
type StockHeatmapData struct {
	TrackInventory    bool                `json:"track_inventory"`
	IsDigital         bool                `json:"is_digital"`
	CurrentStock      int                 `json:"current_stock"`
	LowStockThreshold int                 `json:"low_stock_threshold"`
	Days              int                 `json:"days"`
	Points            []StockHeatmapPoint `json:"points"`
	InStockDays       int                 `json:"in_stock_days"`
	OutOfStockDays    int                 `json:"out_of_stock_days"`
	LowStockDays      int                 `json:"low_stock_days"`
}

// ProductTimelineMeta carries optional fields for timeline event rendering.
type ProductTimelineMeta struct {
	PriceFrom     *float64 `json:"price_from,omitempty"`
	PriceTo       *float64 `json:"price_to,omitempty"`
	StockQuantity *int     `json:"stock_quantity,omitempty"`
	ReviewRating  *int     `json:"review_rating,omitempty"`
	ReviewCount   *int     `json:"review_count,omitempty"`
	WorkflowState string   `json:"workflow_state,omitempty"`
	WorkflowEvent string   `json:"workflow_event,omitempty"`
}

// ProductTimelineEvent is a single lifecycle milestone on the PDP timeline.
type ProductTimelineEvent struct {
	Type       string              `json:"type"`
	OccurredAt time.Time           `json:"occurred_at"`
	Meta       ProductTimelineMeta `json:"meta,omitempty"`
}

// ProductTimelineData summarizes notable product lifecycle events.
type ProductTimelineData struct {
	Days   int                    `json:"days"`
	Events []ProductTimelineEvent `json:"events"`
}

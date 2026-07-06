package dto

// CreateReverseMarketplaceRequest is a buyer-posted wanted listing.
type CreateReverseMarketplaceRequest struct {
	Title       string   `json:"title" validate:"required,min=3,max=200"`
	Description string   `json:"description" validate:"omitempty,max=2000"`
	Category    string   `json:"category" validate:"omitempty,max=100"`
	BudgetMin   *float64 `json:"budget_min" validate:"omitempty,gte=0"`
	BudgetMax   *float64 `json:"budget_max" validate:"omitempty,gte=0"`
}

// CreateReverseMarketplaceOfferRequest is a vendor response to a buyer request.
type CreateReverseMarketplaceOfferRequest struct {
	Message      string  `json:"message" validate:"omitempty,max=2000"`
	OfferedPrice float64 `json:"offered_price" validate:"required,gte=0"`
}

// ReverseMarketplaceOfferResponse is a vendor offer on a request.
type ReverseMarketplaceOfferResponse struct {
	ID           uint    `json:"id"`
	StoreID      uint    `json:"store_id"`
	StoreName    string  `json:"store_name,omitempty"`
	Message      string  `json:"message"`
	OfferedPrice float64 `json:"offered_price"`
	Status       string  `json:"status"`
	CreatedAt    string  `json:"created_at"`
}

// ReverseMarketplaceRequestListItem is a summary card for discovery.
type ReverseMarketplaceRequestListItem struct {
	ID          uint     `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	BudgetMin   *float64 `json:"budget_min,omitempty"`
	BudgetMax   *float64 `json:"budget_max,omitempty"`
	Status      string   `json:"status"`
	OfferCount  int      `json:"offer_count"`
	CreatedAt   string   `json:"created_at"`
}

// ReverseMarketplaceRequestResponse is a full request with offers.
type ReverseMarketplaceRequestResponse struct {
	ID          uint                              `json:"id"`
	Title       string                            `json:"title"`
	Description string                            `json:"description"`
	Category    string                            `json:"category"`
	BudgetMin   *float64                          `json:"budget_min,omitempty"`
	BudgetMax   *float64                          `json:"budget_max,omitempty"`
	Status      string                            `json:"status"`
	CreatedAt   string                            `json:"created_at"`
	Offers      []ReverseMarketplaceOfferResponse `json:"offers,omitempty"`
}

// ReverseMarketplaceRequestListResponse wraps paginated requests.
type ReverseMarketplaceRequestListResponse struct {
	Requests []ReverseMarketplaceRequestListItem `json:"requests"`
	Total    int64                               `json:"total"`
	Limit    int                                 `json:"limit"`
	Offset   int                                 `json:"offset"`
}

package dto

// SmartBundleItem is a compatibility-ranked product bundle for storefront upsell.
type SmartBundleItem struct {
	ID                 string            `json:"id"`
	Title              string            `json:"title"`
	Description        string            `json:"description"`
	Intent             string            `json:"intent"`
	CompatibilityScore int               `json:"compatibility_score"`
	Subtotal           float64           `json:"subtotal"`
	Products           []ProductResponse `json:"products"`
}

// SmartBundlesResponse wraps bundle suggestions for a product or cart context.
type SmartBundlesResponse struct {
	Intent  string            `json:"intent"`
	Bundles []SmartBundleItem `json:"bundles"`
}

// SuggestSmartBundlesRequest generates bundles from one or more anchor product IDs.
type SuggestSmartBundlesRequest struct {
	ProductIDs []uint `json:"product_ids" validate:"required,min=1,max=8,dive,gt=0"`
	Intent     string `json:"intent" validate:"omitempty,max=32"`
	Limit      int    `json:"limit" validate:"omitempty,min=1,max=6"`
}

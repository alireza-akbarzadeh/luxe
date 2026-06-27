package dto

import "time"

// PlusBenefitFeature describes a single marketing benefit on the landing page.
type PlusBenefitFeature struct {
	Key         string `json:"key"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// PlusTierWindow compares return windows by tier.
type PlusTierWindow struct {
	Free int `json:"free"`
	Plus int `json:"plus"`
}

// PlusBenefitsResponse is the public catalog of Luxe Plus perks.
type PlusBenefitsResponse struct {
	PlanName         string             `json:"plan_name"`
	AnnualPrice      float64            `json:"annual_price"`
	Currency         string             `json:"currency"`
	DiscountPercent  int                `json:"discount_percent"`
	ReturnWindowDays PlusTierWindow     `json:"return_window_days"`
	PriorityShipping bool               `json:"priority_shipping"`
	PrioritySupport  bool               `json:"priority_support"`
	Features         []PlusBenefitFeature `json:"features"`
}

// PlusMemberBenefitsSummary reflects active perks for the signed-in user.
type PlusMemberBenefitsSummary struct {
	DiscountPercent  int  `json:"discount_percent"`
	ReturnWindowDays int  `json:"return_window_days"`
	PriorityShipping bool `json:"priority_shipping"`
	PrioritySupport  bool `json:"priority_support"`
}

// MembershipStatusResponse is the authenticated user's Plus state.
type MembershipStatusResponse struct {
	Tier             string                    `json:"tier"`
	IsPlusActive     bool                      `json:"is_plus_active"`
	PlusSubscribedAt *time.Time                `json:"plus_subscribed_at,omitempty"`
	PlusExpiresAt    *time.Time                `json:"plus_expires_at,omitempty"`
	AnnualPrice      float64                   `json:"annual_price"`
	Currency         string                    `json:"currency"`
	Benefits         PlusMemberBenefitsSummary `json:"benefits"`
}

// SubscribePlusRequest selects how the user pays for Luxe Plus.
type SubscribePlusRequest struct {
	PaymentMethod string `json:"payment_method" validate:"required,oneof=wallet gift_card stripe"`
	GiftCardCode  string `json:"gift_card_code,omitempty"`
}

// SubscribePlusResponse is returned after initiating or completing a Plus subscription.
type SubscribePlusResponse struct {
	Membership      MembershipStatusResponse `json:"membership"`
	PaymentMethod   string                   `json:"payment_method"`
	PaymentStatus   string                   `json:"payment_status"`
	CheckoutURL     string                   `json:"checkout_url,omitempty"`
	StripeSessionID string                   `json:"stripe_session_id,omitempty"`
}

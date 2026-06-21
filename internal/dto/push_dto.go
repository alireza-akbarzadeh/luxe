package dto

// PushSubscriptionKeys holds encryption keys from the browser Push API.
type PushSubscriptionKeys struct {
	P256dh string `json:"p256dh" binding:"required"`
	Auth   string `json:"auth" binding:"required"`
}

// RegisterPushSubscriptionRequest registers a Web Push subscription for the authenticated user.
type RegisterPushSubscriptionRequest struct {
	Endpoint string               `json:"endpoint" binding:"required"`
	Keys     PushSubscriptionKeys `json:"keys" binding:"required"`
}

// DeletePushSubscriptionRequest removes a subscription by endpoint.
type DeletePushSubscriptionRequest struct {
	Endpoint string `json:"endpoint" binding:"required"`
}

// VapidPublicKeyResponse exposes the VAPID public key for client-side subscribe().
type VapidPublicKeyResponse struct {
	PublicKey string `json:"public_key"`
	Enabled   bool   `json:"enabled"`
}

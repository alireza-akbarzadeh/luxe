package shipment

// CreateShipmentRequest holds shipment creation input for checkout and admin flows.
type CreateShipmentRequest struct {
	OrderID        uint    `json:"order_id" validate:"required,gt=0"`
	Carrier        string  `json:"carrier" validate:"required"`
	TrackingNumber string  `json:"tracking_number,omitempty"`
	AddressLine1   string  `json:"address_line1" validate:"required"`
	AddressLine2   string  `json:"address_line2,omitempty"`
	City           string  `json:"city" validate:"required"`
	State          string  `json:"state,omitempty"`
	PostalCode     string  `json:"postal_code" validate:"required"`
	Country        string  `json:"country" validate:"required"`
	ProviderID     *uint   `json:"provider_id,omitempty"`
	ShippingPrice  float64 `json:"shipping_price" validate:"gte=0"`
}

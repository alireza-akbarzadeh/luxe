package dto

// PerformShipmentTransitionRequest triggers a workflow event on a shipment (admin).
type PerformShipmentTransitionRequest struct {
	Event string `json:"event" validate:"required,min=1,max=64"`
	Note  string `json:"note"  validate:"omitempty,max=512"`
}

// ShipmentTransitionResponse is returned after a successful shipment workflow transition.
type ShipmentTransitionResponse struct {
	Transition TransitionResultView `json:"transition"`
	Shipment   interface{}          `json:"shipment"` // models.Shipment in responses
}

// AdminShipmentListFilters supports admin listing of shipments.
type AdminShipmentListFilters struct {
	Status  string `form:"status"`
	Carrier string `form:"carrier"`
	OrderID *uint  `form:"order_id"`
	Search  string `form:"search"`
	Limit   int    `form:"limit"`
	Offset  int    `form:"offset"`
}

// AdminShipmentListItem is a row in the admin shipments table.
type AdminShipmentListItem struct {
	ID                uint       `json:"id"`
	OrderID           uint       `json:"order_id"`
	OrderNumber       string     `json:"order_number,omitempty"`
	Carrier           string     `json:"carrier"`
	TrackingNumber    string     `json:"tracking_number,omitempty"`
	Status            string     `json:"status"`
	CustomerName      string     `json:"customer_name,omitempty"`
	City              string     `json:"city,omitempty"`
	Country           string     `json:"country,omitempty"`
	EstimatedDelivery *string    `json:"estimated_delivery,omitempty"`
	ShippedAt         *string    `json:"shipped_at,omitempty"`
	CreatedAt         string     `json:"created_at"`
	State             *StateView `json:"state,omitempty"`
}

// AdminShipmentListData wraps paginated admin shipment rows.
type AdminShipmentListData struct {
	Shipments []AdminShipmentListItem `json:"shipments"`
	Total     int64                   `json:"total"`
	Limit     int                     `json:"limit"`
	Offset    int                     `json:"offset"`
}

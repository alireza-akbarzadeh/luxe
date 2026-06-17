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

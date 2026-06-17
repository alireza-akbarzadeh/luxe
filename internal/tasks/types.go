package tasks

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/dto"
)

// Task type names (Asynq + memory queue).
const (
	TypeProcessOrder    = "order:process"
	TypeProcessShipment = "shipment:process"
)

// ProcessOrderPayload is the durable payload for mock payment fulfillment.
type ProcessOrderPayload struct {
	OrderID  uint          `json:"order_id"`
	CardInfo dto.CardInfo  `json:"card_info"`
}

// ProcessShipmentPayload is the durable payload for shipment background processing.
type ProcessShipmentPayload struct {
	ShipmentID uint `json:"shipment_id"`
}

// Handlers registers business logic invoked by background workers.
type Handlers struct {
	ProcessOrder    func(orderID uint, cardInfo dto.CardInfo) error
	ProcessShipment func(shipmentID uint) error
}

// BindHandlers updates handlers on a running queue (used after services are wired in main).
func BindHandlers(q JobQueue, h Handlers) {
	switch v := q.(type) {
	case *memoryQueue:
		v.handlers = h
	case *asynqQueue:
		v.handlers = h
	}
}

// JobQueue enqueues durable (or in-memory) background work.
type JobQueue interface {
	EnqueueProcessOrder(ctx context.Context, orderID uint, cardInfo dto.CardInfo) error
	EnqueueProcessShipment(ctx context.Context, shipmentID uint) error
	Start() error
	Shutdown()
	Backend() string
}

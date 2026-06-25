package asynq

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
)

// Task type names (Asynq + memory queue).
const (
	TypeProcessOrder    = "order:process"
	TypeProcessShipment = "shipment:process"
	TypeSendEmail       = "email:send"
)

// ProcessOrderPayload is the durable payload for mock payment fulfillment.
type ProcessOrderPayload struct {
	OrderID  uint         `json:"order_id"`
	CardInfo dto.CardInfo `json:"card_info"`
}

// ProcessShipmentPayload is the durable payload for shipment background processing.
type ProcessShipmentPayload struct {
	ShipmentID uint `json:"shipment_id"`
}

// SendEmailPayload carries the minimal data needed to send a transactional email.
type SendEmailPayload struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// Handlers registers business logic invoked by background workers.
type Handlers struct {
	ProcessOrder    func(ctx context.Context, orderID uint, cardInfo dto.CardInfo) error
	ProcessShipment func(ctx context.Context, shipmentID uint) error
	SendEmail       func(ctx context.Context, to, subject, body string) error
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
	EnqueueSendEmail(ctx context.Context, to, subject, body string) error
	Start() error
	Shutdown()
	Backend() string
}

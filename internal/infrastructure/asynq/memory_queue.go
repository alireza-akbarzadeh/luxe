package asynq

import (
	"context"
	"fmt"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
)

type memoryQueue struct {
	pool     *WorkerPool
	handlers Handlers
}

func newMemoryQueue(handlers Handlers) *memoryQueue {
	return &memoryQueue{
		pool:     NewWorkerPool(5, 100),
		handlers: handlers,
	}
}

func (q *memoryQueue) Backend() string {
	return "memory"
}

func (q *memoryQueue) Start() error {
	q.pool.Start()
	return nil
}

func (q *memoryQueue) Shutdown() {
	q.pool.Stop()
}

func (q *memoryQueue) EnqueueProcessOrder(_ context.Context, orderID uint, cardInfo dto.CardInfo) error {
	if q.handlers.ProcessOrder == nil {
		return fmt.Errorf("process order handler not configured")
	}
	card := cardInfo
	q.pool.Enqueue(Job{
		ID:      fmt.Sprintf("fulfill_%d", orderID),
		Payload: ProcessOrderPayload{OrderID: orderID, CardInfo: card},
		Handler: func(payload interface{}) error {
			p, ok := payload.(ProcessOrderPayload)
			if !ok {
				return fmt.Errorf("invalid process order payload type")
			}
			return q.handlers.ProcessOrder(context.Background(), p.OrderID, p.CardInfo)
		},
	})
	return nil
}

func (q *memoryQueue) EnqueueProcessShipment(_ context.Context, shipmentID uint) error {
	if q.handlers.ProcessShipment == nil {
		return fmt.Errorf("process shipment handler not configured")
	}
	q.pool.Enqueue(Job{
		ID:      fmt.Sprintf("shipment_%d", shipmentID),
		Payload: shipmentID,
		Handler: func(payload interface{}) error {
			id, ok := payload.(uint)
			if !ok {
				return fmt.Errorf("invalid shipment payload type")
			}
			return q.handlers.ProcessShipment(context.Background(), id)
		},
	})
	return nil
}

func (q *memoryQueue) EnqueueSendEmail(_ context.Context, to, subject, body string) error {
	if q.handlers.SendEmail == nil {
		return fmt.Errorf("send email handler not configured")
	}
	p := SendEmailPayload{To: to, Subject: subject, Body: body}
	q.pool.Enqueue(Job{
		ID:      fmt.Sprintf("email_%s_%d", to, time.Now().UnixNano()),
		Payload: p,
		Handler: func(payload interface{}) error {
			ep, ok := payload.(SendEmailPayload)
			if !ok {
				return fmt.Errorf("invalid email payload type")
			}
			return q.handlers.SendEmail(context.Background(), ep.To, ep.Subject, ep.Body)
		},
	})
	return nil
}

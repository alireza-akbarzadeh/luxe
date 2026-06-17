package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

type asynqQueue struct {
	client   *asynq.Client
	server   *asynq.Server
	mux      *asynq.ServeMux
	handlers Handlers
}

func newAsynqQueue(redisURL string, handlers Handlers) (*asynqQueue, error) {
	opt, err := parseRedisOpt(redisURL)
	if err != nil {
		return nil, err
	}

	server := asynq.NewServer(opt, asynq.Config{
		Concurrency: 5,
		Queues: map[string]int{
			"critical": 6,
			"default":  3,
		},
		RetryDelayFunc: func(n int, _ error, _ *asynq.Task) time.Duration {
			return time.Duration(n) * time.Second
		},
	})

	q := &asynqQueue{
		client:   asynq.NewClient(opt),
		server:   server,
		mux:      asynq.NewServeMux(),
		handlers: handlers,
	}

	q.registerHandlers()
	return q, nil
}

func parseRedisOpt(redisURL string) (asynq.RedisClientOpt, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return asynq.RedisClientOpt{}, fmt.Errorf("parse REDIS_URL: %w", err)
	}
	return asynq.RedisClientOpt{
		Addr:      opts.Addr,
		Password:  opts.Password,
		DB:        opts.DB,
		TLSConfig: opts.TLSConfig,
	}, nil
}

func (q *asynqQueue) Backend() string {
	return "asynq"
}

func (q *asynqQueue) registerHandlers() {
	q.mux.HandleFunc(TypeProcessOrder, q.handleProcessOrder)
	q.mux.HandleFunc(TypeProcessShipment, q.handleProcessShipment)
}

func (q *asynqQueue) handleProcessOrder(_ context.Context, t *asynq.Task) error {
	if q.handlers.ProcessOrder == nil {
		return fmt.Errorf("process order handler not configured")
	}
	var payload ProcessOrderPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("decode process order payload: %w", err)
	}
		if err := q.handlers.ProcessOrder(payload.OrderID, payload.CardInfo); err != nil {
		utils.Log.WithError(err).WithField("order_id", payload.OrderID).Error("asynq: process order failed")
		return err
	}
	return nil
}

func (q *asynqQueue) handleProcessShipment(_ context.Context, t *asynq.Task) error {
	if q.handlers.ProcessShipment == nil {
		return fmt.Errorf("process shipment handler not configured")
	}
	var payload ProcessShipmentPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("decode process shipment payload: %w", err)
	}
		if err := q.handlers.ProcessShipment(payload.ShipmentID); err != nil {
		utils.Log.WithError(err).WithField("shipment_id", payload.ShipmentID).Error("asynq: process shipment failed")
		return err
	}
	return nil
}

func (q *asynqQueue) Start() error {
	go func() {
		if err := q.server.Run(q.mux); err != nil {
			utils.Log.WithError(err).Error("asynq server stopped")
		}
	}()
	return nil
}

func (q *asynqQueue) Shutdown() {
	q.server.Shutdown()
	if err := q.client.Close(); err != nil {
		utils.Log.WithError(err).Warn("failed to close asynq client")
	}
}

func (q *asynqQueue) EnqueueProcessOrder(ctx context.Context, orderID uint, cardInfo dto.CardInfo) error {
	payload, err := marshalPayload(ProcessOrderPayload{OrderID: orderID, CardInfo: cardInfo})
	if err != nil {
		return err
	}
	task := asynq.NewTask(TypeProcessOrder, payload)
	_, err = q.client.EnqueueContext(ctx, task,
		asynq.Queue("critical"),
		asynq.TaskID(fmt.Sprintf("fulfill_%d", orderID)),
		asynq.MaxRetry(5),
		asynq.Timeout(5*time.Minute),
	)
	return err
}

func (q *asynqQueue) EnqueueProcessShipment(ctx context.Context, shipmentID uint) error {
	payload, err := marshalPayload(ProcessShipmentPayload{ShipmentID: shipmentID})
	if err != nil {
		return err
	}
	task := asynq.NewTask(TypeProcessShipment, payload)
	_, err = q.client.EnqueueContext(ctx, task,
		asynq.Queue("default"),
		asynq.TaskID(fmt.Sprintf("shipment_%d", shipmentID)),
		asynq.MaxRetry(3),
		asynq.Timeout(2*time.Minute),
	)
	return err
}

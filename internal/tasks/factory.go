package tasks

import (
	"encoding/json"
	"fmt"

	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
)

// NewJobQueue creates an Asynq-backed queue when REDIS_URL is set, otherwise in-memory.
// Handlers can be bound later via BindHandlers after services are constructed.
func NewJobQueue(cfg *config.Config, handlers Handlers) (JobQueue, error) {
	if cfg != nil && cfg.Redis.Enabled {
		queue, err := newAsynqQueue(cfg.Redis.URL, handlers)
		if err != nil {
			return nil, err
		}
		utils.Log.WithField("backend", "asynq").Info("background job queue initialized")
		return queue, nil
	}

	queue := newMemoryQueue(handlers)
	utils.Log.WithField("backend", "memory").Info("background job queue initialized (set REDIS_URL for durable Asynq)")
	return queue, nil
}

func marshalPayload(v interface{}) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal task payload: %w", err)
	}
	return data, nil
}

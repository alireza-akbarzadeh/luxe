package integration

import (
	"testing"

	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/asynq"
	"github.com/stretchr/testify/require"
)

func TestJobQueue_UsesAsynqWhenRedisConfigured(t *testing.T) {
	if testCfg == nil || !testCfg.Redis.Enabled {
		t.Skip("REDIS_URL not set; skipping Asynq integration test")
	}

	queue, err := asynq.NewJobQueue(testCfg, asynq.Handlers{})
	require.NoError(t, err)
	require.Equal(t, "asynq", queue.Backend())

	require.NoError(t, queue.Start())
	queue.Shutdown()
}

func TestJobQueue_UsesMemoryWhenRedisUnset(t *testing.T) {
	if testCfg == nil || testCfg.Redis.Enabled {
		t.Skip("REDIS_URL is set; skipping in-memory queue test")
	}

	queue, err := asynq.NewJobQueue(testCfg, asynq.Handlers{})
	require.NoError(t, err)
	require.Equal(t, "memory", queue.Backend())
}

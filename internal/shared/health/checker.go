package health

import (
	"context"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/database"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Status is the aggregate health of dependency checks.
type Status struct {
	OK     bool            `json:"ok"`
	Checks map[string]bool `json:"checks"`
}

// Checker runs liveness and readiness probes against configured dependencies.
type Checker struct {
	db           *gorm.DB
	redisURL     string
	redisEnabled bool
}

func NewChecker(db *gorm.DB, cfg *config.Config) *Checker {
	if cfg == nil {
		return &Checker{db: db}
	}
	return &Checker{
		db:           db,
		redisURL:     cfg.Redis.URL,
		redisEnabled: cfg.Redis.Enabled,
	}
}

// Live indicates the process is running (no dependency checks).
func (c *Checker) Live() Status {
	return Status{
		OK:     true,
		Checks: map[string]bool{"process": true},
	}
}

// Ready checks dependencies required to serve traffic.
func (c *Checker) Ready(ctx context.Context) Status {
	checks := map[string]bool{}

	dbOK := c.db != nil && database.Ping(c.db) == nil
	checks["database"] = dbOK

	redisOK := true
	if c.redisEnabled {
		redisOK = pingRedis(ctx, c.redisURL) == nil
		checks["redis"] = redisOK
	}

	ok := dbOK && redisOK
	return Status{OK: ok, Checks: checks}
}

func pingRedis(ctx context.Context, url string) error {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return err
	}
	client := redis.NewClient(opts)
	defer client.Close()

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return client.Ping(pingCtx).Err()
}

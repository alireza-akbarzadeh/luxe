package audit

import (
	"context"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

// Commands orchestrates audit write use cases.
type Commands struct {
	repo *postgres.AuditRepository
}

// NewCommands creates audit command use cases.
func NewCommands(repo *postgres.AuditRepository) *Commands {
	return &Commands{repo: repo}
}

// Log persists an audit log entry.
func (c *Commands) Log(ctx context.Context, entry *models.AuditLog) error {
	if err := c.repo.Create(ctx, entry); err != nil {
		utils.Log.WithError(err).Warn("failed to write audit log")
		return err
	}
	return nil
}

// Queries orchestrates audit read use cases.
type Queries struct {
	repo *postgres.AuditRepository
}

// NewQueries creates audit query use cases.
func NewQueries(repo *postgres.AuditRepository) *Queries {
	return &Queries{repo: repo}
}

// List returns paginated audit logs with filters.
func (q *Queries) List(ctx context.Context, filters dto.AuditLogListFilters) ([]models.AuditLog, int64, error) {
	total, err := q.repo.CountFiltered(ctx, filters)
	if err != nil {
		return nil, 0, err
	}

	limit := filters.Limit
	if limit == 0 {
		limit = 20
	}

	logs, err := q.repo.ListFiltered(ctx, filters, limit)
	return logs, total, err
}

// Summary returns aggregate audit log statistics.
func (q *Queries) Summary(ctx context.Context) (*dto.AuditLogSummaryResponse, error) {
	total, err := q.repo.CountAll(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	last24h := now.Add(-24 * time.Hour)

	today, err := q.repo.CountSince(ctx, startOfDay)
	if err != nil {
		return nil, err
	}

	last24Hours, err := q.repo.CountSince(ctx, last24h)
	if err != nil {
		return nil, err
	}

	uniqueActors, err := q.repo.CountDistinctUsers(ctx)
	if err != nil {
		return nil, err
	}

	return &dto.AuditLogSummaryResponse{
		Total:        total,
		Last24Hours:  last24Hours,
		Today:        today,
		UniqueActors: uniqueActors,
	}, nil
}

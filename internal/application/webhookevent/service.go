package webhookevent

import (
	"context"
	"strings"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

// EventFilters mirrors services.WebhookEventFilters for listing.
type EventFilters struct {
	Source    string
	EventType string
	Status    string
	Limit     int
	Offset    int
}

// Commands orchestrates webhook event write use cases.
type Commands struct {
	repo *postgres.WebhookEventRepository
}

// NewCommands creates webhook event command use cases.
func NewCommands(repo *postgres.WebhookEventRepository) *Commands {
	return &Commands{repo: repo}
}

// Record saves the raw webhook payload and returns the created record.
func (c *Commands) Record(ctx context.Context, eventID, eventType, source string, payload []byte) (*models.WebhookEvent, error) {
	ev := models.WebhookEvent{
		EventID:   eventID,
		EventType: eventType,
		Source:    source,
		Status:    constants.WebhookStatusReceived,
		Payload:   payload,
	}

	if err := c.repo.Create(ctx, &ev); err != nil {
		if isUniqueViolation(err) {
			return nil, utils.ErrConflict("webhook event " + eventID + " already received")
		}
		return nil, utils.ErrInternal(err)
	}
	return &ev, nil
}

// MarkProcessed sets status=processed and processed_at=now.
func (c *Commands) MarkProcessed(ctx context.Context, eventID string) error {
	if err := c.repo.MarkProcessed(ctx, eventID, time.Now()); err != nil {
		return utils.ErrInternal(err)
	}
	return nil
}

// MarkFailed sets status=failed and stores the error message.
func (c *Commands) MarkFailed(ctx context.Context, eventID, errMsg string) error {
	if err := c.repo.MarkFailed(ctx, eventID, errMsg, time.Now()); err != nil {
		return utils.ErrInternal(err)
	}
	return nil
}

// Queries orchestrates webhook event read use cases.
type Queries struct {
	repo *postgres.WebhookEventRepository
}

// NewQueries creates webhook event query use cases.
func NewQueries(repo *postgres.WebhookEventRepository) *Queries {
	return &Queries{repo: repo}
}

// List returns paginated webhook events with optional filters.
func (q *Queries) List(ctx context.Context, filters EventFilters) ([]models.WebhookEvent, int64, error) {
	total, err := q.repo.CountFiltered(ctx, filters.Source, filters.EventType, filters.Status)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}

	events, err := q.repo.ListFiltered(ctx, filters.Source, filters.EventType, filters.Status, filters.Limit, filters.Offset)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	return events, total, nil
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "unique constraint") ||
		strings.Contains(msg, "duplicate key") ||
		strings.Contains(msg, "UNIQUE constraint failed")
}

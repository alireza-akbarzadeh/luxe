package push

import (
	"context"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

// Commands orchestrates push subscription write use cases.
type Commands struct {
	repo *postgres.PushRepository
}

// NewCommands creates push command use cases.
func NewCommands(repo *postgres.PushRepository) *Commands {
	return &Commands{repo: repo}
}

// RegisterSubscription upserts a web push subscription.
func (c *Commands) RegisterSubscription(
	ctx context.Context,
	userID uint,
	req dto.RegisterPushSubscriptionRequest,
	userAgent string,
) error {
	sub := models.PushSubscription{
		UserID:    userID,
		Endpoint:  strings.TrimSpace(req.Endpoint),
		P256dh:    strings.TrimSpace(req.Keys.P256dh),
		Auth:      strings.TrimSpace(req.Keys.Auth),
		UserAgent: userAgent,
	}
	if err := c.repo.UpsertSubscription(ctx, &sub); err != nil {
		return utils.ErrInternal(err)
	}
	return nil
}

// DeleteSubscription removes a subscription for a user.
func (c *Commands) DeleteSubscription(ctx context.Context, userID uint, endpoint string) error {
	rows, err := c.repo.DeleteSubscription(ctx, userID, endpoint)
	if err != nil {
		return utils.ErrInternal(err)
	}
	if rows == 0 {
		return utils.ErrNotFound("push subscription not found")
	}
	return nil
}

// DeleteByID removes a subscription by primary key.
func (c *Commands) DeleteByID(ctx context.Context, id uint) error {
	return c.repo.DeleteByID(ctx, id)
}

// Queries orchestrates push subscription read use cases.
type Queries struct {
	repo *postgres.PushRepository
}

// NewQueries creates push query use cases.
func NewQueries(repo *postgres.PushRepository) *Queries {
	return &Queries{repo: repo}
}

// ListByUser returns push subscriptions for a user.
func (q *Queries) ListByUser(ctx context.Context, userID uint) ([]models.PushSubscription, error) {
	subs, err := q.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	return subs, nil
}

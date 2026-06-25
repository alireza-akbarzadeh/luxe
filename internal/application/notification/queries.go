package notification

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

// Queries orchestrates notification read use cases.
type Queries struct {
	repo *postgres.NotificationRepository
}

// NewQueries creates notification query use cases.
func NewQueries(repo *postgres.NotificationRepository) *Queries {
	return &Queries{repo: repo}
}

// ListForUser returns paginated notifications for a user.
func (q *Queries) ListForUser(ctx context.Context, userID uint, limit, offset int) ([]models.Notification, int64, error) {
	total, err := q.repo.CountForUser(ctx, userID)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	notifications, err := q.repo.ListForUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	return notifications, total, nil
}

// ListChatMessages returns chat messages for a room.
func (q *Queries) ListChatMessages(ctx context.Context, roomID string, limit, offset int) ([]models.Message, error) {
	messages, err := q.repo.ListMessages(ctx, roomID, limit, offset)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	return messages, nil
}

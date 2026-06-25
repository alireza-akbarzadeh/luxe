package services

import (
	"context"

	appwebhook "github.com/alireza-akbarzadeh/luxe/internal/application/webhookevent"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

type WebhookEventServiceInterface interface {
	Record(ctx context.Context, eventID, eventType, source string, payload []byte) (*models.WebhookEvent, error)
	MarkProcessed(ctx context.Context, eventID string) error
	MarkFailed(ctx context.Context, eventID, errMsg string) error
	List(ctx context.Context, filters WebhookEventFilters) ([]models.WebhookEvent, int64, error)
}

type WebhookEventFilters struct {
	Source    string
	EventType string
	Status    string
	Limit     int
	Offset    int
}

type webhookEventService struct {
	commands *appwebhook.Commands
	queries  *appwebhook.Queries
}

func NewWebhookEventService(db *gorm.DB) WebhookEventServiceInterface {
	repo := postgres.NewWebhookEventRepository(db)
	return &webhookEventService{
		commands: appwebhook.NewCommands(repo),
		queries:  appwebhook.NewQueries(repo),
	}
}

func (s *webhookEventService) Record(ctx context.Context, eventID, eventType, source string, payload []byte) (*models.WebhookEvent, error) {
	return s.commands.Record(ctx, eventID, eventType, source, payload)
}

func (s *webhookEventService) MarkProcessed(ctx context.Context, eventID string) error {
	return s.commands.MarkProcessed(ctx, eventID)
}

func (s *webhookEventService) MarkFailed(ctx context.Context, eventID, errMsg string) error {
	return s.commands.MarkFailed(ctx, eventID, errMsg)
}

func (s *webhookEventService) List(ctx context.Context, filters WebhookEventFilters) ([]models.WebhookEvent, int64, error) {
	return s.queries.List(ctx, appwebhook.EventFilters{
		Source:    filters.Source,
		EventType: filters.EventType,
		Status:    filters.Status,
		Limit:     filters.Limit,
		Offset:    filters.Offset,
	})
}

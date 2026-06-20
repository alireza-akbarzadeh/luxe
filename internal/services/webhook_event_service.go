package services

import (
	"context"
	"strings"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"gorm.io/gorm"
)

type WebhookEventServiceInterface interface {
	// Record saves the raw webhook payload and returns the created record.
	// Returns (nil, ErrConflict) when the event_id was already seen (idempotency).
	Record(ctx context.Context, eventID, eventType, source string, payload []byte) (*models.WebhookEvent, error)

	// MarkProcessed sets status=processed and processed_at=now.
	MarkProcessed(ctx context.Context, eventID string) error

	// MarkFailed sets status=failed and stores the error message.
	MarkFailed(ctx context.Context, eventID, errMsg string) error

	// List returns paginated webhook events with optional filters.
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
	db *gorm.DB
}

func NewWebhookEventService(db *gorm.DB) WebhookEventServiceInterface {
	return &webhookEventService{db: db}
}

func (s *webhookEventService) Record(ctx context.Context, eventID, eventType, source string, payload []byte) (*models.WebhookEvent, error) {
	ev := models.WebhookEvent{
		EventID:   eventID,
		EventType: eventType,
		Source:    source,
		Status:    constants.WebhookStatusReceived,
		Payload:   payload,
	}

	result := s.db.WithContext(ctx).Create(&ev)
	if result.Error != nil {
		// Unique constraint violation → already seen.
		if isUniqueViolation(result.Error) {
			return nil, utils.ErrConflict("webhook event " + eventID + " already received")
		}
		return nil, utils.ErrInternal(result.Error)
	}
	return &ev, nil
}

func (s *webhookEventService) MarkProcessed(ctx context.Context, eventID string) error {
	now := time.Now()
	res := s.db.WithContext(ctx).Model(&models.WebhookEvent{}).
		Where("event_id = ?", eventID).
		Updates(map[string]interface{}{
			"status":       constants.WebhookStatusProcessed,
			"processed_at": now,
		})
	if res.Error != nil {
		return utils.ErrInternal(res.Error)
	}
	return nil
}

func (s *webhookEventService) MarkFailed(ctx context.Context, eventID, errMsg string) error {
	now := time.Now()
	res := s.db.WithContext(ctx).Model(&models.WebhookEvent{}).
		Where("event_id = ?", eventID).
		Updates(map[string]interface{}{
			"status":       constants.WebhookStatusFailed,
			"error_msg":    errMsg,
			"processed_at": now,
		})
	if res.Error != nil {
		return utils.ErrInternal(res.Error)
	}
	return nil
}

func (s *webhookEventService) List(ctx context.Context, filters WebhookEventFilters) ([]models.WebhookEvent, int64, error) {
	db := s.db.WithContext(ctx).Model(&models.WebhookEvent{})
	if filters.Source != "" {
		db = db.Where("source = ?", filters.Source)
	}
	if filters.EventType != "" {
		db = db.Where("event_type = ?", filters.EventType)
	}
	if filters.Status != "" {
		db = db.Where("status = ?", filters.Status)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, utils.ErrInternal(err)
	}

	var events []models.WebhookEvent
	if err := db.Order("created_at DESC").Limit(filters.Limit).Offset(filters.Offset).
		Find(&events).Error; err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	return events, total, nil
}

// isUniqueViolation returns true for PostgreSQL unique-constraint errors.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "unique constraint") ||
		strings.Contains(msg, "duplicate key") ||
		strings.Contains(msg, "UNIQUE constraint failed")
}

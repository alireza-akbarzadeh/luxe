package postgres

import (
	"context"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// WebhookEventRepository persists webhook events with GORM.
type WebhookEventRepository struct {
	db *gorm.DB
}

// NewWebhookEventRepository creates a GORM-backed webhook event repository.
func NewWebhookEventRepository(db *gorm.DB) *WebhookEventRepository {
	return &WebhookEventRepository{db: db}
}

// Create inserts a webhook event row.
func (r *WebhookEventRepository) Create(ctx context.Context, ev *models.WebhookEvent) error {
	return r.db.WithContext(ctx).Create(ev).Error
}

// MarkProcessed sets status to processed with timestamp.
func (r *WebhookEventRepository) MarkProcessed(ctx context.Context, eventID string, processedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&models.WebhookEvent{}).
		Where("event_id = ?", eventID).
		Updates(map[string]interface{}{
			"status":       constants.WebhookStatusProcessed,
			"processed_at": processedAt,
		}).Error
}

// MarkFailed sets status to failed with error message and timestamp.
func (r *WebhookEventRepository) MarkFailed(ctx context.Context, eventID, errMsg string, processedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&models.WebhookEvent{}).
		Where("event_id = ?", eventID).
		Updates(map[string]interface{}{
			"status":       constants.WebhookStatusFailed,
			"error_msg":    errMsg,
			"processed_at": processedAt,
		}).Error
}

// CountFiltered counts webhook events matching filters.
func (r *WebhookEventRepository) CountFiltered(ctx context.Context, source, eventType, status string) (int64, error) {
	var total int64
	err := r.applyFilters(r.db.WithContext(ctx).Model(&models.WebhookEvent{}), source, eventType, status).
		Count(&total).Error
	return total, err
}

// ListFiltered returns paginated webhook events matching filters.
func (r *WebhookEventRepository) ListFiltered(ctx context.Context, source, eventType, status string, limit, offset int) ([]models.WebhookEvent, error) {
	var events []models.WebhookEvent
	err := r.applyFilters(r.db.WithContext(ctx).Model(&models.WebhookEvent{}), source, eventType, status).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&events).Error
	return events, err
}

func (r *WebhookEventRepository) applyFilters(db *gorm.DB, source, eventType, status string) *gorm.DB {
	if source != "" {
		db = db.Where("source = ?", source)
	}
	if eventType != "" {
		db = db.Where("event_type = ?", eventType)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	return db
}

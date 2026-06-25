package postgres

import (
	"context"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PushRepository persists web push subscriptions with GORM.
type PushRepository struct {
	db *gorm.DB
}

// NewPushRepository creates a GORM-backed push subscription repository.
func NewPushRepository(db *gorm.DB) *PushRepository {
	return &PushRepository{db: db}
}

// UpsertSubscription inserts or updates a push subscription by endpoint.
func (r *PushRepository) UpsertSubscription(ctx context.Context, sub *models.PushSubscription) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "endpoint"}},
			DoUpdates: clause.AssignmentColumns([]string{"user_id", "p256dh", "auth", "user_agent", "updated_at"}),
		}).
		Create(sub).Error
}

// DeleteSubscription removes a subscription for a user/endpoint pair.
func (r *PushRepository) DeleteSubscription(ctx context.Context, userID uint, endpoint string) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND endpoint = ?", userID, strings.TrimSpace(endpoint)).
		Delete(&models.PushSubscription{})
	return result.RowsAffected, result.Error
}

// ListByUser returns push subscriptions for a user.
func (r *PushRepository) ListByUser(ctx context.Context, userID uint) ([]models.PushSubscription, error) {
	var subs []models.PushSubscription
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&subs).Error
	return subs, err
}

// DeleteByID removes a subscription by primary key.
func (r *PushRepository) DeleteByID(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.PushSubscription{}).Error
}

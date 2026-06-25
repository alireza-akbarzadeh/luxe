package postgres

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// NotificationRepository persists notifications and chat messages with GORM.
type NotificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository creates a GORM-backed notification repository.
func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create inserts a notification.
func (r *NotificationRepository) Create(ctx context.Context, notification *models.Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

// CountForUser counts notifications for a user.
func (r *NotificationRepository) CountForUser(ctx context.Context, userID uint) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&models.Notification{}).Where("user_id = ?", userID).Count(&total).Error
	return total, err
}

// ListForUser returns paginated notifications for a user.
func (r *NotificationRepository) ListForUser(ctx context.Context, userID uint, limit, offset int) ([]models.Notification, error) {
	var notifications []models.Notification
	err := r.db.WithContext(ctx).Model(&models.Notification{}).Where("user_id = ?", userID).
		Order("created_at DESC").Limit(limit).Offset(offset).
		Find(&notifications).Error
	return notifications, err
}

// MarkRead marks a notification as read for a user.
func (r *NotificationRepository) MarkRead(ctx context.Context, notificationID, userID uint) (int64, error) {
	result := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Update("is_read", true)
	return result.RowsAffected, result.Error
}

// MarkAllRead marks all unread notifications as read for a user.
func (r *NotificationRepository) MarkAllRead(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error
}

// CreateChatRoom inserts a chat room.
func (r *NotificationRepository) CreateChatRoom(ctx context.Context, room *models.ChatRoom) error {
	return r.db.WithContext(ctx).Create(room).Error
}

// CreateMessage inserts a chat message.
func (r *NotificationRepository) CreateMessage(ctx context.Context, message *models.Message) error {
	return r.db.WithContext(ctx).Create(message).Error
}

// TouchChatRoom updates last_message_at for a room.
func (r *NotificationRepository) TouchChatRoom(ctx context.Context, roomID string) error {
	return r.db.WithContext(ctx).Model(&models.ChatRoom{}).
		Where("room_id = ?", roomID).
		Update("last_message_at", gorm.Expr("NOW()")).Error
}

// ListMessages returns chat messages for a room.
func (r *NotificationRepository) ListMessages(ctx context.Context, roomID string, limit, offset int) ([]models.Message, error) {
	var messages []models.Message
	err := r.db.WithContext(ctx).Where("room_id = ?", roomID).
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error
	return messages, err
}

// FindChatRoomByRoomAndUser loads a chat room owned by a user.
func (r *NotificationRepository) FindChatRoomByRoomAndUser(ctx context.Context, roomID string, userID uint) (*models.ChatRoom, error) {
	var room models.ChatRoom
	err := r.db.WithContext(ctx).Where("room_id = ? AND user_id = ?", roomID, userID).First(&room).Error
	if err != nil {
		return nil, err
	}
	return &room, nil
}

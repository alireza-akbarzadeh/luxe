package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

// Commands orchestrates notification write use cases.
type Commands struct {
	repo *postgres.NotificationRepository
}

// NewCommands creates notification command use cases.
func NewCommands(repo *postgres.NotificationRepository) *Commands {
	return &Commands{repo: repo}
}

// Create persists a notification record.
func (c *Commands) Create(ctx context.Context, userID uint, notificationType, title, message string, data interface{}) (*models.Notification, error) {
	var dataStr string
	if data != nil {
		if jsonData, err := json.Marshal(data); err == nil {
			dataStr = string(jsonData)
		}
	}

	notification := &models.Notification{
		UserID:  userID,
		Type:    notificationType,
		Title:   title,
		Message: message,
		Data:    dataStr,
		IsRead:  false,
	}

	if err := c.repo.Create(ctx, notification); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return notification, nil
}

// MarkAsRead marks a notification as read.
func (c *Commands) MarkAsRead(ctx context.Context, notificationID, userID uint) error {
	rows, err := c.repo.MarkRead(ctx, notificationID, userID)
	if err != nil {
		return utils.ErrInternal(err)
	}
	if rows == 0 {
		return utils.ErrNotFound("notification not found")
	}
	return nil
}

// MarkAllAsRead marks all notifications as read for a user.
func (c *Commands) MarkAllAsRead(ctx context.Context, userID uint) error {
	if err := c.repo.MarkAllRead(ctx, userID); err != nil {
		return utils.ErrInternal(err)
	}
	return nil
}

// CreateChatRoom creates a support chat room.
func (c *Commands) CreateChatRoom(ctx context.Context, userID uint, title string) (*models.ChatRoom, error) {
	roomID := fmt.Sprintf("chat_%d_%d", userID, time.Now().Unix())
	chatRoom := &models.ChatRoom{
		RoomID:        roomID,
		UserID:        userID,
		Title:         title,
		Status:        constants.ChatRoomStatusActive,
		LastMessageAt: time.Now(),
	}
	if err := c.repo.CreateChatRoom(ctx, chatRoom); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return chatRoom, nil
}

// SendChatMessage persists and returns a chat message.
func (c *Commands) SendChatMessage(ctx context.Context, senderID uint, roomID, content string) (*models.Message, error) {
	message := &models.Message{
		SenderID: senderID,
		RoomID:   roomID,
		Content:  content,
		Type:     "text",
		IsRead:   false,
	}
	if err := c.repo.CreateMessage(ctx, message); err != nil {
		return nil, utils.ErrInternal(err)
	}
	_ = c.repo.TouchChatRoom(ctx, roomID)
	return message, nil
}

package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/alireza-akbarzadeh/luxe/internal/websocket"
	"gorm.io/gorm"
)

type NotificationServiceInterface interface {
	CreateNotification(userID uint, notificationType, title, message string, data interface{}) error
	GetUserNotifications(userID uint, limit, offset int) ([]models.Notification, int64, error)
	MarkAsRead(notificationID uint, userID uint) error
	MarkAllAsRead(userID uint) error
	BroadcastToUser(userID uint, eventType string, data interface{})
	BroadcastToRoom(roomID string, eventType string, data interface{})
	CreateChatRoom(userID uint, title string) (*models.ChatRoom, error)
	SendChatMessage(senderID uint, roomID string, content string) error
	GetChatMessages(roomID string, limit, offset int) ([]models.Message, error)
}

type notificationService struct {
	db    *gorm.DB
	wsHub *websocket.Hub
}

// NewNotificationService creates a new notification service
func NewNotificationService(db *gorm.DB, wsHub *websocket.Hub) NotificationServiceInterface {
	return &notificationService{
		db:    db,
		wsHub: wsHub,
	}
}

// CreateNotification creates and sends a notification to a user
func (s *notificationService) CreateNotification(userID uint, notificationType, title, message string, data interface{}) error {
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

	if err := s.db.Create(notification).Error; err != nil {
		return utils.ErrInternal(err)
	}

	// Send real-time notification via WebSocket
	s.BroadcastToUser(userID, websocket.EventNotification, map[string]interface{}{
		"id":         notification.ID,
		"type":       notificationType,
		"title":      title,
		"message":    message,
		"data":       data,
		"created_at": notification.CreatedAt,
	})

	return nil
}

// GetUserNotifications retrieves paginated notifications for a user
func (s *notificationService) GetUserNotifications(userID uint, limit, offset int) ([]models.Notification, int64, error) {
	var notifications []models.Notification
	var total int64

	query := s.db.Model(&models.Notification{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, utils.ErrInternal(err)
	}

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&notifications).Error; err != nil {
		return nil, 0, utils.ErrInternal(err)
	}

	return notifications, total, nil
}

// MarkAsRead marks a specific notification as read
func (s *notificationService) MarkAsRead(notificationID uint, userID uint) error {
	result := s.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Update("is_read", true)

	if result.Error != nil {
		return utils.ErrInternal(result.Error)
	}

	if result.RowsAffected == 0 {
		return utils.ErrNotFound("notification not found")
	}

	return nil
}

// MarkAllAsRead marks all notifications as read for a user
func (s *notificationService) MarkAllAsRead(userID uint) error {
	if err := s.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error; err != nil {
		return utils.ErrInternal(err)
	}

	return nil
}

// BroadcastToUser sends a WebSocket message to all connections of a user
func (s *notificationService) BroadcastToUser(userID uint, eventType string, data interface{}) {
	message := websocket.Message{
		Type:      eventType,
		UserID:    userID,
		Data:      data,
		Timestamp: time.Now(),
	}

	if jsonData, err := json.Marshal(message); err == nil {
		s.wsHub.SendToUser(userID, jsonData)
	}
}

// BroadcastToRoom sends a WebSocket message to all clients in a room
func (s *notificationService) BroadcastToRoom(roomID string, eventType string, data interface{}) {
	message := websocket.Message{
		Type:      eventType,
		RoomID:    roomID,
		Data:      data,
		Timestamp: time.Now(),
	}

	if jsonData, err := json.Marshal(message); err == nil {
		s.wsHub.SendToRoom(roomID, jsonData)
	}
}

// CreateChatRoom creates a new chat room for user support
func (s *notificationService) CreateChatRoom(userID uint, title string) (*models.ChatRoom, error) {
	roomID := fmt.Sprintf("chat_%d_%d", userID, time.Now().Unix())

	chatRoom := &models.ChatRoom{
		RoomID:        roomID,
		UserID:        userID,
		Title:         title,
		Status:        constants.ChatRoomStatusActive,
		LastMessageAt: time.Now(),
	}

	if err := s.db.Create(chatRoom).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	return chatRoom, nil
}

// SendChatMessage sends a chat message and broadcasts it via WebSocket
func (s *notificationService) SendChatMessage(senderID uint, roomID string, content string) error {
	message := &models.Message{
		SenderID: senderID,
		RoomID:   roomID,
		Content:  content,
		Type:     "text",
		IsRead:   false,
	}

	if err := s.db.Create(message).Error; err != nil {
		return utils.ErrInternal(err)
	}

	// Update chat room's last message time
	s.db.Model(&models.ChatRoom{}).
		Where("room_id = ?", roomID).
		Update("last_message_at", time.Now())

	// Broadcast message to room
	s.BroadcastToRoom(roomID, websocket.EventChatMessage, map[string]interface{}{
		"id":         message.ID,
		"sender_id":  senderID,
		"content":    content,
		"type":       "text",
		"created_at": message.CreatedAt,
	})

	return nil
}

// GetChatMessages retrieves chat messages for a room
func (s *notificationService) GetChatMessages(roomID string, limit, offset int) ([]models.Message, error) {
	var messages []models.Message

	if err := s.db.Where("room_id = ?", roomID).
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	return messages, nil
}

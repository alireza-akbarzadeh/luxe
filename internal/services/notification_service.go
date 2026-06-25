package services

import (
	"context"
	"encoding/json"
	"time"

	appnotification "github.com/alireza-akbarzadeh/luxe/internal/application/notification"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
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
	commands *appnotification.Commands
	queries  *appnotification.Queries
	wsHub    *websocket.Hub
	push     PushServiceInterface
}

func NewNotificationService(db *gorm.DB, wsHub *websocket.Hub, push PushServiceInterface) NotificationServiceInterface {
	repo := postgres.NewNotificationRepository(db)
	return &notificationService{
		commands: appnotification.NewCommands(repo),
		queries:  appnotification.NewQueries(repo),
		wsHub:    wsHub,
		push:     push,
	}
}

func (s *notificationService) CreateNotification(userID uint, notificationType, title, message string, data interface{}) error {
	ctx := context.Background()
	notification, err := s.commands.Create(ctx, userID, notificationType, title, message, data)
	if err != nil {
		return err
	}

	s.BroadcastToUser(userID, websocket.EventNotification, map[string]interface{}{
		"id":         notification.ID,
		"type":       notificationType,
		"title":      title,
		"message":    message,
		"data":       data,
		"created_at": notification.CreatedAt,
	})

	if s.push != nil && s.push.Enabled() {
		go func(uid uint, nType, nTitle, nMessage string, nData interface{}, nID uint) {
			_ = s.push.SendNotificationToUser(uid, nType, nTitle, nMessage, nData, nID)
		}(userID, notificationType, title, message, data, notification.ID)
	}

	return nil
}

func (s *notificationService) GetUserNotifications(userID uint, limit, offset int) ([]models.Notification, int64, error) {
	return s.queries.ListForUser(context.Background(), userID, limit, offset)
}

func (s *notificationService) MarkAsRead(notificationID uint, userID uint) error {
	return s.commands.MarkAsRead(context.Background(), notificationID, userID)
}

func (s *notificationService) MarkAllAsRead(userID uint) error {
	return s.commands.MarkAllAsRead(context.Background(), userID)
}

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

func (s *notificationService) CreateChatRoom(userID uint, title string) (*models.ChatRoom, error) {
	return s.commands.CreateChatRoom(context.Background(), userID, title)
}

func (s *notificationService) SendChatMessage(senderID uint, roomID string, content string) error {
	message, err := s.commands.SendChatMessage(context.Background(), senderID, roomID, content)
	if err != nil {
		return err
	}

	s.BroadcastToRoom(roomID, websocket.EventChatMessage, map[string]interface{}{
		"id":         message.ID,
		"sender_id":  senderID,
		"content":    content,
		"type":       "text",
		"created_at": message.CreatedAt,
	})

	return nil
}

func (s *notificationService) GetChatMessages(roomID string, limit, offset int) ([]models.Message, error) {
	return s.queries.ListChatMessages(context.Background(), roomID, limit, offset)
}

package notification

import (
	"context"
	"encoding/json"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/websocket"
	"gorm.io/gorm"
)

// PushSender delivers web push notifications to subscribed devices.
type PushSender interface {
	Enabled() bool
	SendNotificationToUser(userID uint, notificationType, title, message string, data interface{}, notificationID uint) error
}

// Service orchestrates in-app notifications and WebSocket broadcasts.
type Service struct {
	commands *Commands
	queries  *Queries
	wsHub    *websocket.Hub
	push     PushSender
}

// NewService wires notification commands and queries.
func NewService(db *gorm.DB, wsHub *websocket.Hub, push PushSender) *Service {
	repo := postgres.NewNotificationRepository(db)
	return &Service{
		commands: NewCommands(repo),
		queries:  NewQueries(repo),
		wsHub:    wsHub,
		push:     push,
	}
}

func (s *Service) CreateNotification(userID uint, notificationType, title, message string, data interface{}) error {
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

func (s *Service) GetUserNotifications(userID uint, limit, offset int) ([]models.Notification, int64, error) {
	return s.queries.ListForUser(context.Background(), userID, limit, offset)
}

func (s *Service) MarkAsRead(notificationID uint, userID uint) error {
	return s.commands.MarkAsRead(context.Background(), notificationID, userID)
}

func (s *Service) MarkAllAsRead(userID uint) error {
	return s.commands.MarkAllAsRead(context.Background(), userID)
}

func (s *Service) BroadcastToUser(userID uint, eventType string, data interface{}) {
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

func (s *Service) BroadcastToRoom(roomID string, eventType string, data interface{}) {
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

func (s *Service) CreateChatRoom(userID uint, title string) (*models.ChatRoom, error) {
	return s.commands.CreateChatRoom(context.Background(), userID, title)
}

func (s *Service) SendChatMessage(senderID uint, roomID string, content string) error {
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

func (s *Service) GetChatMessages(roomID string, limit, offset int) ([]models.Message, error) {
	return s.queries.ListChatMessages(context.Background(), roomID, limit, offset)
}

// GetChatRoomForUser loads a chat room when owned by the user.
func (s *Service) GetChatRoomForUser(userID uint, roomID string) (*models.ChatRoom, error) {
	return s.queries.FindChatRoomForUser(context.Background(), roomID, userID)
}

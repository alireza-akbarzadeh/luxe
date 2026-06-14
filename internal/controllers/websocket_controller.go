package controllers

import (
	"net/http"
	"strconv"

	"github.com/alireza-akbarzadeh/luxe/internal/middleware"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/services"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/alireza-akbarzadeh/luxe/internal/websocket"
	"github.com/gin-gonic/gin"
)

type WebSocketController struct {
	services *services.Services
	handler  *websocket.Handler
}

func NewWebSocketController(services *services.Services) *WebSocketController {
	return &WebSocketController{
		services: services,
		handler:  websocket.NewHandler(services.WebSocketHub),
	}
}

// Connect handles WebSocket connection upgrade
func (wc *WebSocketController) Connect(c *gin.Context) {
	wc.handler.HandleConnection(c)
}

// GetNotifications retrieves user notifications
func (wc *WebSocketController) GetNotifications(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, "authentication required")
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	notifications, total, err := wc.services.Notification.GetUserNotifications(userID, limit, offset)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, "notifications retrieved successfully", gin.H{
		"notifications": notifications,
		"total":         total,
		"limit":         limit,
		"offset":        offset,
	})
}

// MarkNotificationAsRead marks a notification as read
func (wc *WebSocketController) MarkNotificationAsRead(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, "authentication required")
		return
	}

	notificationIDStr := c.Param("id")
	notificationID, err := strconv.ParseUint(notificationIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid notification ID")
		return
	}

	if err := wc.services.Notification.MarkAsRead(uint(notificationID), userID); err != nil {
		if err.Error() == "notification not found" {
			utils.NotFoundResponse(c, "notification not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, "notification marked as read", nil)
}

// MarkAllNotificationsAsRead marks all user notifications as read
func (wc *WebSocketController) MarkAllNotificationsAsRead(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, "authentication required")
		return
	}

	if err := wc.services.Notification.MarkAllAsRead(userID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, "all notifications marked as read", nil)
}

// SendTestNotification creates a notification and pushes it over WebSocket (admin only).
func (wc *WebSocketController) SendTestNotification(c *gin.Context) {
	var req struct {
		UserID  uint                   `json:"user_id" binding:"required"`
		Type    string                 `json:"type"`
		Title   string                 `json:"title"`
		Message string                 `json:"message"`
		Data    map[string]interface{} `json:"data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if req.Type == "" {
		req.Type = "order_update"
	}
	if req.Title == "" {
		req.Title = "Test notification"
	}
	if req.Message == "" {
		req.Message = "This is a test notification from the API."
	}
	if req.Data == nil {
		req.Data = map[string]interface{}{"source": "test_endpoint"}
	}

	if err := wc.services.Notification.CreateNotification(
		req.UserID,
		req.Type,
		req.Title,
		req.Message,
		req.Data,
	); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, "test notification sent successfully", gin.H{
		"user_id": req.UserID,
		"type":    req.Type,
		"title":   req.Title,
		"message": req.Message,
	})
}

// CreateChatRoom creates a new chat room for customer support
func (wc *WebSocketController) CreateChatRoom(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, "authentication required")
		return
	}

	var req struct {
		Title string `json:"title" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	chatRoom, err := wc.services.Notification.CreateChatRoom(userID, req.Title)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, "chat room created successfully", gin.H{
		"chat_room": chatRoom,
	})
}

// SendChatMessage sends a chat message
func (wc *WebSocketController) SendChatMessage(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, "authentication required")
		return
	}

	roomID := c.Param("room_id")

	var req struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := wc.services.Notification.SendChatMessage(userID, roomID, req.Content); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, "message sent successfully", nil)
}

// GetChatMessages retrieves chat messages for a room
func (wc *WebSocketController) GetChatMessages(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, "authentication required")
		return
	}

	roomID := c.Param("room_id")

	// Verify user has access to this room
	var chatRoom models.ChatRoom
	if err := wc.services.DB.Where("room_id = ? AND user_id = ?", roomID, userID).First(&chatRoom).Error; err != nil {
		utils.NotFoundResponse(c, "chat room not found")
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	messages, err := wc.services.Notification.GetChatMessages(roomID, limit, offset)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, "messages retrieved successfully", gin.H{
		"messages": messages,
		"room_id":  roomID,
	})
}

package websocket

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/middleware"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // adjust for production
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Handler struct {
	hub *Hub
}

func NewHandler(hub *Hub) *Handler {
	return &Handler{hub: hub}
}

func (h *Handler) HandleConnection(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, "authentication required for WebSocket connection")
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		utils.Log.WithError(err).Error("Failed to upgrade connection to WebSocket")
		return
	}

	client := &Client{
		ID:     uint(time.Now().UnixNano()),
		UserID: userID,
		Conn:   conn,
		Hub:    h.hub,
		Send:   make(chan []byte, 256),
		Rooms:  make(map[string]bool),
	}

	h.hub.register <- client

	welcomeMsg := Message{
		Type:      EventUserOnline,
		UserID:    userID,
		Data:      map[string]interface{}{"message": "Connected to WebSocket"},
		Timestamp: time.Now(),
	}
	if data, err := json.Marshal(welcomeMsg); err == nil {
		client.Send <- data
	}

	go h.writePump(client)
	go h.readPump(client)
}

// readPump handles incoming messages from the client
func (h *Handler) readPump(client *Client) {
	defer func() {
		h.hub.unregister <- client
		client.Conn.Close()
	}()

	client.Conn.SetReadLimit(512)
	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				utils.Log.WithError(err).WithField("user_id", client.UserID).Error("WebSocket read error")
			}
			break
		}

		// Handle incoming message
		h.handleMessage(client, message)
	}
}

// writePump handles outgoing messages to the client
func (h *Handler) writePump(client *Client) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				utils.Log.WithError(err).WithField("user_id", client.UserID).Error("WebSocket write error")
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (h *Handler) handleMessage(client *Client, data []byte) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		utils.Log.WithError(err).WithField("user_id", client.UserID).Error("Failed to parse WebSocket message")
		return
	}

	switch msg.Type {
	case "join_room":
		if roomID, ok := msg.Data.(string); ok {
			h.hub.JoinRoom(client, roomID)
			utils.Log.WithFields(map[string]interface{}{
				"user_id": client.UserID,
				"room_id": roomID,
			}).Info("Client joined room")
		}

	case "leave_room":
		if roomID, ok := msg.Data.(string); ok {
			h.hub.LeaveRoom(client, roomID)
			utils.Log.WithFields(map[string]interface{}{
				"user_id": client.UserID,
				"room_id": roomID,
			}).Info("Client left room")
		}

	case "ping":
		// Respond with pong
		pongMsg := Message{
			Type:      "pong",
			UserID:    client.UserID,
			Timestamp: time.Now(),
		}
		if data, err := json.Marshal(pongMsg); err == nil {
			client.Send <- data
		}

	default:
		utils.Log.WithFields(map[string]interface{}{
			"user_id":  client.UserID,
			"msg_type": msg.Type,
		}).Warn("Unknown WebSocket message type")
	}
}

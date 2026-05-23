package websocket

import "time"

type Message struct {
	Type      string      `json:"type"`
	UserID    uint        `json:"user_id,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	RoomID    string      `json:"room_id,omitempty"`
}

const (
	EventNotification = "notification"
	EventChatMessage  = "chat_message"
)

// SendToUser sends raw bytes to all WebSocket connections belonging to a user.
func (h *Hub) SendToUser(userID uint, data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		if client.UserID == userID {
			select {
			case client.Send <- data:
			default:
				// skip if client's buffer is full
			}
		}
	}
}

// SendToRoom sends raw bytes to all clients joined to a specific room.
func (h *Hub) SendToRoom(roomID string, data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.rooms[roomID]
	if !ok {
		return
	}
	for client := range clients {
		select {
		case client.Send <- data:
		default:
		}
	}
}

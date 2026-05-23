package websocket

import (
	"encoding/json"
	"sync"

	"github.com/alireza-akbarzadeh/luxe/internal/utils"
)

type Hub struct {
	// Connected clients
	clients map[*Client]bool

	// Rooms maps room ID -> set of clients
	rooms map[string]map[*Client]bool

	// Inbound messages from clients (if needed)
	broadcast chan []byte

	// Register / unregister requests
	register   chan *Client
	unregister chan *Client

	mu sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		rooms:      make(map[string]map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				// Remove client from all rooms
				for room := range client.Rooms {
					h.removeClientFromRoom(client, room)
				}
				close(client.Send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			// broadcast to all clients (optional)
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
				}
			}
			h.mu.RUnlock()
		}
	}
}

// JoinRoom adds a client to a chat room.
func (h *Hub) JoinRoom(client *Client, roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms[roomID] == nil {
		h.rooms[roomID] = make(map[*Client]bool)
	}
	h.rooms[roomID][client] = true
	client.Rooms[roomID] = true
}

// LeaveRoom removes a client from a room.
func (h *Hub) LeaveRoom(client *Client, roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.removeClientFromRoom(client, roomID)
}

// removeClientFromRoom (caller must hold lock)
func (h *Hub) removeClientFromRoom(client *Client, roomID string) {
	if clients, ok := h.rooms[roomID]; ok {
		if _, exists := clients[client]; exists {
			delete(clients, client)
			delete(client.Rooms, roomID)
			if len(clients) == 0 {
				delete(h.rooms, roomID)
			}
		}
	}
}

// BroadcastToRoom sends a Message to all clients in the specified room.
func (h *Hub) BroadcastToRoom(roomID string, msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		utils.Log.WithError(err).Error("Failed to marshal broadcast message")
		return
	}

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
			// Client send buffer full, skip
		}
	}
}

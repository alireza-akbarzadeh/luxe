package websocket

import (
	"encoding/json"
	"sync"

	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

type RoomChangeHook func(roomID string, clientCount int)

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

	onRoomChange RoomChangeHook

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

// SetRoomChangeHook notifies when room membership changes (e.g. live feed viewer count).
func (h *Hub) SetRoomChangeHook(hook RoomChangeHook) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.onRoomChange = hook
}

// ClientCountInRoom returns the number of connected clients in a room.
func (h *Hub) ClientCountInRoom(roomID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.rooms[roomID])
}

func (h *Hub) emitRoomChange(roomID string) {
	h.mu.RLock()
	hook := h.onRoomChange
	count := len(h.rooms[roomID])
	h.mu.RUnlock()

	if hook != nil {
		hook(roomID, count)
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
			changedRooms := make([]string, 0, len(client.Rooms))
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				for room := range client.Rooms {
					changedRooms = append(changedRooms, room)
				}
				for _, room := range changedRooms {
					h.removeClientFromRoom(client, room)
				}
				close(client.Send)
			}
			h.mu.Unlock()
			for _, room := range changedRooms {
				h.emitRoomChange(room)
			}

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
	if h.rooms[roomID] == nil {
		h.rooms[roomID] = make(map[*Client]bool)
	}
	h.rooms[roomID][client] = true
	client.Rooms[roomID] = true
	h.mu.Unlock()
	h.emitRoomChange(roomID)
}

// LeaveRoom removes a client from a room.
func (h *Hub) LeaveRoom(client *Client, roomID string) {
	h.mu.Lock()
	h.removeClientFromRoom(client, roomID)
	h.mu.Unlock()
	h.emitRoomChange(roomID)
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

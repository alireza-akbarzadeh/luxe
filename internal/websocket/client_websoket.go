package websocket

import "github.com/gorilla/websocket"

const (
	EventUserOnline = "user_online"
)

type Client struct {
	ID     uint
	UserID uint
	Conn   *websocket.Conn
	Hub    *Hub
	Send   chan []byte
	Rooms  map[string]bool
}

package salesfeed

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/websocket"
)

// Service publishes live dashboard events to admin clients.
type Service struct {
	hub             *websocket.Hub
	lastViewerCount int
}

// NewService creates a sales feed publisher.
func NewService(hub *websocket.Hub) *Service {
	return &Service{hub: hub}
}

func (s *Service) publish(payload map[string]interface{}) {
	if s.hub == nil {
		return
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	s.hub.SendToRoom(websocket.SalesFeedRoom, data)
}

// PublishOrderEvent emits a sales-feed activity event for the live dashboard.
func (s *Service) PublishOrderEvent(
	eventType, title, subtitle string,
	amount float64,
) {
	s.publish(map[string]interface{}{
		"type": "event",
		"payload": map[string]interface{}{
			"id":        fmt.Sprintf("%d", time.Now().UnixNano()),
			"type":      eventType,
			"title":     title,
			"subtitle":  subtitle,
			"amount":    amount,
			"timestamp": time.Now().UnixMilli(),
		},
	})
}

// PublishRevenueSnapshot emits a chart point for the live dashboard.
func (s *Service) PublishRevenueSnapshot(revenue float64, orders int) {
	now := time.Now()
	s.publish(map[string]interface{}{
		"type": "revenue_snapshot",
		"payload": map[string]interface{}{
			"time":    fmt.Sprintf("%02d:%02d:%02d", now.Hour(), now.Minute(), now.Second()),
			"revenue": revenue,
			"orders":  orders,
		},
	})
}

// PublishActiveUsers emits the number of admin clients watching the live feed.
func (s *Service) PublishActiveUsers(count int) {
	if count == s.lastViewerCount {
		return
	}
	s.lastViewerCount = count
	s.publish(map[string]interface{}{
		"type":    "active_users",
		"payload": count,
	})
}

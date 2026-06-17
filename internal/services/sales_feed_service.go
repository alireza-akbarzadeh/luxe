package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/websocket"
)

// SalesFeedService publishes live dashboard events to admin clients.
type SalesFeedService struct {
	hub *websocket.Hub
}

func NewSalesFeedService(hub *websocket.Hub) *SalesFeedService {
	return &SalesFeedService{hub: hub}
}

func (s *SalesFeedService) publish(payload map[string]interface{}) {
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
func (s *SalesFeedService) PublishOrderEvent(
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
func (s *SalesFeedService) PublishRevenueSnapshot(revenue float64, orders int) {
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
func (s *SalesFeedService) PublishActiveUsers(count int) {
	s.publish(map[string]interface{}{
		"type":    "active_users",
		"payload": count,
	})
}

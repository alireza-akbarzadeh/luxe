package services

import (
	appsalesfeed "github.com/alireza-akbarzadeh/luxe/internal/application/salesfeed"
	"github.com/alireza-akbarzadeh/luxe/internal/websocket"
)

// SalesFeedService publishes live dashboard events to admin clients.
type SalesFeedService struct {
	inner *appsalesfeed.Service
}

func NewSalesFeedService(hub *websocket.Hub) *SalesFeedService {
	return &SalesFeedService{inner: appsalesfeed.NewService(hub)}
}

func (s *SalesFeedService) PublishOrderEvent(
	eventType, title, subtitle string,
	amount float64,
) {
	s.inner.PublishOrderEvent(eventType, title, subtitle, amount)
}

func (s *SalesFeedService) PublishRevenueSnapshot(revenue float64, orders int) {
	s.inner.PublishRevenueSnapshot(revenue, orders)
}

func (s *SalesFeedService) PublishActiveUsers(count int) {
	s.inner.PublishActiveUsers(count)
}

package bootstrap

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/application/apps"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/asynq"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/alireza-akbarzadeh/luxe/internal/websocket"
	"gorm.io/gorm"
)

// Runtime holds shared infrastructure wired at bootstrap (WebSocket hub, DB, apps).
type Runtime struct {
	DB           *gorm.DB
	Apps         *apps.Applications
	WebSocketHub *websocket.Hub
}

// JobHandlers wires application services into background task handlers.
func (r *Runtime) JobHandlers() asynq.Handlers {
	return asynq.Handlers{
		ProcessOrder: func(ctx context.Context, orderID uint, cardInfo dto.CardInfo) error {
			return r.Apps.Checkout.ProcessOrder(ctx, orderID, cardInfo)
		},
		ProcessShipment: func(ctx context.Context, shipmentID uint) error {
			return r.Apps.Shipment.ProcessShipmentBackground(ctx, shipmentID)
		},
		SendEmail: func(_ context.Context, to, subject, body string) error {
			utils.SendEmailDirect(to, subject, body)
			return nil
		},
	}
}

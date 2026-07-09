package order

import (
	"context"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

// Commands orchestrates order persistence writes (workflow orchestration stays in legacy service).
type Commands struct {
	writer Writer
}

// NewCommands creates order command use cases.
func NewCommands(writer Writer) *Commands {
	return &Commands{writer: writer}
}

// UpdateStatus writes order status directly when workflow does not handle it.
func (c *Commands) UpdateStatus(ctx context.Context, orderID uint, status string) error {
	return c.writer.UpdateStatus(ctx, orderID, status)
}

// UpdateStatusByIDs bulk-updates order status for the given ids.
func (c *Commands) UpdateStatusByIDs(ctx context.Context, orderIDs []uint, status string) (int64, error) {
	return c.writer.UpdateStatusByIDs(ctx, orderIDs, status)
}

// FindOverduePaid loads paid orders older than the cutoff that are not terminal.
func (c *Commands) FindOverduePaid(ctx context.Context) ([]models.Order, error) {
	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	excluded := []string{
		constants.OrderStatusDelivered,
		constants.OrderStatusCancelled,
		constants.OrderStatusRefunded,
	}
	return c.writer.FindOverduePaid(ctx, cutoff, excluded)
}

// SaveDelayed persists a delayed status on an order row.
func (c *Commands) SaveDelayed(ctx context.Context, order *models.Order) error {
	order.Status = constants.OrderStatusDelayed
	return c.writer.Save(ctx, order)
}

// MarkShipmentShipped records carrier tracking when an order is marked shipped.
func (c *Commands) MarkShipmentShipped(ctx context.Context, orderID uint, trackingNumber string) error {
	updates := map[string]interface{}{
		"status":     constants.ShipmentStatusShipped,
		"shipped_at": time.Now(),
	}
	if trackingNumber != "" {
		updates["tracking_number"] = trackingNumber
	}
	return c.writer.UpdateShipmentByOrderID(ctx, orderID, updates)
}

// UpdateNotes replaces admin notes on an order.
func (c *Commands) UpdateNotes(ctx context.Context, orderID uint, notes string) error {
	return c.writer.UpdateNotes(ctx, orderID, notes)
}

// ReplaceTags replaces all tags on an order.
func (c *Commands) ReplaceTags(ctx context.Context, orderID uint, tags []string) error {
	return c.writer.ReplaceTags(ctx, orderID, tags)
}

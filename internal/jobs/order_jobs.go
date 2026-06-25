package jobs

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

func (c *CronJobs) registerOrderJobs() {
	c.addJob(constants.CronUpdateOverdueOrders, "update_overdue_orders", c.updateOverdueOrders)
	c.addJob(constants.CronEvery5Minutes, "simulate_deliveries", c.simulateDeliveries) // new

}

func (c *CronJobs) updateOverdueOrders() {
	utils.Log.Info("Updating overdue orders...")
	if err := c.svc.Order.UpdateOverdueOrders(context.Background()); err != nil {
		utils.Log.WithError(err).Error("Overdue orders update failed")
	}
}

func (c *CronJobs) simulateDeliveries() {
	utils.Log.Info("Simulating deliveries...")
	if err := c.svc.Shipment.SimulateDeliveries(); err != nil {
		utils.Log.WithError(err).Error("Shipment delivery simulation failed")
	}
}

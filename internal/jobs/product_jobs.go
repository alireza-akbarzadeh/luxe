package jobs

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
)

func (c *CronJobs) registerProductJobs() {
	c.addJob(constants.CronLowStockAlert, "low_stock_alert", c.checkLowStock)
	c.addJob(constants.CronSyncProductPrices, "sync_product_prices", c.syncProductPrices) // example
}

func (c *CronJobs) checkLowStock() {
	utils.Log.Info("Checking low stock products...")
	if err := c.svc.Inventory.SendLowStockAlerts(context.Background()); err != nil {
		utils.Log.WithError(err).Error("Low stock check failed")
	}
}

func (c *CronJobs) syncProductPrices() {
	utils.Log.Info("Syncing product prices with external feed...")
	// external API call, update DB, etc.
}

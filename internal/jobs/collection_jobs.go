package jobs

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

func (c *CronJobs) registerCollectionJobs() {
	c.addJob(constants.CronEveryHour, "activate_due_collections", c.activateDueCollections)
	c.addJob(constants.CronEveryHour, "expire_ended_collections", c.expireEndedCollections)
}

func (c *CronJobs) activateDueCollections() {
	utils.Log.Info("Activating due scheduled collections...")
	n, err := c.apps.Collection.ActivateDueCollections(context.Background())
	if err != nil {
		utils.Log.WithError(err).Error("Activate due collections failed")
		return
	}
	if n > 0 {
		utils.Log.WithField("count", n).Info("Activated scheduled collections")
	}
}

func (c *CronJobs) expireEndedCollections() {
	utils.Log.Info("Expiring ended collections...")
	n, err := c.apps.Collection.ExpireEndedCollections(context.Background())
	if err != nil {
		utils.Log.WithError(err).Error("Expire ended collections failed")
		return
	}
	if n > 0 {
		utils.Log.WithField("count", n).Info("Expired active collections")
	}
}

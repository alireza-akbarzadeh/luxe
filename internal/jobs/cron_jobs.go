package jobs

import (
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/application/apps"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/robfig/cron/v3"
)

type CronJobs struct {
	scheduler *cron.Cron
	apps      *apps.Applications
}

// Recoverer is a job wrapper that recovers from panics to prevent the scheduler from crashing.
func Recoverer(next cron.Job) cron.Job {
	return cron.FuncJob(func() {
		defer func() {
			if r := recover(); r != nil {
				utils.Log.WithField("panic", r).Error("Cron job panic recovered")
			}
		}()
		next.Run()
	})
}

func NewCronJobs(apps *apps.Applications) *CronJobs {
	scheduler := cron.New(
		cron.WithLocation(time.UTC),
		cron.WithChain(Recoverer),
	)
	return &CronJobs{
		scheduler: scheduler,
		apps:      apps,
	}
}

// registerJobs centralizes all your scheduled tasks.
func (c *CronJobs) registerJobs() {
	c.registerCartJobs()
	c.registerProductJobs()
	c.registerCollectionJobs()
	c.registerOrderJobs()
}

// addJob is a helper to add a job to the scheduler with error logging.
func (c *CronJobs) addJob(schedule, name string, cmd func()) {
	_, err := c.scheduler.AddFunc(schedule, cmd)
	if err != nil {
		utils.Log.WithFields(map[string]interface{}{
			"job_name": name,
			"schedule": schedule,
		}).WithError(err).Error("Failed to schedule job")
	}
}

func (c *CronJobs) Start() {
	c.registerJobs()
	c.scheduler.Start()
	utils.Log.Info("CronJobs started.")
}

func (c *CronJobs) Stop() {
	ctx := c.scheduler.Stop()
	<-ctx.Done()
	utils.Log.Info("CronJobs stopped.")
}

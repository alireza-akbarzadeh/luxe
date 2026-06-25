package bootstrap

import (
	"github.com/alireza-akbarzadeh/luxe/internal/application/apps"
	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/asynq"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/workflow"
	"gorm.io/gorm"
)

// Applications is the wired application-layer registry (alias for apps.Applications).
type Applications = apps.Applications

// WireApplications constructs all application use cases.
func WireApplications(
	db *gorm.DB,
	cfg *config.Config,
	jobQueue asynq.JobQueue,
	engine *workflow.Engine,
) *Applications {
	return apps.WireApplications(db, cfg, jobQueue, engine)
}

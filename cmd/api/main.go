// Package main is the entry point of the shopping platform API. It initializes configuration, logger, database connection, services, HTTP handlers, routes, and starts the server.
package main

import (
	"fmt"

	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/alireza-akbarzadeh/luxe/internal/i18n"
	"github.com/alireza-akbarzadeh/luxe/internal/jobs"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/observability"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/routes"
	"github.com/alireza-akbarzadeh/luxe/internal/application/bootstrap"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/asynq"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
)

// @title           Shopping Platform API
// @version         1.0
// @description     Production-grade e-commerce backend
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@luxe.com

// @license.name   MIT
// @license.url    https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT token.
func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	if err := utils.InitLoggerWithConfig(utils.LoggerConfig{
		Level:          cfg.Log.Level,
		AppEnv:         cfg.AppEnv,
		ServiceName:    cfg.Observability.ServiceName,
		ServiceVersion: cfg.Observability.ServiceVersion,
	}); err != nil {
		panic(fmt.Sprintf("failed to init logger: %v", err))
	}

	if err := i18n.Init(); err != nil {
		panic(fmt.Sprintf("failed to init i18n: %v", err))
	}

	sentryShutdown, err := observability.Init(cfg, utils.Log)
	if err != nil {
		utils.Log.WithError(err).Fatal("failed to init observability")
	}
	if sentryShutdown != nil {
		defer sentryShutdown()
		utils.Log.WithFields(map[string]interface{}{
			"environment": cfg.Observability.SentryEnvironment,
			"release":     cfg.Observability.ServiceVersion,
		}).Info("Sentry enabled")
	}

	tracingShutdown, err := observability.InitTracing(cfg)
	if err != nil {
		utils.Log.WithError(err).Fatal("failed to init tracing")
	}
	if tracingShutdown != nil {
		defer tracingShutdown()
		utils.Log.Info("OpenTelemetry tracing enabled")
	}

	gin.SetMode(cfg.Server.Mode)

	db := connectDatabase(cfg)
	defer closeDatabase(db)

	jobQueue, err := asynq.NewJobQueue(cfg, asynq.Handlers{})
	if err != nil {
		utils.Log.WithError(err).Fatal("failed to initialize job queue")
	}

	newRuntime := bootstrap.NewRuntime(db, cfg, jobQueue)
	asynq.BindHandlers(jobQueue, newRuntime.JobHandlers())

	if err := jobQueue.Start(); err != nil {
		utils.Log.WithError(err).Fatal("failed to start job queue")
	}

	utils.Log.WithField("job_backend", jobQueue.Backend()).Info("background workers ready")

	if newRuntime.Apps.Upload.IsEnabled() {
		utils.Log.Info("R2 presigned uploads enabled")
	}

	cronService := jobs.NewCronJobs(newRuntime.Apps)
	cronService.Start()

	ctrl := handlers.NewContainer(db, newRuntime.Apps, newRuntime, cfg)
	engine := setupGin()
	router := routes.NewRouter(engine, ctrl, cfg, newRuntime.Apps)
	router.Setup()

	bootStrap(engine, cfg, func() {
		utils.Log.Info("stopping background workers")
		cronService.Stop()
		jobQueue.Shutdown()
	})
}

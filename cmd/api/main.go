// Package main is the entry point of the shopping platform API. It initializes configuration, logger, database connection, services, controllers, routes, and starts the server. It also manages the lifecycle of background workers and cron jobs for asynchronous tasks and scheduled operations.
package main

import (
	"fmt"

	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/controllers"
	"github.com/alireza-akbarzadeh/luxe/internal/jobs"
	"github.com/alireza-akbarzadeh/luxe/internal/routes"
	"github.com/alireza-akbarzadeh/luxe/internal/services"
	"github.com/alireza-akbarzadeh/luxe/internal/tasks"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
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
	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	// 2. Initialize logger
	if err := utils.InitLogger(cfg.Log.Level); err != nil {
		panic(fmt.Sprintf("failed to init logger: %v", err))
	}

	// 3. Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// 4. Connect to database
	db := connectDatabase(cfg)
	defer closeDatabase(db)

	// 1. Start the Worker Pool (for async tasks)
	workerPool := tasks.NewWorkerPool(5, 100)
	workerPool.Start()
	defer workerPool.Stop()

	// 5. Initialize newServices
	newServices := services.NewServices(db, cfg, workerPool)

	// 2. Start the Cron Service (for scheduled tasks)
	cronService := jobs.NewCronJobs(workerPool, newServices)
	cronService.Start()
	defer cronService.Stop()

	// 6. Initialize controllers
	ctrl := controllers.NewContainer(db, newServices, cfg)
	// 7. Setup Gin engine and routes
	engine := setupGin()
	router := routes.NewRouter(engine, ctrl, cfg)
	router.Setup()

	// 8. Start server
	bootStrap(engine, cfg)
}

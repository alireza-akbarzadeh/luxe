// Package bootstrap wires application services, infrastructure, and cross-cutting dependencies.
package bootstrap

import (
	appadmin "github.com/alireza-akbarzadeh/luxe/internal/application/admin"
	appcatalog "github.com/alireza-akbarzadeh/luxe/internal/application/catalog"
	appcheckout "github.com/alireza-akbarzadeh/luxe/internal/application/checkout"
	importdata "github.com/alireza-akbarzadeh/luxe/internal/application/import"
	appinventory "github.com/alireza-akbarzadeh/luxe/internal/application/inventory"
	appnotification "github.com/alireza-akbarzadeh/luxe/internal/application/notification"
	orderfacade "github.com/alireza-akbarzadeh/luxe/internal/application/order/facade"
	apppdp "github.com/alireza-akbarzadeh/luxe/internal/application/pdp"
	apppush "github.com/alireza-akbarzadeh/luxe/internal/application/push"
	appsalesfeed "github.com/alireza-akbarzadeh/luxe/internal/application/salesfeed"
	appshipment "github.com/alireza-akbarzadeh/luxe/internal/application/shipment"
	appworkflow "github.com/alireza-akbarzadeh/luxe/internal/application/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/asynq"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	infraworkflow "github.com/alireza-akbarzadeh/luxe/internal/infrastructure/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/websocket"
	"gorm.io/gorm"
)

// NewRuntime constructs orchestrators and wires application use cases.
func NewRuntime(db *gorm.DB, cfg *config.Config, jobQueue asynq.JobQueue) *Runtime {
	wsHub := websocket.NewHub()
	go wsHub.Run()

	salesFeedSvc := appsalesfeed.NewService(wsHub)
	wsHub.SetRoomChangeHook(func(roomID string, clientCount int) {
		if roomID == websocket.SalesFeedRoom {
			salesFeedSvc.PublishActiveUsers(clientCount)
		}
	})

	pushSvc := apppush.NewWebPushService(db, cfg)
	notificationSvc := appnotification.NewService(db, wsHub, pushSvc)

	workflowEngine := infraworkflow.NewEngine(db)
	apps := WireApplications(db, cfg, jobQueue, workflowEngine)

	workflowHooksRepo := postgres.NewWorkflowHooksRepository(db)
	appworkflow.RegisterGuardsAndHooks(appworkflow.HookDeps{
		Engine:   workflowEngine,
		Repo:     workflowHooksRepo,
		Notify:   notificationSvc,
		Wallet:   apps.Wallet,
		JobQueue: jobQueue,
	})

	productSvc := appcatalog.NewService(db, workflowEngine)
	pdpSvc := apppdp.NewService(db, notificationSvc, productSvc, apps.AI)
	inventorySvc := appinventory.NewService(db, workflowEngine, pdpSvc, notificationSvc, jobQueue, cfg.InventoryAlertEmails)
	productSvc.SetInventory(inventorySvc)
	appworkflow.RegisterInventoryHooks(workflowEngine, inventorySvc)
	shipmentSvc := appshipment.NewService(db, jobQueue, notificationSvc, wsHub, workflowEngine)

	apps.Product = productSvc
	apps.Pdp = pdpSvc
	apps.Inventory = inventorySvc
	apps.Notification = notificationSvc
	apps.Push = pushSvc
	apps.SalesFeed = salesFeedSvc
	apps.Shipment = shipmentSvc
	apps.Checkout = appcheckout.NewService(
		db,
		notificationSvc,
		apps.Coupon,
		apps.Payment,
		shipmentSvc,
		apps.Wallet,
		apps.Invoice.Commands,
		jobQueue,
		wsHub,
		salesFeedSvc,
		workflowEngine,
		inventorySvc,
		appcheckout.StripeEnabled(cfg),
	)
	apps.Order = orderfacade.NewService(db, notificationSvc, wsHub, salesFeedSvc, jobQueue, workflowEngine)
	apps.Admin = appadmin.NewService(db, workflowEngine, apps.Role.Queries)
	apps.Import = importdata.NewService(productSvc, apps.Category)

	return &Runtime{
		DB:           db,
		Apps:         apps,
		WebSocketHub: wsHub,
	}
}

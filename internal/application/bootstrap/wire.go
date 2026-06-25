// Package bootstrap wires application services, infrastructure, and cross-cutting dependencies.
package bootstrap

import (
	appworkflow "github.com/alireza-akbarzadeh/luxe/internal/application/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/asynq"
	infraworkflow "github.com/alireza-akbarzadeh/luxe/internal/infrastructure/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/services"
	"github.com/alireza-akbarzadeh/luxe/internal/websocket"
	"gorm.io/gorm"
)

// NewServices constructs the full service registry (facades + shared infrastructure).
func NewServices(db *gorm.DB, cfg *config.Config, jobQueue asynq.JobQueue) *services.Services {
	wsHub := websocket.NewHub()
	go wsHub.Run()
	salesFeedSvc := services.NewSalesFeedService(wsHub)
	wsHub.SetRoomChangeHook(func(roomID string, clientCount int) {
		if roomID == websocket.SalesFeedRoom {
			salesFeedSvc.PublishActiveUsers(clientCount)
		}
	})

	pushSvc := services.NewPushService(db, cfg)
	notificationSvc := services.NewNotificationService(db, wsHub, pushSvc)
	paymentSvc := services.NewPaymentService(db, cfg)
	walletSvc := services.NewWalletService(db, cfg)

	workflowEngine := infraworkflow.NewEngine(db)
	appworkflow.RegisterGuardsAndHooks(appworkflow.HookDeps{
		Engine:   workflowEngine,
		DB:       db,
		Notify:   notificationSvc,
		Wallet:   walletSvc,
		JobQueue: jobQueue,
	})
	couponSvc := services.NewCouponService(db, workflowEngine)
	roleSvc := services.NewRoleService(db)

	productSvc := services.NewProductService(db, workflowEngine)
	aiSvc := services.NewAiService(db, cfg.AI)
	pdpSvc := services.NewPdpService(db, notificationSvc, productSvc, aiSvc)
	inventorySvc := services.NewInventoryService(db, workflowEngine, pdpSvc, notificationSvc, jobQueue, cfg.InventoryAlertEmails)
	productSvc.SetInventory(inventorySvc)
	appworkflow.RegisterInventoryHooks(workflowEngine, inventorySvc)
	shipmentSvc := services.NewShipmentService(db, jobQueue, notificationSvc, wsHub, workflowEngine)
	invoiceSvc := services.NewInvoiceService(db, jobQueue, cfg)
	categorySvc := services.NewCategoryService(db, workflowEngine)
	settingsSvc := services.NewSettingService(db)

	return &services.Services{
		DB:           db,
		Auth:         services.NewAuthServices(db, cfg, jobQueue, workflowEngine, settingsSvc),
		Search:       services.NewSearchService(db),
		User:         services.NewUserService(db, cfg),
		Cart:         services.NewCartService(db),
		NavMenu:      services.NewNavMenuService(db),
		Product:      productSvc,
		Pdp:          pdpSvc,
		Compare:      services.NewCompareService(db),
		Category:     categorySvc,
		Address:      services.NewAddressService(db),
		Menu:         services.NewMenuService(db),
		Review:       services.NewReviewService(db, workflowEngine),
		UserLike:     services.NewUserLikeService(db),
		Shipment:     shipmentSvc,
		Wallet:       walletSvc,
		Payment:      paymentSvc,
		Store:        services.NewStoreService(db),
		Brand:        services.NewBrandService(db, workflowEngine),
		Collection:   services.NewCollectionService(db, workflowEngine),
		Settings:     settingsSvc,
		Checkout:     services.NewCheckoutService(db, notificationSvc, couponSvc, paymentSvc, shipmentSvc, walletSvc, invoiceSvc, jobQueue, wsHub, salesFeedSvc, workflowEngine, inventorySvc, services.StripeEnabled(cfg)),
		Order:        services.NewOrderService(db, notificationSvc, wsHub, salesFeedSvc, jobQueue, workflowEngine),
		Coupon:       couponSvc,
		Notification: notificationSvc,
		Push:         pushSvc,
		WebSocketHub: wsHub,
		SalesFeed:    salesFeedSvc,
		Audit:        services.NewAuditService(db),
		Upload:       services.NewUploadService(cfg),
		Admin:        services.NewAdminService(db, workflowEngine, roleSvc),
		Import:       services.NewImportService(productSvc, categorySvc),
		WebhookEvent: services.NewWebhookEventService(db),
		Workflow:     services.NewWorkflowService(db, workflowEngine),
		Return:       services.NewReturnService(db, workflowEngine),
		Invoice:      invoiceSvc,
		Role:         roleSvc,
		Inventory:    inventorySvc,
		Ai:           aiSvc,
	}
}

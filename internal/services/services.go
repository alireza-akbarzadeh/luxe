package services

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/services/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/tasks"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/alireza-akbarzadeh/luxe/internal/websocket"
	"gorm.io/gorm"
)

type Services struct {
	DB           *gorm.DB
	Auth         AuthServiceInterface
	User         UserServiceInterface
	Cart         CartServiceInterface
	Product      ProductServiceInterface
	Category     CategoryServiceInterface
	Order        OrderServiceInterface
	Shipment     ShipmentServiceInterface
	Notification NotificationServiceInterface
	WebSocketHub *websocket.Hub
	Coupon       CouponServiceInterface
	Address      AddressServiceInterface
	Menu         UserMenuServicesInterface
	Review       ReviewServiceInterface
	UserLike     UsertLikeServiceInterface
	Wallet       WalletServiceInterface
	Checkout     CheckoutServiceInterface
	Payment      PaymentServiceInterface
	Store        StoreServiceInterface
	Search       SearchServiceInterface
	Compare      CompareServiceInterface
	NavMenu      NavMenuServiceInterface
	Brand        BrandServiceInterface
	Collection   CollectionServiceInterface
	Settings     SettingServiceInterface
	Pdp          PdpServiceInterface
	SalesFeed    *SalesFeedService
	Audit        AuditServiceInterface
	Upload       UploadServiceInterface
	Admin        AdminServiceInterface
	Import       ImportServiceInterface
	WebhookEvent WebhookEventServiceInterface
	Workflow     WorkflowServiceInterface
	Return       ReturnServiceInterface
	Invoice      InvoiceServiceInterface
	Role         RoleServiceInterface
	Inventory    InventoryServiceInterface
}

func NewServices(db *gorm.DB, cfg *config.Config, jobQueue tasks.JobQueue) *Services {
	wsHub := websocket.NewHub()
	go wsHub.Run()
	salesFeedSvc := NewSalesFeedService(wsHub)
	wsHub.SetRoomChangeHook(func(roomID string, clientCount int) {
		if roomID == websocket.SalesFeedRoom {
			salesFeedSvc.PublishActiveUsers(clientCount)
		}
	})

	notificationSvc := NewNotificationService(db, wsHub)
	couponSvc := NewCouponService(db)
	paymentSvc := NewPaymentService(db, cfg)
	walletSvc := NewWalletService(db, cfg)

	workflowEngine := workflow.NewEngine(db)
	RegisterWorkflowGuardsAndHooks(workflowEngine, db, notificationSvc, walletSvc, jobQueue)
	roleSvc := NewRoleService(db)

	productSvc := NewProductService(db, workflowEngine)
	pdpSvc := NewPdpService(db, notificationSvc, productSvc)
	inventorySvc := NewInventoryService(db, workflowEngine, pdpSvc, notificationSvc, jobQueue, cfg.InventoryAlertEmails)
	productSvc.SetInventory(inventorySvc)
	RegisterInventoryWorkflowHooks(workflowEngine, inventorySvc)
	shipmentSvc := NewShipmentService(db, jobQueue, notificationSvc, wsHub, workflowEngine)
	invoiceSvc := NewInvoiceService(db, jobQueue, cfg)
	categorySvc := NewCategoryService(db, workflowEngine)

	return &Services{
		DB:           db,
		Auth:         NewAuthServices(db, cfg, jobQueue, workflowEngine),
		Search:       NewSearchService(db),
		User:         NewUserService(db, cfg),
		Cart:         NewCartService(db),
		NavMenu:      NewNavMenuService(db),
		Product:      productSvc,
		Pdp:          pdpSvc,
		Compare:      NewCompareService(db),
		Category:     categorySvc,
		Address:      NewAddressService(db),
		Menu:         NewMenuService(db),
		Review:       NewReviewService(db),
		UserLike:     NewUserLikeService(db),
		Shipment:     shipmentSvc,
		Wallet:       walletSvc,
		Payment:      paymentSvc,
		Store:        NewStoreService(db),
		Brand:        NewBrandService(db, workflowEngine),
		Collection:   NewCollectionService(db, workflowEngine),
		Settings:     NewSettingService(db),
		Checkout:     NewCheckoutService(db, notificationSvc, couponSvc, paymentSvc, shipmentSvc, walletSvc, invoiceSvc, jobQueue, wsHub, salesFeedSvc, workflowEngine, inventorySvc, StripeEnabled(cfg)),
		Order:        NewOrderService(db, notificationSvc, wsHub, salesFeedSvc, jobQueue, workflowEngine),
		Coupon:       couponSvc,
		Notification: notificationSvc,
		WebSocketHub: wsHub,
		SalesFeed:    salesFeedSvc,
		Audit:        NewAuditService(db),
		Upload:       NewUploadService(cfg),
		Admin:        NewAdminService(db, workflowEngine, roleSvc),
		Import:       NewImportService(productSvc, categorySvc),
		WebhookEvent: NewWebhookEventService(db),
		Workflow:     NewWorkflowService(db, workflowEngine),
		Return:       NewReturnService(db, workflowEngine),
		Invoice:      invoiceSvc,
		Role:         roleSvc,
		Inventory:    inventorySvc,
	}
}

// JobHandlers wires service methods into background task handlers.
func (s *Services) JobHandlers() tasks.Handlers {
	return tasks.Handlers{
		ProcessOrder: func(ctx context.Context, orderID uint, cardInfo dto.CardInfo) error {
			return s.Checkout.ProcessOrder(ctx, orderID, cardInfo)
		},
		ProcessShipment: func(ctx context.Context, shipmentID uint) error {
			return s.Shipment.ProcessShipmentBackground(ctx, shipmentID)
		},
		SendEmail: func(_ context.Context, to, subject, body string) error {
			utils.SendEmailDirect(to, subject, body)
			return nil
		},
	}
}

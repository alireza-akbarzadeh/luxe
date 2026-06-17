// Package services defines the core business logic of the shopping platform.
package services

import (
	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/repositories"
	"github.com/alireza-akbarzadeh/luxe/internal/tasks"
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
	Settings     SettingServiceInterface
	Pdp          PdpServiceInterface
	SalesFeed    *SalesFeedService
	Audit        AuditServiceInterface
}

func NewServices(db *gorm.DB, cfg *config.Config, workerPool *tasks.WorkerPool) *Services {
	// 1. WebSocket hub
	wsHub := websocket.NewHub()
	go wsHub.Run()
	salesFeedSvc := NewSalesFeedService(wsHub)

	// 2. Services that depend on hub
	notificationSvc := NewNotificationService(db, wsHub)
	couponSvc := NewCouponService(db)

	// 3. Payment service
	paymentSvc := NewPaymentService(db, cfg)

	// 4. Shipment service (now also receives the hub for delivery broadcasts)
	shipmentSvc := NewShipmentService(db, workerPool, notificationSvc, wsHub)

	// 5. Order service with all dependencies
	orderSvc := NewOrderService(db, notificationSvc, wsHub, salesFeedSvc)
	checkoutSvc := NewCheckoutService(db, notificationSvc, couponSvc, paymentSvc, shipmentSvc, workerPool, wsHub, salesFeedSvc, StripeEnabled(cfg))
	// 5. Assemble all services
	productSvc := NewProductService(db)
	auditRepo := repositories.NewAuditRepository(db)
	auditSvc := NewAuditService(auditRepo)
	return &Services{
		DB:           db,
		Auth:         NewAuthServices(db, cfg),
		Search:       NewSearchService(db),
		User:         NewUserService(db, cfg),
		Cart:         NewCartService(db),
		NavMenu:      NewNavMenuService(db),
		Product:      productSvc,
		Pdp:          NewPdpService(db, notificationSvc, productSvc),
		Compare:      NewCompareService(db),
		Category:     NewCategoryService(db),
		Address:      NewAddressService(db),
		Menu:         NewMenuService(db),
		Review:       NewReviewService(db),
		UserLike:     NewUserLikeService(db),
		Shipment:     NewShipmentService(db, workerPool, notificationSvc, wsHub),
		Wallet:       NewWalletService(db),
		Payment:      paymentSvc,
		Store:        NewStoreService(db),
		Brand:        NewBrandService(db),
		Settings:     NewSettingService(db),
		Checkout:     checkoutSvc,
		Order:        orderSvc,
		Coupon:       couponSvc,
		Notification: notificationSvc,
		WebSocketHub: wsHub,
		SalesFeed:    salesFeedSvc,
		Audit:        auditSvc,
	}
}

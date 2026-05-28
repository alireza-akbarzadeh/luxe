// Package services defines the core business logic of the shopping platform.
package services

import (
	"github.com/alireza-akbarzadeh/luxe/internal/config"
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
}

func NewServices(db *gorm.DB, cfg *config.Config, workerPool *tasks.WorkerPool) *Services {
	// 1. WebSocket hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// 2. Services that depend on hub
	notificationSvc := NewNotificationService(db, wsHub)
	couponSvc := NewCouponService(db)

	// 3. New payment service (no hub needed)
	paymentSvc := NewPaymentService(db)

	// 4. Shipment service (now also receives the hub for delivery broadcasts)
	shipmentSvc := NewShipmentService(db, workerPool, notificationSvc, wsHub)

	// 5. Order service with all dependencies
	orderSvc := NewOrderService(db, notificationSvc, wsHub)
	checkoutSvc := NewCheckoutService(db, notificationSvc, couponSvc, paymentSvc, shipmentSvc, workerPool, wsHub)
	// 5. Assemble all services
	return &Services{
		DB:           db,
		Auth:         NewAuthServices(db, cfg),
		Search:       NewSearchService(db),
		User:         NewUserService(db, cfg),
		Cart:         NewCartService(db),
		Product:      NewProductService(db),
		Category:     NewCategoryService(db),
		Address:      NewAddressService(db),
		Menu:         NewMenuService(db),
		Review:       NewReviewService(db),
		UserLike:     NewUserLikeService(db),
		Shipment:     NewShipmentService(db, workerPool, notificationSvc, wsHub),
		Wallet:       NewWalletService(db),
		Payment:      NewPaymentService(db),
		Store:        NewStoreService(db),
		Checkout:     checkoutSvc,
		Order:        orderSvc,
		Coupon:       couponSvc,
		Notification: notificationSvc,
		WebSocketHub: wsHub,
	}
}

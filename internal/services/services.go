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
}

func NewServices(db *gorm.DB, cfg *config.Config, workerPool *tasks.WorkerPool) *Services {
	// 1. WebSocket hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	notificationSvc := NewNotificationService(db, wsHub)
	couponSvc := NewCouponService(db)
	orderSvc := NewOrderService(db, notificationSvc, couponSvc)

	// 5. Assemble all services
	return &Services{
		DB:           db,
		Auth:         NewAuthServices(db, cfg),
		User:         NewUserService(db, cfg),
		Cart:         NewCartService(db),
		Product:      NewProductService(db),
		Category:     NewCategoryService(db),
		Address:      NewAddressService(db),
		Menu:         NewMenuService(db),
		Review:       NewReviewService(db),
		UserLike:     NewUserLikeService(db),
		Shipment:     NewShipmentService(db, workerPool, notificationSvc),
		Wallet:       NewWalletService(db),
		Order:        orderSvc,
		Coupon:       couponSvc,
		Notification: notificationSvc,
		WebSocketHub: wsHub,
	}
}

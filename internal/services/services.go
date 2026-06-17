package services

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
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
	Settings     SettingServiceInterface
	Pdp          PdpServiceInterface
	SalesFeed    *SalesFeedService
	Audit        AuditServiceInterface
	Upload       UploadServiceInterface
	Admin        AdminServiceInterface
	Import       ImportServiceInterface
	WebhookEvent WebhookEventServiceInterface
}

func NewServices(db *gorm.DB, cfg *config.Config, jobQueue tasks.JobQueue) *Services {
	wsHub := websocket.NewHub()
	go wsHub.Run()
	salesFeedSvc := NewSalesFeedService(wsHub)

	notificationSvc := NewNotificationService(db, wsHub)
	couponSvc := NewCouponService(db)
	paymentSvc := NewPaymentService(db, cfg)
	shipmentSvc := NewShipmentService(db, jobQueue, notificationSvc, wsHub)
	productSvc := NewProductService(db)

	return &Services{
		DB:           db,
		Auth:         NewAuthServices(db, cfg, jobQueue),
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
		Shipment:     shipmentSvc,
		Wallet:       NewWalletService(db, cfg),
		Payment:      paymentSvc,
		Store:        NewStoreService(db),
		Brand:        NewBrandService(db),
		Settings:     NewSettingService(db),
		Checkout:     NewCheckoutService(db, notificationSvc, couponSvc, paymentSvc, shipmentSvc, jobQueue, wsHub, salesFeedSvc, StripeEnabled(cfg)),
		Order:        NewOrderService(db, notificationSvc, wsHub, salesFeedSvc),
		Coupon:       couponSvc,
		Notification: notificationSvc,
		WebSocketHub: wsHub,
		SalesFeed:    salesFeedSvc,
		Audit:        NewAuditService(db),
		Upload:       NewUploadService(cfg),
		Admin:        NewAdminService(db),
		Import:       NewImportService(productSvc, NewCategoryService(db)),
		WebhookEvent: NewWebhookEventService(db),
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

package services

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/asynq"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/alireza-akbarzadeh/luxe/internal/websocket"
	"gorm.io/gorm"
)

// Services is the application service registry (facades + shared runtime deps).
// Wired via application/bootstrap.NewServices.
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
	Push         PushServiceInterface
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
	Ai           AiServiceInterface
}

// JobHandlers wires service methods into background task handlers.
func (s *Services) JobHandlers() asynq.Handlers {
	return asynq.Handlers{
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

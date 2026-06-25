// Package handlers contains HTTP handlers for incoming requests. Each handler corresponds to a domain (e.g., Auth, Product, Order), validates input, calls services, and returns HTTP responses. Container aggregates all handlers for route wiring.
package handlers

import (
	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/health"
	"github.com/alireza-akbarzadeh/luxe/internal/services"
	"gorm.io/gorm"
)

type Container struct {
	Health   *HealthHandler
	Auth     *AuthHandler
	User     *UserHandler
	Page     *PageHandler
	Cart     *CartHandler
	Product  *ProductHandler
	Category *CategoryHandler
	Order    *OrderHandler
	Shipment *ShipmentHandler
	Coupon   *CouponHandler
	Address  *AddressHandler
	Menu     *MenuHandler
	Review   *ReviewHandler
	UserLike *UserLikeHandler
	Account  *AccountHandler
	Wallet   *WalletHandler
	Payment  *PaymentProviderHandler
	Store    *StoreHandler
	Search   *SearchHandler
	Compare  *CompareHandler
	NavMenu  *NavMenuHandler
	Brand      *BrandHandler
	Collection *CollectionHandler
	Settings   *SettingHandler
	Pdp      *PdpHandler
	WebSocket *WebSocketHandler
	Stripe    *StripeWebhookHandler
	Audit     *AuditHandler
	Upload    *UploadHandler
	Admin     *AdminHandler
	Import    *ImportHandler
	Workflow  *WorkflowHandler
	Return    *ReturnHandler
	Invoice   *InvoiceHandler
	Role      *RoleHandler
	Inventory *InventoryHandler
	Push      *PushHandler
	Ai        *AiHandler
}

// NewContainer initializes all handlers with their dependencies.
func NewContainer(db *gorm.DB, svc *services.Services, cfg *config.Config) *Container {
	return &Container{
		Health:   NewHealthHandler(health.NewChecker(db, cfg)),
		Search:   NewSearchHandler(svc.Search),
		Auth:     NewAuthHandler(svc.Auth),
		User:     NewUserHandler(svc.User, svc.Address),
		Cart:     NewCartHandler(svc.Cart),
		Product:  NewProductHandler(svc.Product, svc.UserLike, svc.Pdp),
		Pdp:      NewPdpHandler(svc.Pdp, svc.Product),
		Compare:  NewCompareHandler(svc.Compare),
		Category: NewCategoryHandler(svc.Category),
		Order:    NewOrderHandler(svc.Order, svc.Checkout),
		Shipment: NewShipmentHandler(svc.Shipment),
		Page:     NewPageHandler(),
		Store:    NewStoreHandler(svc.Store, svc.Product),
		Account:  NewAccountHandler(svc.Address, svc.UserLike, svc.Order, svc.User),
		Coupon:   NewCouponHandler(svc.Coupon),
		Address:  NewAddressHandler(svc.Address),
		Menu:     NewMenuHandler(svc.Menu),
		Review:   NewReviewHandler(svc.Review),
		UserLike: NewUserLikeHandler(svc.UserLike, svc.Product),
		Wallet:   NewWalletHandler(svc.Wallet),
		Payment:  NewPaymentMethodHandler(svc.Payment, cfg),
		NavMenu:  NewNavMenuHandler(svc.NavMenu),
		Brand:      NewBrandHandler(svc.Brand),
		Collection: NewCollectionHandler(svc.Collection),
		Settings:   NewSettingHandler(svc.Settings),
		WebSocket: NewWebSocketHandler(svc),
		Stripe:    NewStripeWebhookHandler(svc.Payment, svc.Checkout, svc.Wallet, svc.WebhookEvent, cfg),
		Audit:     NewAuditHandler(svc.Audit),
		Upload:    NewUploadHandler(svc.Upload),
		Admin:     NewAdminHandler(svc.Admin, svc.Order, svc.WebhookEvent),
		Import:    NewImportHandler(svc.Import),
		Workflow:  NewWorkflowHandler(svc.Workflow),
		Return:    NewReturnHandler(svc.Return),
		Invoice:   NewInvoiceHandler(svc.Invoice),
		Role:      NewRoleHandler(svc.Role),
		Inventory: NewInventoryHandler(svc.Inventory),
		Push:      NewPushHandler(svc.Push),
		Ai:        NewAiHandler(svc.Ai),
	}
}

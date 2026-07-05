// Package handlers contains HTTP handlers for incoming requests. Each handler corresponds to a domain (e.g., Auth, Product, Order), validates input, calls application use cases, and returns HTTP responses. Container aggregates all handlers for route wiring.
package handlers

import (
	"github.com/alireza-akbarzadeh/luxe/internal/application/bootstrap"
	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/health"
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
	GiftCard  *GiftCardHandler
	Plus      *PlusHandler
	Home      *HomeHandler
	ShopLook  *ShopLookHandler
	Bundle    *BundleHandler
}

// NewContainer initializes all handlers with their dependencies.
func NewContainer(db *gorm.DB, apps *bootstrap.Applications, runtime *bootstrap.Runtime, cfg *config.Config) *Container {
	return &Container{
		Health:   NewHealthHandler(health.NewChecker(db, cfg)),
		Search:   NewSearchHandler(apps.Search.Commands, apps.Search.Queries),
		Auth:     NewAuthHandler(apps.Auth),
		User:     NewUserHandler(apps.User.Commands, apps.User.Queries, apps.Address.Commands, apps.Address.Queries),
		Cart:     NewCartHandler(apps.Cart.Commands, apps.Cart.Queries),
		Product:  NewProductHandler(apps.Product, apps.UserLike.Commands, apps.UserLike.Queries, apps.Pdp),
		Pdp:      NewPdpHandler(apps.Pdp, apps.Product),
		Compare:  NewCompareHandler(apps.Compare.Commands, apps.Compare.Queries),
		Category: NewCategoryHandler(apps.Category),
		Order:    NewOrderHandler(apps.Order, apps.Checkout, apps.Store.Queries),
		Shipment: NewShipmentHandler(apps.Shipment),
		Page:     NewPageHandler(),
		Store:    NewStoreHandler(apps.Store.Commands, apps.Store.Queries, apps.Product),
		Account:  NewAccountHandler(apps.Address.Commands, apps.Address.Queries, apps.UserLike.Commands, apps.UserLike.Queries, apps.Order, apps.User.Queries),
		Coupon:   NewCouponHandler(apps.Coupon),
		Address:  NewAddressHandler(apps.Address.Commands, apps.Address.Queries),
		Menu:     NewMenuHandler(apps.Menu.Commands, apps.Menu.Queries),
		Review:   NewReviewHandler(apps.Review.Commands, apps.Review.Queries),
		UserLike: NewUserLikeHandler(apps.UserLike.Commands, apps.UserLike.Queries, apps.Product),
		Wallet:   NewWalletHandler(apps.Wallet),
		Payment:  NewPaymentMethodHandler(apps.Payment, cfg),
		NavMenu:  NewNavMenuHandler(apps.NavMenu.Commands, apps.NavMenu.Queries),
		Brand:      NewBrandHandler(apps.Brand),
		Collection: NewCollectionHandler(apps.Collection),
		Settings:   NewSettingHandler(apps.Settings.Commands, apps.Settings.Queries),
		WebSocket: NewWebSocketHandler(runtime.WebSocketHub, apps.Notification),
		Stripe:    NewStripeWebhookHandler(apps.Payment, apps.Checkout, apps.Wallet, apps.Membership, apps.GiftCard, apps.Webhook.Commands, cfg),
		Audit:     NewAuditHandler(apps.Audit.Queries),
		Upload:    NewUploadHandler(apps.Upload),
		Admin:     NewAdminHandler(apps.Admin, apps.Order, apps.Webhook.Queries),
		Import:    NewImportHandler(apps.Import),
		Workflow:  NewWorkflowHandler(apps.Workflow),
		Return:    NewReturnHandler(apps.Return.Commands, apps.Return.Queries),
		Invoice:   NewInvoiceHandler(apps.Invoice.Commands, apps.Invoice.Queries),
		Role:      NewRoleHandler(apps.Role.Commands, apps.Role.Queries),
		Inventory: NewInventoryHandler(apps.Inventory),
		Push:      NewPushHandler(apps.Push),
		Ai:        NewAiHandler(apps.AI, apps.Search.Queries, apps.Compare.Queries, apps.Review.Queries, apps.Return.Queries, pdpPriceHistoryAdapter{svc: apps.Pdp}, shipmentDeliveryStatsAdapter{repo: postgres.NewShipmentRepository(db)}, apps.UserLike.Queries, shoppingMemoryAdapter{home: postgres.NewHomeRepository(db), wishlist: apps.UserLike.Queries}, replenishmentAdapter{orders: postgres.NewOrderRepository(db)}),
		GiftCard:  NewGiftCardHandler(apps.GiftCard),
		Plus:      NewPlusHandler(apps.Membership),
		Home:      NewHomeHandler(apps.Home),
		ShopLook:  NewShopLookHandler(apps.ShopLook),
		Bundle:    NewBundleHandler(apps.Bundle),
	}
}

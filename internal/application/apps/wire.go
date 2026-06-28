// Package apps wires application-layer use cases without importing service orchestrators.
package apps

import (
	"context"
	"strings"

	appaddress "github.com/alireza-akbarzadeh/luxe/internal/application/address"
	appbrand "github.com/alireza-akbarzadeh/luxe/internal/application/brand"
	appcategory "github.com/alireza-akbarzadeh/luxe/internal/application/category"
	appcollection "github.com/alireza-akbarzadeh/luxe/internal/application/collection"
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
	appadmin "github.com/alireza-akbarzadeh/luxe/internal/application/admin"
	appai "github.com/alireza-akbarzadeh/luxe/internal/application/ai"
	appaudit "github.com/alireza-akbarzadeh/luxe/internal/application/audit"
	appauth "github.com/alireza-akbarzadeh/luxe/internal/application/auth"
	appcart "github.com/alireza-akbarzadeh/luxe/internal/application/cart"
	appcompare "github.com/alireza-akbarzadeh/luxe/internal/application/compare"
	appcoupon "github.com/alireza-akbarzadeh/luxe/internal/application/coupon"
	appgiftcard "github.com/alireza-akbarzadeh/luxe/internal/application/giftcard"
	apphome "github.com/alireza-akbarzadeh/luxe/internal/application/home"
	appinvoice "github.com/alireza-akbarzadeh/luxe/internal/application/invoice"
	appmenu "github.com/alireza-akbarzadeh/luxe/internal/application/menu"
	appmembership "github.com/alireza-akbarzadeh/luxe/internal/application/membership"
	appnavmenu "github.com/alireza-akbarzadeh/luxe/internal/application/navmenu"
	apppayment "github.com/alireza-akbarzadeh/luxe/internal/application/payment"
	appreturn "github.com/alireza-akbarzadeh/luxe/internal/application/returnorder"
	approle "github.com/alireza-akbarzadeh/luxe/internal/application/role"
	appreview "github.com/alireza-akbarzadeh/luxe/internal/application/review"
	appsearch "github.com/alireza-akbarzadeh/luxe/internal/application/search"
	appsettings "github.com/alireza-akbarzadeh/luxe/internal/application/settings"
	appstore "github.com/alireza-akbarzadeh/luxe/internal/application/store"
	appupload "github.com/alireza-akbarzadeh/luxe/internal/application/upload"
	appuser "github.com/alireza-akbarzadeh/luxe/internal/application/user"
	appuserlike "github.com/alireza-akbarzadeh/luxe/internal/application/userlike"
	appwallet "github.com/alireza-akbarzadeh/luxe/internal/application/wallet"
	appwebhook "github.com/alireza-akbarzadeh/luxe/internal/application/webhookevent"
	appworkflow "github.com/alireza-akbarzadeh/luxe/internal/application/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/asynq"
	stripeintegration "github.com/alireza-akbarzadeh/luxe/internal/infrastructure/integrations/stripe"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

type auditApp struct {
	Commands *appaudit.Commands
	Queries  *appaudit.Queries
}

type searchApp struct {
	Commands *appsearch.Commands
	Queries  *appsearch.Queries
}

type addressApp struct {
	Commands *appaddress.Commands
	Queries  *appaddress.Queries
}

type cartApp struct {
	Commands *appcart.Commands
	Queries  *appcart.Queries
}

type compareApp struct {
	Commands *appcompare.Commands
	Queries  *appcompare.Queries
}

type userApp struct {
	Commands *appuser.Commands
	Queries  *appuser.Queries
}

type navMenuApp struct {
	Commands *appnavmenu.Commands
	Queries  *appnavmenu.Queries
}

type menuApp struct {
	Commands *appmenu.Commands
	Queries  *appmenu.Queries
}

type reviewApp struct {
	Commands *appreview.Commands
	Queries  *appreview.Queries
}

type returnApp struct {
	Commands *appreturn.Commands
	Queries  *appreturn.Queries
}

type roleApp struct {
	Commands *approle.Commands
	Queries  *approle.Queries
}

type invoiceApp struct {
	Commands *appinvoice.Commands
	Queries  *appinvoice.Queries
}

type settingsApp struct {
	Commands *appsettings.Commands
	Queries  *appsettings.Queries
}

type storeApp struct {
	Commands *appstore.Commands
	Queries  *appstore.Queries
}

type userLikeApp struct {
	Commands *appuserlike.Commands
	Queries  *appuserlike.Queries
}

type webhookApp struct {
	Commands *appwebhook.Commands
	Queries  *appwebhook.Queries
}

// Applications holds wired application-layer use cases (handlers call these directly).
type Applications struct {
	AI       *appai.Service
	Auth     *appauth.Service
	Audit    auditApp
	Search   searchApp
	Address  addressApp
	Cart     cartApp
	Compare  compareApp
	User     userApp
	NavMenu  navMenuApp
	Menu     menuApp
	Review   reviewApp
	UserLike userLikeApp
	Store    storeApp
	Upload   *appupload.Service
	Webhook  webhookApp
	Workflow *appworkflow.Module
	Return   returnApp
	Role     roleApp
	Invoice  invoiceApp
	Settings settingsApp
	Coupon   *appcoupon.Service
	Payment  *apppayment.Service
	Wallet   *appwallet.Service
	Brand    *appbrand.Service
	Category *appcategory.Service
	Collection *appcollection.Service
	Product      *appcatalog.Service
	Pdp          *apppdp.Service
	Inventory    *appinventory.Service
	Admin        *appadmin.Service
	Import       *importdata.Service
	Push         *apppush.WebPushService
	Notification *appnotification.Service
	Checkout     *appcheckout.Service
	Order        *orderfacade.Service
	Shipment     *appshipment.Service
	SalesFeed    *appsalesfeed.Service
	GiftCard     *appgiftcard.Service
	Membership   *appmembership.Service
	Home         *apphome.Service
}

type legalSettingReader struct {
	queries *appsettings.Queries
}

func (r legalSettingReader) GetSettingJSON(ctx context.Context, key string) ([]byte, error) {
	setting, err := r.queries.Get(ctx, key)
	if err != nil || setting == nil {
		return nil, err
	}
	return setting.Value, nil
}

// WireApplications constructs all application use cases from infrastructure dependencies.
func WireApplications(
	db *gorm.DB,
	cfg *config.Config,
	jobQueue asynq.JobQueue,
	engine *workflow.Engine,
) *Applications {
	stripeEnabled := cfg != nil && cfg.Stripe.Enabled
	var stripeGateway *stripeintegration.Gateway
	if stripeEnabled && cfg != nil {
		stripeGateway = stripeintegration.NewGateway(cfg.Stripe.SecretKey, cfg.Email.FrontendURL)
	}

	settingsRepo := postgres.NewSettingRepository(db)
	settingsQueries := appsettings.NewQueries(settingsRepo)
	settingsCommands := appsettings.NewCommands(settingsRepo)

	auditRepo := postgres.NewAuditRepository(db)
	searchRepo := postgres.NewSearchRepository(db)
	addressRepo := postgres.NewAddressRepository(db)
	cartRepo := postgres.NewCartRepository(db)
	compareRepo := postgres.NewCompareRepository(db)
	userRepo := postgres.NewUserRepository(db)
	navMenuRepo := postgres.NewNavMenuRepository(db)
	menuRepo := postgres.NewMenuRepository(db)
	reviewRepo := postgres.NewReviewRepository(db)
	userLikeRepo := postgres.NewUserLikeRepository(db)
	storeRepo := postgres.NewStoreRepository(db)
	storeQueries := appstore.NewQueries(storeRepo)
	webhookRepo := postgres.NewWebhookEventRepository(db)
	returnRepo := postgres.NewReturnRepository(db)
	roleRepo := postgres.NewRoleRepository(db)
	roleQueries := approle.NewQueries(roleRepo)
	invoiceRepo := postgres.NewInvoiceRepository(db)

	frontendURL := ""
	if cfg != nil {
		frontendURL = strings.TrimRight(cfg.Email.FrontendURL, "/")
	}

	walletSvc := appwallet.NewService(postgres.NewWalletRepository(db), stripeGateway, stripeEnabled)
	giftCardSvc := appgiftcard.NewService(postgres.NewGiftCardRepository(db))
	membershipSvc := appmembership.NewService(userRepo, walletSvc, giftCardSvc, stripeGateway, stripeEnabled)

	return &Applications{
		AI:   appai.NewService(postgres.NewProductRepository(db), cfg.AI),
		Auth: appauth.NewService(postgres.NewAuthRepository(db), cfg, jobQueue, engine, legalSettingReader{queries: settingsQueries}),
		Audit: auditApp{
			Commands: appaudit.NewCommands(auditRepo),
			Queries:  appaudit.NewQueries(auditRepo),
		},
		Search: searchApp{
			Commands: appsearch.NewCommands(searchRepo),
			Queries:  appsearch.NewQueries(searchRepo),
		},
		Address: addressApp{
			Commands: appaddress.NewCommands(addressRepo),
			Queries:  appaddress.NewQueries(addressRepo),
		},
		Cart: cartApp{
			Commands: appcart.NewCommands(cartRepo, cartRepo),
			Queries:  appcart.NewQueries(cartRepo),
		},
		Compare: compareApp{
			Commands: appcompare.NewCommands(compareRepo),
			Queries:  appcompare.NewQueries(compareRepo),
		},
		User: userApp{
			Commands: appuser.NewCommands(userRepo),
			Queries:  appuser.NewQueries(userRepo),
		},
		NavMenu: navMenuApp{
			Commands: appnavmenu.NewCommands(navMenuRepo),
			Queries:  appnavmenu.NewQueries(navMenuRepo),
		},
		Menu: menuApp{
			Commands: appmenu.NewCommands(menuRepo),
			Queries:  appmenu.NewQueries(menuRepo),
		},
		Review: reviewApp{
			Commands: appreview.NewCommands(reviewRepo, engine),
			Queries:  appreview.NewQueries(reviewRepo),
		},
		UserLike: userLikeApp{
			Commands: appuserlike.NewCommands(userLikeRepo),
			Queries:  appuserlike.NewQueries(userLikeRepo),
		},
		Store: storeApp{
			Commands: appstore.NewCommands(storeRepo, storeQueries),
			Queries:  storeQueries,
		},
		Upload: appupload.NewService(cfg),
		Webhook: webhookApp{
			Commands: appwebhook.NewCommands(webhookRepo),
			Queries:  appwebhook.NewQueries(webhookRepo),
		},
		Workflow: appworkflow.NewModule(db, engine),
		Return: returnApp{
			Commands: appreturn.NewCommands(returnRepo, engine, membershipSvc),
			Queries:  appreturn.NewQueries(returnRepo),
		},
		Role: roleApp{
			Commands: approle.NewCommands(roleRepo, roleQueries),
			Queries:  roleQueries,
		},
		Invoice: invoiceApp{
			Commands: appinvoice.NewCommands(invoiceRepo, jobQueue, frontendURL),
			Queries:  appinvoice.NewQueries(invoiceRepo),
		},
		Settings: settingsApp{
			Commands: settingsCommands,
			Queries:  settingsQueries,
		},
		Coupon:  appcoupon.NewService(postgres.NewCouponRepository(db), engine),
		Payment: apppayment.NewService(postgres.NewPaymentRepository(db), stripeGateway, stripeEnabled),
		Wallet:  walletSvc,
		Brand:       appbrand.NewService(db, engine),
		Category:    appcategory.NewService(db, engine),
		Collection:  appcollection.NewService(db, engine),
		GiftCard:    giftCardSvc,
		Membership:  membershipSvc,
		Home: apphome.NewService(
			postgres.NewStorefrontRepository(db),
			postgres.NewHomeRepository(db),
		),
	}
}

// Log implements middleware.AuditLogger.
func (a *Applications) Log(ctx context.Context, entry *models.AuditLog) error {
	return a.Audit.Commands.Log(ctx, entry)
}

// LogAudit is an alias for Log (audit middleware compatibility).
func (a *Applications) LogAudit(ctx context.Context, entry *models.AuditLog) error {
	return a.Log(ctx, entry)
}

// HasPermission implements role checks for route middleware.
func (a *Applications) HasPermission(ctx context.Context, roleSlug, permissionKey string) (bool, error) {
	return a.Role.Queries.HasPermission(ctx, roleSlug, permissionKey)
}

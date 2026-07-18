// Package constants holds shared immutable configuration values, error messages,
// and HTTP status codes used across the shopping platform.
package constants

import (
	"errors"
	"time"
)


type ContextKey string

// ==================== Defaults ====================
const (
	DefaultProtectedAPIPort       int    = 8080
	DefaultPublicAPIPort          int    = 8081
	DefaultHiddenAPIPort          int    = 8079
	DefaultHost                   string = "0.0.0.0"
	DefaultDevHost                string = "127.0.0.1"
	DefaultLogLevel               string = "warn"
	DefaultDevLogLevel            string = "debug"
	DefaultCORSAllowOrigins       string = "*"
	DefaultDBPlatform             string = DBPlatformSQLite
	DefaultDBTimezone             string = DBTimezoneUTC
	DefaultDBSSLMode              string = DBSSLModeDisabled
	DefaultSQLiteDBName           string = "sqlite.db"
	DefaultLoggerTimestampFormat  string = "2006-01-02 15:04:05.00000"
	DefaultRequestTimeoutDuration        = 60 * time.Second
	DefaultWatcherSleepInterval          = 5 * time.Second
	DefaultGzipLevel              int    = 5
	DefaultLimit                  int    = 20
	MaxLimit                      int    = 100
	MinLimit                      int    = 1
	MinOffset                     int    = 0
)

// ==================== Feature flags ====================
const (
	FeatureService   = "service"
	FeatureOryKratos = "ory_kratos"
	FeatureOryKeto   = "ory_keto"
	FeatureDatabase  = "database"
	FeatureCORS      = "cors"
	FeatureGzip      = "gzip"
	FeatureRedis     = "redis"
)

// ==================== Generic words ====================
const (
	WordDatabase       = "database"
	WordDatabaseServer = "database_server"
	WordInternalCode   = "internalCode"
	WordServiceCode    = "serviceCode"
)

// ==================== App‑specific names ====================
const (
	NameHealthPath      = "/alive"
	NameHealthReadyPath = "/ready"
	NameCORSConfig      = "CORSAllowOrigins"
	NameTimeoutDuration = "RequestTimeoutDuration"
	NameCmdDBMigrate    = "migrate"
	NameCmdDBRollback   = "rollback"
	NameCmdDBSeed       = "seed"
)

// ==================== Database platforms & settings ====================
const (
	DBPlatformPostgres  = "postgres"
	DBPlatformMySQL     = "mysql"
	DBPlatformSQLite    = "sqlite"
	DBSSLModeEnabled    = "require"
	DBSSLModeDisabled   = "disable"
	DBTimezoneUTC       = "Etc/GMT"
	DBTimezoneMelbourne = "Australia/Melbourne"
)

// ==================== Headers & context keys ====================
const (
	HeaderContentType     = "Content-Type"
	HeaderContentTypeJSON = "application/json; charset=utf-8"
	HeaderAuthorization   = "Authorization"
	HeaderAuthBearerWord  = "Bearer"
	HeaderKratosCookie    = "ory_kratos_session"
)

var (
	RequestIDKey  ContextKey = "request_id"
	UserIDKey     ContextKey = "uid"
	LocaleContextKey ContextKey = "locale"
)

// ==================== Output messages ====================
const (
	MsgServerShuttingDown       = "server is shutting down"
	MsgNotAcceptable            = "not acceptable"
	MsgMissingAcceptHeader      = "unknown accept format"
	MsgMissingContentTypeHeader = "unknown content format"
	MsgSuccess                  = "success"
	MsgError                    = "error"
	MsgValidationError          = "validation error"
	MsgRouteNotFound            = "route not found"
	MsgRecordNotFound           = "record not found"
	MsgDependencyNotFound       = "dependency not found"
	MsgSessionNotFound          = "session not found"
	MsgAccessIDsNotFound        = "access ids not found or not readable"
	MsgNotAuthorized            = "not authorized"
	MsgIDNotReadable            = "ID not found or not readable"
	MsgUnableToBindBody         = "error binding body"
	MsgForbidden                = "forbidden"
	MsgUnknownDBPlatform        = "unknown database platform"
	MsgInternalServer           = "internal server error"
	MsgShutdownServerCompleted  = "Graceful shutdown complete."
	MsgUserRegisterSuccess      = "User registered successfully"
	ErrShipmentNotFound         = "shipment not found"
)

// ==================== Log levels ====================
var (
	LogLevels = []string{"debug", "info", "warn", "error", "fatal", "panic"}
)

// ==================== Predefined errors ====================
var (
	ErrNotAuthorized     = errors.New(MsgNotAuthorized)
	ErrSessionNotFound   = errors.New(MsgSessionNotFound)
	ErrIDNotFound        = errors.New(MsgIDNotReadable)
	ErrAccessIDsNotFound = errors.New(MsgAccessIDsNotFound)
	ErrBindingBody       = errors.New(MsgUnableToBindBody)
	ErrUnknownDBPlatform = errors.New(MsgUnknownDBPlatform)
	ErrInternalServer    = errors.New(MsgInternalServer)
)

const (
	Day   = 24 * time.Hour
	Week  = 7 * Day
	Month = 30 * Day
	Year  = 365 * Day
)

// ==================== Cron job schedules ====================
const (
	// Cart jobs
	CronAbandonedCartCleanup = "@every 30m"

	// Order jobs
	CronUpdateOverdueOrders = "0 2 * * *" // Daily at 2 AM UTC

	// Product jobs
	CronLowStockAlert     = "0 9 * * *" // Daily at 9 AM UTC
	CronSyncProductPrices = "0 1 * * *" // Daily at 1 AM UTC

	// Common schedules for future use
	CronEvery5Minutes  = "@every 5m"   // Every 5 minutes
	CronEvery15Minutes = "@every 15m"  // Every 15 minutes
	CronEvery30Minutes = "@every 30m"  // Every 30 minutes
	CronEveryHour      = "@every 1h"   // Every hour
	CronEvery2Hours    = "@every 2h"   // Every 2 hours
	CronEvery6Hours    = "@every 6h"   // Every 6 hours
	CronEvery12Hours   = "@every 12h"  // Every 12 hours
	CronDailyMidnight  = "0 0 * * *"   // Daily at midnight UTC
	CronDaily6AM       = "0 6 * * *"   // Daily at 6 AM UTC
	CronDailyNoon      = "0 12 * * *"  // Daily at noon UTC
	CronDaily6PM       = "0 18 * * *"  // Daily at 6 PM UTC
	CronWeeklySunday   = "0 0 * * 0"   // Weekly on Sunday at midnight
	CronWeeklyMonday   = "0 0 * * 1"   // Weekly on Monday at midnight
	CronMonthlyFirst   = "0 0 1 * *"   // Monthly on the 1st at midnight
	CronQuarterly      = "0 0 1 */3 *" // Quarterly on the 1st of Jan, Apr, Jul, Oct
)

// ==================== Domain statuses ====================
const (
	// Cart statuses
	CartStatusActive    = "active"
	CartStatusAbandoned = "abandoned"
	CartStatusConverted = "converted"

	// Order statuses
	OrderStatusPending   = "pending"
	OrderStatusPaid      = "paid"
	OrderStatusShipped   = "shipped"
	OrderStatusDelivered = "delivered"
	OrderStatusCancelled = "cancelled"
	OrderStatusRefunded  = "refunded"
	OrderStatusDelayed   = "delayed"

	// Payment statuses
	PaymentStatusPending   = "pending"
	PaymentStatusCompleted = "completed"
	PaymentStatusSucceeded = "succeeded"
	PaymentStatusFailed    = "failed"
	PaymentStatusRefunded  = "refunded"

	// Wallet transaction types
	WalletTxTypeDeposit     = "deposit"
	WalletTxTypePayment     = "payment"
	WalletTxTypeRefund      = "refund"
	WalletTxTypeAdjustment  = "adjustment"
	WalletTxTypeMembership  = "membership"

	// Wallet transaction statuses
	WalletTxStatusPending   = "pending"
	WalletTxStatusCompleted = "completed"
	WalletTxStatusFailed    = "failed"
	WalletTxStatusCancelled = "cancelled"

	// Wallet reference types
	WalletRefTypeOrder = "order"
	WalletRefTypeAdmin = "admin"

	// Gift card statuses
	GiftCardStatusPending   = "pending"
	GiftCardStatusActive    = "active"
	GiftCardStatusRedeemed  = "redeemed"
	GiftCardStatusExpired   = "expired"
	GiftCardStatusCancelled = "cancelled"

	// Reverse marketplace request statuses
	ReverseMarketplaceRequestStatusOpen      = "open"
	ReverseMarketplaceRequestStatusClosed    = "closed"
	ReverseMarketplaceRequestStatusFulfilled = "fulfilled"

	// Reverse marketplace offer statuses
	ReverseMarketplaceOfferStatusPending   = "pending"
	ReverseMarketplaceOfferStatusAccepted  = "accepted"
	ReverseMarketplaceOfferStatusDeclined  = "declined"
	ReverseMarketplaceOfferStatusWithdrawn = "withdrawn"

	// Shipment statuses
	ShipmentStatusPending   = "pending"
	ShipmentStatusShipped   = "shipped"
	ShipmentStatusDelivered = "delivered"

	// Invoice statuses
	InvoiceStatusDraft    = "draft"
	InvoiceStatusIssued   = "issued"
	InvoiceStatusPaid     = "paid"
	InvoiceStatusVoid     = "void"
	InvoiceStatusRefunded = "refunded"

	// User roles
	RoleUser      = "user"
	RoleAdmin     = "admin"
	RoleModerator = "moderator"

	// Membership tiers (Luxe Plus)
	MembershipTierFree = "free"
	MembershipTierPlus = "plus"

	// Customer CRM segments (admin-assigned)
	CustomerSegmentVIP    = "vip"
	CustomerSegmentLoyal  = "loyal"
	CustomerSegmentNew    = "new"
	CustomerSegmentAtRisk = "at_risk"

	// Support ticket statuses
	SupportTicketStatusOpen            = "open"
	SupportTicketStatusPending         = "pending"
	SupportTicketStatusWaitingCustomer = "waiting_customer"
	SupportTicketStatusResolved        = "resolved"
	SupportTicketStatusClosed          = "closed"

	// Support channels
	SupportChannelEmail = "email"
	SupportChannelChat  = "chat"
	SupportChannelWeb   = "web"

	// Support priorities
	SupportPriorityLow    = "low"
	SupportPriorityNormal = "normal"
	SupportPriorityHigh   = "high"
	SupportPriorityUrgent = "urgent"

	// Support message author roles
	SupportAuthorCustomer = "customer"
	SupportAuthorStaff    = "staff"
	SupportAuthorSystem   = "system"

	// Coupon application types
	CouponApplicationCode      = "code"
	CouponApplicationAutomatic = "automatic"
	CouponApplicationBOGO      = "bogo"

	// Luxe Plus benefit values (enforced in checkout, returns, shipping)
	PlusAnnualPriceUSD        = 49.99
	PlusCheckoutDiscountPct   = 10
	PlusReturnWindowDays      = 60
	FreeReturnWindowDays      = 30
	PlusSupportPriority       = "priority"
	FreeSupportPriority       = "standard"
	PlusShippingPriorityLabel = "express"

	// Luxe Plus subscription payment methods
	PlusPaymentWallet        = "wallet"
	PlusPaymentGiftCard      = "gift_card"
	PlusPaymentStripe        = "stripe"
	PlusStripeMetadataType   = "plus_membership"
	PlusPaymentStatusCompleted = "completed"
	PlusPaymentStatusPending   = "pending"

	// Product statuses
	ProductStatusDraft    = "draft"
	ProductStatusActive   = "active"
	ProductStatusInactive = "inactive"
	ProductStatusArchived = "archived"

	// Store statuses
	StoreStatusActive    = "active"
	StoreStatusPending   = "pending"
	StoreStatusSuspended = "suspended"

	// Chat / stock notification statuses
	ChatRoomStatusActive          = "active"
	StockNotificationStatusActive = "active"

	// In-app notification types (customer)
	NotificationTypeMembershipActivated = "membership_activated"
	NotificationTypeWalletDeposit       = "wallet_deposit"
	NotificationTypeGiftCardSent          = "gift_card_sent"
	NotificationTypeGiftCardReceived    = "gift_card_received"
	NotificationTypeGiftCardTransferred = "gift_card_transferred"

	// Webhook event statuses
	WebhookStatusReceived  = "received"
	WebhookStatusProcessed = "processed"
	WebhookStatusFailed    = "failed"

	// Workflow entity types (also the workflow keys for the seeded workflows)
	WorkflowEntityOrder    = "order"
	WorkflowEntityProduct  = "product"
	WorkflowEntityShipment = "shipment"
	WorkflowEntityReturn   = "return"
	WorkflowEntityUser     = "user"
	WorkflowEntityCategory = "category"
	WorkflowEntityBrand      = "brand"
	WorkflowEntityCollection = "collection"
	WorkflowEntityCoupon     = "coupon"
	WorkflowEntityReview     = "review"
	WorkflowEntityBlogPost    = "blog_post"
	WorkflowEntityPrivacyRule = "privacy_rule"

	// Privacy rule statuses
	PrivacyRuleStatusDraft    = "draft"
	PrivacyRuleStatusActive   = "active"
	PrivacyRuleStatusInactive = "inactive"
	PrivacyRuleStatusArchived = "archived"

	// Privacy rule providers (which integration / surface the rule applies to)
	PrivacyRuleProviderPlatform = "platform"
	PrivacyRuleProviderStripe   = "stripe"
	PrivacyRuleProviderPaypal   = "paypal"
	PrivacyRuleProviderWallet   = "wallet"
	PrivacyRuleProviderGiftCard = "gift_card"
	PrivacyRuleProviderShipping = "shipping"
	PrivacyRuleProviderAI       = "ai"
	PrivacyRuleProviderAll      = "all"
)

// Inventory adjustment types (ledger).
const (
	InventoryAdjSale         = "sale"
	InventoryAdjOrderCancel  = "order_cancel"
	InventoryAdjAdminSet     = "admin_set"
	InventoryAdjAdminDelta   = "admin_delta"
	InventoryAdjImport       = "import"
	InventoryAdjInitial      = "initial"
	InventoryAdjCorrection   = "correction"
	InventoryAdjDamage       = "damage"
	InventoryAdjReceive      = "receive"
	InventoryAdjReturnRestock = "return_restock"
)

// Inventory reference types (ledger).
const (
	InventoryRefOrder     = "order"
	InventoryRefOrderItem = "order_item"
	InventoryRefProduct   = "product"
	InventoryRefImport    = "import"
	InventoryRefReturn    = "return"
)

// Inventory list stock_status filter values.
const (
	InventoryStockAll        = "all"
	InventoryStockLow        = "low"
	InventoryStockOut        = "out"
	InventoryStockHealthy    = "healthy"
	InventoryStockNotTracked = "not_tracked"
)

// ==================== API routes ====================
const (
	// API versions
	APIVersionV1 = "/api/v1"

	// Route groups
	RouteAuth       = "/auth"
	RouteProducts   = "/products"
	RouteCategories = "/categories"
	RouteUsers      = "/users"
	RouteCart       = "/cart"
	RouteOrders     = "/orders"
	RouteHealth     = "/health"
	RouteStatic     = "/static"
	RouteSwagger    = "/swagger/*any"
	RouteRoot       = "/"

	// Auth endpoints
	RouteAuthRegister = "/register"
	RouteAuthLogin    = "/login"
	RouteAuthRefresh  = "/refresh"
	RouteAuthLogout   = "/logout"

	// Product endpoints
	RouteProductBulk = "/bulk"

	// Cart endpoints
	RouteCartItems = "/items"

	// Order endpoints
	RouteOrdersMy = "/my"
)

package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/observability"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func (r *Router) Setup() {
	r.RegisterMiddlewares()
	middleware.SetRolePermissionChecker(r.apps)

	r.engine.GET(constants.RouteRoot, r.handlerContainer.Page.LandingPage)
	r.engine.Static(constants.RouteStatic, "./views/static")

	r.engine.GET(constants.RouteSwagger, ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.engine.GET("/openapi", func(c *gin.Context) {
		spec, err := getOpenAPI3Spec()
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to convert spec", "details": err.Error()})
			return
		}
		c.Data(200, "application/json", spec)
	})

	r.engine.GET("/metrics", observability.MetricsHandler())
	r.engine.GET("/health/live", r.handlerContainer.Health.Live)
	r.engine.GET("/health/ready", r.handlerContainer.Health.Ready)

	v1 := r.engine.Group(constants.APIVersionV1)
	{
		v1.GET(constants.RouteHealth, r.handlerContainer.Health.Check)
		v1.GET("/health/live", r.handlerContainer.Health.Live)
		v1.GET("/health/ready", r.handlerContainer.Health.Ready)
		SetupStripeRoutes(v1, r.handlerContainer)

		// ✅ PUBLIC (guest + optional auth)
		public := v1.Group(constants.RouteRoot)
		public.Use(middleware.GuestAuthMiddleware(r.cfg))
		// 🔒 PROTECTED (must login)
		protected := v1.Group(constants.RouteRoot)
		protected.Use(middleware.AuthMiddleware(r.cfg))
		// required role protected

		SetupAuditRoutes(protected, r.handlerContainer)
		SetupAuthRoutes(public, protected, r.handlerContainer)
		SetupAddressRoutes(protected, r.handlerContainer)
		SetupSearchRoutes(public, r.handlerContainer)
		SetupBundleRoutes(public, r.handlerContainer)
		SetupAccountRoutes(protected, r.handlerContainer)
		SetupProductRoutes(public, protected, r.handlerContainer)
		SetupCompareRoutes(public, protected, r.handlerContainer)
		SetupNavMenuRoutes(public, protected, r.handlerContainer)
		SetupPaymentRoutes(public, protected, r.handlerContainer)
		SetupCategoryRoutes(public, protected, r.handlerContainer)
		SetupSettingRoutes(public, protected, r.handlerContainer)
		SetupStoreRoutes(public, protected, r.handlerContainer)
		SetupCouponRoutes(public, protected, r.handlerContainer)
		SetupReviewRoutes(public, protected, r.handlerContainer)
		SetupUserRoutes(protected, r.handlerContainer)
		SetupCartRoutes(public, protected, r.handlerContainer)
		SetupMenuRoutes(protected, r.handlerContainer)
		SetupOrderRoutes(protected, r.handlerContainer)
		SetupShipmentRoutes(public, protected, r.handlerContainer)
		SetupWalletRoutes(protected, r.handlerContainer)
		SetupBrandRoutes(public, protected, r.handlerContainer)
		SetupCollectionRoutes(public, protected, r.handlerContainer)
		SetupUploadRoutes(public, protected, r.handlerContainer)
		SetupAdminRoutes(protected, r.handlerContainer)
		SetupImportRoutes(protected, r.handlerContainer)
		SetupWorkflowRoutes(protected, r.handlerContainer)
		SetupReturnRoutes(protected, r.handlerContainer)
		SetupInvoiceRoutes(protected, r.handlerContainer)
		SetupGiftCardRoutes(protected, r.handlerContainer)
		SetupPlusRoutes(public, protected, r.handlerContainer)
		SetupWebSocketRoutes(v1, protected, r.handlerContainer, r.cfg)
		SetupPushRoutes(v1, protected, r.handlerContainer)
		SetupAiRoutes(public, protected, r.handlerContainer)
		SetupHomeRoutes(public, protected, r.handlerContainer)
		SetupShopLookRoutes(public, r.handlerContainer)
		SetupCreatorStorefrontRoutes(public, r.handlerContainer)
	}
}

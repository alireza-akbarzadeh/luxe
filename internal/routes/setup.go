package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/middleware"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func (r *Router) Setup() {
	r.RegisterMiddlewares()

	r.engine.GET(constants.RouteRoot, r.controllers.Page.LandingPage)
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

	v1 := r.engine.Group(constants.APIVersionV1)
	{
		v1.GET(constants.RouteHealth, r.controllers.Health.Check)

		// ✅ PUBLIC (guest + optional auth)
		public := v1.Group(constants.RouteRoot)
		public.Use(middleware.GuestAuthMiddleware(r.cfg))
		// 🔒 PROTECTED (must login)
		protected := v1.Group(constants.RouteRoot)
		protected.Use(middleware.AuthMiddleware(r.cfg))
		// required role protected

		SetupAuthRoutes(public, protected, r.controllers)
		SetupAddressRoutes(protected, r.controllers)
		SetupSearchRoutes(public, r.controllers)
		SetupAccountRoutes(protected, r.controllers)
		SetupProductRoutes(public, protected, r.controllers)
		SetupCompareRoutes(public, protected, r.controllers)
		SetupNavMenuRoutes(public, r.controllers)
		SetupPaymentRoutes(protected, r.controllers)
		SetupCategoryRoutes(public, protected, r.controllers)
		SetupSettingRoutes(public, protected, r.controllers)
		SetupStoreRoutes(public, protected, r.controllers)
		SetupCouponRoutes(public, protected, r.controllers)
		SetupReviewRoutes(public, protected, r.controllers)
		SetupUserRoutes(protected, r.controllers)
		SetupCartRoutes(public, protected, r.controllers)
		SetupMenuRoutes(protected, r.controllers)
		SetupOrderRoutes(protected, r.controllers)
		SetupShipmentRoutes(public, protected, r.controllers)
		SetupWalletRoutes(protected, r.controllers)
		SetupBrandRoutes(public, protected, r.controllers)
	}
}

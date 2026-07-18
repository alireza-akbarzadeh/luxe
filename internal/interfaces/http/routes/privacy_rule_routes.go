package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
)

// SetupPrivacyRuleRoutes registers public and admin privacy rule endpoints.
func SetupPrivacyRuleRoutes(public, protected *gin.RouterGroup, ctrl *handlers.Container) {
	// Public — apps fetch active markdown rules for client-side parsing
	public.GET("/privacy-rules", ctrl.PrivacyRule.ListActivePrivacyRules)
	public.GET("/privacy-rules/key/:key", ctrl.PrivacyRule.GetActivePrivacyRuleByKey)
	public.GET("/privacy-rules/provider/:provider", ctrl.PrivacyRule.ListActivePrivacyRulesByProvider)

	admin := protected.Group("/admin/privacy-rules")
	admin.Use(middleware.ModuleGuard("settings"))
	{
		admin.GET("", ctrl.PrivacyRule.AdminListPrivacyRules)
		admin.GET("/:id", ctrl.PrivacyRule.AdminGetPrivacyRule)
		admin.POST("", ctrl.PrivacyRule.AdminCreatePrivacyRule)
		admin.PUT("/:id", ctrl.PrivacyRule.AdminUpdatePrivacyRule)
		admin.DELETE("/:id", ctrl.PrivacyRule.AdminDeletePrivacyRule)
	}
}

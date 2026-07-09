package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
)

func SetupEmailMarketingRoutes(api *gin.RouterGroup, protected *gin.RouterGroup, ctrl *handlers.Container) {
	api.POST("/newsletters/subscribe", ctrl.EmailMarketing.Subscribe)
	api.POST("/newsletters/unsubscribe", ctrl.EmailMarketing.Unsubscribe)

	admin := protected.Group("/admin")
	admin.Use(middleware.ModuleGuard("settings"))
	{
		admin.GET("/email-marketing/kpis", ctrl.EmailMarketing.GetKPIs)

		admin.GET("/newsletter-subscribers", ctrl.EmailMarketing.ListSubscribers)
		admin.GET("/newsletter-subscribers/export", ctrl.EmailMarketing.ExportSubscribers)
		admin.DELETE("/newsletter-subscribers/:id", ctrl.EmailMarketing.DeleteSubscriber)

		admin.GET("/email-templates", ctrl.EmailMarketing.ListTemplates)
		admin.POST("/email-templates", ctrl.EmailMarketing.CreateTemplate)
		admin.GET("/email-templates/:id", ctrl.EmailMarketing.GetTemplate)
		admin.PUT("/email-templates/:id", ctrl.EmailMarketing.UpdateTemplate)
		admin.DELETE("/email-templates/:id", ctrl.EmailMarketing.DeleteTemplate)

		admin.GET("/email-campaigns", ctrl.EmailMarketing.ListCampaigns)
		admin.POST("/email-campaigns", ctrl.EmailMarketing.CreateCampaign)
		admin.GET("/email-campaigns/:id", ctrl.EmailMarketing.GetCampaign)
		admin.PUT("/email-campaigns/:id", ctrl.EmailMarketing.UpdateCampaign)
		admin.DELETE("/email-campaigns/:id", ctrl.EmailMarketing.DeleteCampaign)
		admin.POST("/email-campaigns/:id/schedule", ctrl.EmailMarketing.ScheduleCampaign)
		admin.POST("/email-campaigns/:id/send", ctrl.EmailMarketing.SendCampaign)
	}
}

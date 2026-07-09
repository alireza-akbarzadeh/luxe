package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
)

func SetupPromotionRoutes(protected *gin.RouterGroup, ctrl *handlers.Container) {
	admin := protected.Group("/admin")
	admin.Use(middleware.ModuleGuard("products"))
	{
		admin.GET("/promotions/kpis", ctrl.Promotion.GetKPIs)

		admin.GET("/flash-deals", ctrl.Promotion.ListFlashDeals)
		admin.POST("/flash-deals", ctrl.Promotion.CreateFlashDeal)
		admin.GET("/flash-deals/:id", ctrl.Promotion.GetFlashDeal)
		admin.PUT("/flash-deals/:id", ctrl.Promotion.UpdateFlashDeal)
		admin.DELETE("/flash-deals/:id", ctrl.Promotion.DeleteFlashDeal)

		admin.GET("/homepage-sections", ctrl.Promotion.ListHomepageSections)
		admin.POST("/homepage-sections", ctrl.Promotion.CreateHomepageSection)
		admin.GET("/homepage-sections/:id", ctrl.Promotion.GetHomepageSection)
		admin.PUT("/homepage-sections/:id", ctrl.Promotion.UpdateHomepageSection)
		admin.DELETE("/homepage-sections/:id", ctrl.Promotion.DeleteHomepageSection)

		admin.GET("/campaigns", ctrl.Promotion.ListCampaigns)
		admin.POST("/campaigns", ctrl.Promotion.CreateCampaign)
		admin.GET("/campaigns/:id", ctrl.Promotion.GetCampaign)
		admin.PUT("/campaigns/:id", ctrl.Promotion.UpdateCampaign)
		admin.DELETE("/campaigns/:id", ctrl.Promotion.DeleteCampaign)
	}
}

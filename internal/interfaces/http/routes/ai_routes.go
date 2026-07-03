package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
)

func SetupAiRoutes(public *gin.RouterGroup, protected *gin.RouterGroup, ctrl *handlers.Container) {
	public.POST("/ai/chat", ctrl.Ai.Chat)
	public.POST("/ai/product-brief", ctrl.Ai.ProductBrief)
	public.POST("/ai/shopping-assistant", ctrl.Ai.ShoppingAssistant)

	admin := protected.Group("/admin")
	admin.Use(middleware.RequireStaff())
	ai := admin.Group("")
	ai.Use(middleware.ModuleGuard("products"))
	{
		ai.GET("/ai/status", ctrl.Ai.GetStatus)
		ai.POST("/ai/generate", ctrl.Ai.Generate)
	}
}

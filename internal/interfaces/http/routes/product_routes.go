package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
)

func SetupProductRoutes(public, protected *gin.RouterGroup, ctrl *handlers.Container) {
	public.GET("/products", ctrl.Product.List)
	public.GET("/products/:id/price-history", ctrl.Pdp.GetPriceHistory)
	public.GET("/products/:id/stock-heatmap", ctrl.Pdp.GetStockHeatmap)
	public.GET("/products/:id/alternatives", ctrl.Pdp.GetAlternatives)
	public.GET("/products/:id/questions", ctrl.Pdp.GetQuestions)
	public.GET("/products/:id/related", ctrl.Product.GetRelated)
	public.GET("/products/:id", ctrl.Product.GetOne)

	protected.GET("/products/:id/stock-notifications", ctrl.Pdp.GetStockStatus)
	protected.POST("/products/:id/stock-notifications", ctrl.Pdp.SubscribeStock)
	protected.DELETE("/products/:id/stock-notifications", ctrl.Pdp.UnsubscribeStock)
	protected.POST("/products/:id/questions", ctrl.Pdp.CreateQuestion)
	protected.POST("/products/:id/questions/:questionId/answers", ctrl.Pdp.CreateAnswer)
	protected.POST("/products/:id/like", ctrl.UserLike.ToggleLike)
	protected.GET("/products/:id/liked", ctrl.UserLike.IsLikedByUser)
	protected.POST("/products/suggestions", ctrl.Product.GetProductSuggestions)

	admin := protected.Group("/products")
	admin.Use(middleware.ModuleGuard("products"))
	{
		admin.POST("/", ctrl.Product.Create)
		admin.POST("/bulk", ctrl.Product.BulkCreate)
		admin.GET("/:id/available-transitions", ctrl.Product.GetAvailableTransitions)
		admin.POST("/:id/transition", ctrl.Product.PerformTransition)
		admin.PUT("/:id", ctrl.Product.Update)
		admin.DELETE("/:id", ctrl.Product.Delete)
		admin.DELETE("/bulk", ctrl.Product.BulkDelete)
	}
}

package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
)

func SetupReviewRoutes(public, protected *gin.RouterGroup, ctrl *handlers.Container) {
	public.GET("/reviews", ctrl.Review.GetProductReviews)

	protected.GET("/reviews/me", ctrl.Review.GetMyProductReview)
	protected.POST("/reviews", ctrl.Review.Create)
	protected.PUT("/reviews/:id", ctrl.Review.Update)
	protected.DELETE("/reviews/:id", ctrl.Review.Delete)

	admin := protected.Group("/admin/reviews")
	admin.Use(middleware.ModuleGuard("products"))
	{
		admin.GET("", ctrl.Review.ListReviewsAdmin)
		admin.POST("/:id/transition", ctrl.Review.PerformReviewTransition)
	}
}

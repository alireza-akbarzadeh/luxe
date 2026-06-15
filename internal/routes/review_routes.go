package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/controllers"
	"github.com/gin-gonic/gin"
)

func SetupReviewRoutes(public, protected *gin.RouterGroup, ctrl *controllers.Container) {
	public.GET("/reviews", ctrl.Review.GetProductReviews)

	protected.GET("/reviews/me", ctrl.Review.GetMyProductReview)
	protected.POST("/reviews", ctrl.Review.Create)
	protected.PUT("/reviews/:id", ctrl.Review.Update)
	protected.DELETE("/reviews/:id", ctrl.Review.Delete)
}

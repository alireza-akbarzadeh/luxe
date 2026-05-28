package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/controllers"
	"github.com/gin-gonic/gin"
)

func SetupSearchRoutes(public *gin.RouterGroup, ctrl *controllers.Container) {
	search := public.Group("/search")
	{
		search.GET("", ctrl.Search.GlobalSearch)
		search.GET("/suggestions", ctrl.Search.Suggestions)
		search.GET("/trending", ctrl.Search.Trending)
	}
}

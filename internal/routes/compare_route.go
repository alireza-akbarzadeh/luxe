package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/controllers"
	"github.com/gin-gonic/gin"
)

func SetupCompareRoutes(public, protected *gin.RouterGroup, ctrl *controllers.Container) {
	compare := protected.Group("/compare")
	{
		compare.GET("", ctrl.Compare.GetCompareList)
		compare.PUT("", ctrl.Compare.SyncCompareList)
		compare.POST("", ctrl.Compare.CompareProducts)
	}
}

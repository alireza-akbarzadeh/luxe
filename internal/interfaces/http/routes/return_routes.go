package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
)

func SetupReturnRoutes(protected *gin.RouterGroup, ctrl *handlers.Container) {
	returns := protected.Group("/returns")
	{
		returns.POST("", ctrl.Return.CreateReturn)
		returns.GET("/my", ctrl.Return.GetMyReturns)
		returns.GET("/:id", ctrl.Return.GetReturn)
	}

	admin := protected.Group("/admin/returns")
	admin.Use(middleware.ModuleGuard("orders"))
	{
		admin.GET("", ctrl.Return.ListReturnsAdmin)
		admin.GET("/:id", ctrl.Return.GetReturnAdmin)
		admin.POST("/:id/transition", ctrl.Return.PerformReturnTransition)
	}
}

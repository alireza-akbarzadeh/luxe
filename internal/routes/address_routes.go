package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/controllers"
	"github.com/gin-gonic/gin"
)

func SetupAddressRoutes(protected *gin.RouterGroup, ctrl *controllers.Container) {
	// Register without a trailing slash — Gin's RedirectTrailingSlash 301 breaks browser CORS.
	protected.GET("/addresses", ctrl.Address.List)
	protected.POST("/addresses", ctrl.Address.Create)
	protected.PUT("/addresses/:id", ctrl.Address.Update)
	protected.DELETE("/addresses/:id", ctrl.Address.Delete)
	protected.PATCH("/addresses/:id/default", ctrl.Address.SetDefault)
	protected.GET("/addresses/default", ctrl.Address.GetDefault)
}

package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/controllers"
	"github.com/gin-gonic/gin"
)

func SetupUploadRoutes(public, protected *gin.RouterGroup, ctrl *controllers.Container) {
	public.GET("/uploads/config", ctrl.Upload.GetUploadConfig)
	protected.POST("/uploads/presign", ctrl.Upload.PresignUpload)
}

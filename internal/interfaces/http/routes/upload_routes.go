package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/gin-gonic/gin"
)

func SetupUploadRoutes(public, protected *gin.RouterGroup, ctrl *handlers.Container) {
	public.GET("/uploads/config", ctrl.Upload.GetUploadConfig)
	protected.POST("/uploads/presign", ctrl.Upload.PresignUpload)
}

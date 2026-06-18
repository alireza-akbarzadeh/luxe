package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/controllers"
	"github.com/gin-gonic/gin"
)

func setupRoleRoutes(admin *gin.RouterGroup, ctrl *controllers.Container) {
	admin.GET("/roles", ctrl.Role.ListRoles)
	admin.POST("/roles", ctrl.Role.CreateRole)
	admin.GET("/roles/:id", ctrl.Role.GetRole)
	admin.PUT("/roles/:id", ctrl.Role.UpdateRole)
	admin.DELETE("/roles/:id", ctrl.Role.DeleteRole)
	admin.PUT("/roles/:id/permissions", ctrl.Role.SetRolePermissions)
	admin.GET("/permissions", ctrl.Role.ListPermissions)
}

package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
)

func SetupMenuRoutes(router *gin.RouterGroup, ctrl *handlers.Container) {
	adminGroup := router.Group("/admin/menu")
	adminGroup.Use(middleware.ModuleGuard("menus"))
	{
		adminGroup.GET("/groups", ctrl.Menu.GetAllGroups)
		adminGroup.GET("/groups/:id", ctrl.Menu.GetGroupByID)
		adminGroup.POST("/groups", ctrl.Menu.CreateGroup)
		adminGroup.PUT("/groups/:id", ctrl.Menu.UpdateGroup)
		adminGroup.DELETE("/groups/:id", ctrl.Menu.DeleteGroup)

		adminGroup.GET("/items", ctrl.Menu.GetAllItems)
		adminGroup.GET("/items/:id", ctrl.Menu.GetItemByID)
		adminGroup.POST("/items", ctrl.Menu.CreateItem)
		adminGroup.PUT("/items/:id", ctrl.Menu.UpdateItem)
		adminGroup.DELETE("/items/:id", ctrl.Menu.DeleteItem)
	}

	userGroup := router.Group("/user/menu")
	{
		userGroup.GET("/structure", ctrl.Menu.GetUserMenuStructure)
		userGroup.GET("/", ctrl.Menu.GetUserMenu)
	}
}

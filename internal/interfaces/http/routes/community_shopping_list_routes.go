package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/gin-gonic/gin"
)

// SetupCommunityShoppingListRoutes registers public community shopping list endpoints.
func SetupCommunityShoppingListRoutes(public *gin.RouterGroup, ctrl *handlers.Container) {
	public.GET("/community-lists", ctrl.CommunityShoppingList.ListCommunityLists)
	public.GET("/community-lists/:slug", ctrl.CommunityShoppingList.GetCommunityList)
}

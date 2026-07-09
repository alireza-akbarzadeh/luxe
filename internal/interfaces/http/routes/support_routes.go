package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
)

func SetupSupportRoutes(public, protected *gin.RouterGroup, ctrl *handlers.Container) {
	public.POST("/support/tickets", ctrl.Support.CreateSupportTicket)

	support := protected.Group("/support/tickets")
	{
		support.GET("/my", ctrl.Support.GetMySupportTickets)
		support.GET("/:id", ctrl.Support.GetSupportTicket)
		support.POST("/:id/messages", ctrl.Support.AddSupportTicketMessage)
	}

	admin := protected.Group("/admin/support")
	admin.Use(middleware.ModuleGuard("users"))
	{
		admin.GET("/stats", ctrl.Support.GetSupportStatsAdmin)
		admin.GET("/tickets", ctrl.Support.ListSupportTicketsAdmin)
		admin.GET("/tickets/:id", ctrl.Support.GetSupportTicketAdmin)
		admin.POST("/tickets/:id/messages", ctrl.Support.AddSupportTicketMessageAdmin)
		admin.PATCH("/tickets/:id/notes", ctrl.Support.UpdateSupportTicketNotesAdmin)
		admin.PATCH("/tickets/:id/status", ctrl.Support.UpdateSupportTicketStatusAdmin)
		admin.PATCH("/tickets/:id/assign", ctrl.Support.UpdateSupportTicketAssigneeAdmin)
		admin.POST("/tickets/:id/suggest-reply", ctrl.Support.SuggestSupportReplyAdmin)
	}
}

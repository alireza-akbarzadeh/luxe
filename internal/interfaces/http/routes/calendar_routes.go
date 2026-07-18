package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
)

// SetupCalendarRoutes registers the store working calendar & delivery
// availability admin API. All routes require JWT auth + admin role.
func SetupCalendarRoutes(protected *gin.RouterGroup, ctrl *handlers.Container) {
	admin := protected.Group("/admin/calendar")
	admin.Use(middleware.RequireAdmin())
	{
		admin.GET("/summary", ctrl.Calendar.GetSummary)
		admin.GET("/events", ctrl.Calendar.ListEvents)
		admin.GET("/day/:date", ctrl.Calendar.GetDay)
		admin.GET("/upcoming-events", ctrl.Calendar.ListUpcomingEvents)
		admin.GET("/stores-status-today", ctrl.Calendar.ListStoresStatusToday)

		admin.GET("/holidays", ctrl.Calendar.ListHolidays)
		admin.POST("/holidays", ctrl.Calendar.CreateHoliday)
		admin.GET("/holidays/:id", ctrl.Calendar.GetHoliday)
		admin.PUT("/holidays/:id", ctrl.Calendar.UpdateHoliday)
		admin.DELETE("/holidays/:id", ctrl.Calendar.DeleteHoliday)
		admin.POST("/holidays/:id/duplicate", ctrl.Calendar.DuplicateHoliday)
		admin.POST("/holidays/:id/publish", ctrl.Calendar.PublishHoliday)

		admin.GET("/schedules", ctrl.Calendar.ListSchedules)
		admin.POST("/schedules", ctrl.Calendar.CreateSchedule)
		admin.PUT("/schedules/:id", ctrl.Calendar.UpdateSchedule)

		admin.GET("/vendor-off-days", ctrl.Calendar.ListVendorOffDays)
		admin.POST("/vendor-off-days", ctrl.Calendar.CreateVendorOffDay)
		admin.PUT("/vendor-off-days/:id", ctrl.Calendar.UpdateVendorOffDay)
		admin.DELETE("/vendor-off-days/:id", ctrl.Calendar.DeleteVendorOffDay)

		admin.GET("/rules", ctrl.Calendar.ListRules)
		admin.PUT("/rules", ctrl.Calendar.UpdateRules)

		admin.POST("/simulate", ctrl.Calendar.Simulate)
	}
}

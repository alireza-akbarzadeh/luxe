package handlers

import (
	"net/http"
	"strconv"

	appcalendar "github.com/alireza-akbarzadeh/luxe/internal/application/calendar"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// CalendarHandler handles the store working calendar & delivery availability admin API.
type CalendarHandler struct {
	service  *appcalendar.Service
	validate *validator.Validate
}

// NewCalendarHandler creates a new CalendarHandler.
func NewCalendarHandler(service *appcalendar.Service) *CalendarHandler {
	return &CalendarHandler{service: service, validate: validator.New()}
}

// ─── Summary / events / day detail ─────────────────────────────────────────

// GetSummary godoc
// @Summary      Calendar summary KPIs
// @Description  Returns store working calendar dashboard KPIs (active stores, closures, delivery delay, capacity).
// @Tags         admin-calendar
// @Produce      json
// @Success      200  {object}  utils.Response{data=dto.CalendarSummaryResponse}
// @Failure      500  {object}  utils.Response
// @Security     BearerAuth
// @Router       /admin/calendar/summary [get]
func (h *CalendarHandler) GetSummary(c *gin.Context) {
	summary, err := h.service.GetSummary(c.Request.Context())
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load calendar summary")
		return
	}
	utils.SuccessResponse(c, "calendar summary retrieved", summary)
}

// ListEvents godoc
// @Summary      Month calendar grid
// @Description  Returns per-day status/badges for the calendar month grid.
// @Tags         admin-calendar
// @Produce      json
// @Param        year      query     int     true   "Year"
// @Param        month     query     int     true   "Month (1-12)"
// @Param        store_id  query     int     false  "Filter by store"
// @Param        region    query     string  false  "Filter by region"
// @Param        status    query     string  false  "Filter by holiday status"
// @Success      200  {object}  dto.ListCalendarEventsResponse
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Security     BearerAuth
// @Router       /admin/calendar/events [get]
func (h *CalendarHandler) ListEvents(c *gin.Context) {
	var req dto.ListCalendarEventsRequest
	if !utils.BindAndValidateQuery(c, &req, h.validate) {
		return
	}

	events, err := h.service.ListEvents(c.Request.Context(), &req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list calendar events")
		return
	}
	utils.SuccessResponse(c, "calendar events retrieved", events)
}

// GetDay godoc
// @Summary      Day detail
// @Description  Returns holidays, off days, and schedule detail for a single calendar day.
// @Tags         admin-calendar
// @Produce      json
// @Param        date      path      string  true   "Date (YYYY-MM-DD)"
// @Param        store_id  query     int     false  "Store id"
// @Success      200  {object}  utils.Response{data=dto.CalendarDayDetailResponse}
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Security     BearerAuth
// @Router       /admin/calendar/day/{date} [get]
func (h *CalendarHandler) GetDay(c *gin.Context) {
	date := c.Param("date")
	if date == "" {
		utils.BadRequestResponse(c, "date is required")
		return
	}
	var storeID uint
	if v, ok := parseOptionalUintQuery(c, "store_id"); ok {
		storeID = v
	}

	detail, err := h.service.GetDayDetail(c.Request.Context(), date, storeID)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load day detail")
		return
	}
	utils.SuccessResponse(c, "day detail retrieved", detail)
}

// ListUpcomingEvents godoc
// @Summary      Upcoming holidays and off days
// @Tags         admin-calendar
// @Produce      json
// @Param        limit  query  int  false  "Max results" default(10)
// @Success      200  {object}  dto.ListUpcomingEventsResponse
// @Failure      500  {object}  utils.Response
// @Security     BearerAuth
// @Router       /admin/calendar/upcoming-events [get]
func (h *CalendarHandler) ListUpcomingEvents(c *gin.Context) {
	limit, _ := paginationParams(c, 10)
	events, err := h.service.ListUpcomingEvents(c.Request.Context(), limit)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list upcoming events")
		return
	}
	utils.SuccessResponse(c, "upcoming events retrieved", events)
}

// ListStoresStatusToday godoc
// @Summary      Store open/closed status today
// @Tags         admin-calendar
// @Produce      json
// @Success      200  {object}  dto.ListStoresStatusTodayResponse
// @Failure      500  {object}  utils.Response
// @Security     BearerAuth
// @Router       /admin/calendar/stores-status-today [get]
func (h *CalendarHandler) ListStoresStatusToday(c *gin.Context) {
	statuses, err := h.service.ListStoresStatusToday(c.Request.Context())
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load store status")
		return
	}
	utils.SuccessResponse(c, "store status retrieved", statuses)
}

// ─── Holidays ────────────────────────────────────────────────────────────────

// ListHolidays godoc
// @Summary      List store holidays
// @Tags         admin-calendar
// @Produce      json
// @Param        page          query  int     false  "Page number"    default(1)
// @Param        limit         query  int     false  "Items per page" default(20)
// @Param        search        query  string  false  "Search name/description"
// @Param        holiday_type  query  string  false  "national|regional|store|vendor"
// @Param        status        query  string  false  "draft|published"
// @Param        store_id      query  int     false  "Filter by scoped store"
// @Param        region        query  string  false  "Filter by region"
// @Param        year          query  int     false  "Filter by year"
// @Param        month         query  int     false  "Filter by month"
// @Success      200  {object}  dto.StoreHolidayListResponse
// @Failure      500  {object}  utils.Response
// @Security     BearerAuth
// @Router       /admin/calendar/holidays [get]
func (h *CalendarHandler) ListHolidays(c *gin.Context) {
	var req dto.ListStoreHolidaysRequest
	if !utils.BindAndValidateQuery(c, &req, h.validate) {
		return
	}
	holidays, total, err := h.service.ListHolidays(c.Request.Context(), &req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list holidays")
		return
	}
	c.JSON(http.StatusOK, dto.StoreHolidayListResponse{
		BaseResponse: dto.BaseResponse{Success: true, Message: "holidays retrieved", Code: http.StatusOK},
		Data: dto.StoreHolidayListData{
			Holidays: holidays,
			Total:    total,
			Page:     req.Page,
			Limit:    req.Limit,
		},
	})
}

// GetHoliday godoc
// @Summary      Get a store holiday
// @Tags         admin-calendar
// @Produce      json
// @Param        id  path  int  true  "Holiday id"
// @Success      200  {object}  utils.Response{data=dto.StoreHolidayResponse}
// @Failure      404  {object}  utils.Response
// @Security     BearerAuth
// @Router       /admin/calendar/holidays/{id} [get]
func (h *CalendarHandler) GetHoliday(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	holiday, err := h.service.GetHoliday(c.Request.Context(), id)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to retrieve holiday")
		return
	}
	utils.SuccessResponse(c, "holiday retrieved", holiday)
}

// CreateHoliday godoc
// @Summary      Create a store holiday
// @Tags         admin-calendar
// @Accept       json
// @Produce      json
// @Param        request  body  dto.CreateStoreHolidayRequest  true  "Holiday payload"
// @Success      201  {object}  utils.Response{data=dto.StoreHolidayResponse}
// @Failure      400  {object}  utils.Response
// @Security     BearerAuth
// @Router       /admin/calendar/holidays [post]
func (h *CalendarHandler) CreateHoliday(c *gin.Context) {
	var req dto.CreateStoreHolidayRequest
	if !utils.BindAndValidate(c, &req, h.validate) {
		return
	}
	var createdBy *uint
	if userID, ok := middleware.GetUserID(c); ok {
		createdBy = &userID
	}
	holiday, err := h.service.CreateHoliday(c.Request.Context(), &req, createdBy)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to create holiday")
		return
	}
	utils.CreatedResponse(c, "holiday created successfully", holiday)
}

// UpdateHoliday godoc
// @Summary      Update a store holiday
// @Tags         admin-calendar
// @Accept       json
// @Produce      json
// @Param        id       path  int                            true  "Holiday id"
// @Param        request  body  dto.UpdateStoreHolidayRequest  true  "Update payload"
// @Success      200  {object}  utils.Response{data=dto.StoreHolidayResponse}
// @Failure      400  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Security     BearerAuth
// @Router       /admin/calendar/holidays/{id} [put]
func (h *CalendarHandler) UpdateHoliday(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateStoreHolidayRequest
	if !utils.BindAndValidate(c, &req, h.validate) {
		return
	}
	holiday, err := h.service.UpdateHoliday(c.Request.Context(), id, &req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to update holiday")
		return
	}
	utils.SuccessResponse(c, "holiday updated successfully", holiday)
}

// DeleteHoliday godoc
// @Summary      Delete a store holiday
// @Tags         admin-calendar
// @Produce      json
// @Param        id  path  int  true  "Holiday id"
// @Success      200  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Security     BearerAuth
// @Router       /admin/calendar/holidays/{id} [delete]
func (h *CalendarHandler) DeleteHoliday(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.service.DeleteHoliday(c.Request.Context(), id); err != nil {
		utils.HandleServiceError(c, err, "failed to delete holiday")
		return
	}
	utils.SuccessResponse(c, "holiday deleted successfully", nil)
}

// DuplicateHoliday godoc
// @Summary      Duplicate a store holiday
// @Tags         admin-calendar
// @Produce      json
// @Param        id  path  int  true  "Holiday id"
// @Success      201  {object}  utils.Response{data=dto.StoreHolidayResponse}
// @Failure      404  {object}  utils.Response
// @Security     BearerAuth
// @Router       /admin/calendar/holidays/{id}/duplicate [post]
func (h *CalendarHandler) DuplicateHoliday(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	holiday, err := h.service.DuplicateHoliday(c.Request.Context(), id)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to duplicate holiday")
		return
	}
	utils.CreatedResponse(c, "holiday duplicated successfully", holiday)
}

// PublishHoliday godoc
// @Summary      Publish a store holiday
// @Tags         admin-calendar
// @Produce      json
// @Param        id  path  int  true  "Holiday id"
// @Success      200  {object}  utils.Response{data=dto.StoreHolidayResponse}
// @Failure      404  {object}  utils.Response
// @Security     BearerAuth
// @Router       /admin/calendar/holidays/{id}/publish [post]
func (h *CalendarHandler) PublishHoliday(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	holiday, err := h.service.PublishHoliday(c.Request.Context(), id)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to publish holiday")
		return
	}
	utils.SuccessResponse(c, "holiday published successfully", holiday)
}

// ─── Store working schedules ────────────────────────────────────────────────

// ListSchedules godoc
// @Summary      List store working schedules
// @Tags         admin-calendar
// @Produce      json
// @Param        store_id  query  int  false  "Filter by store"
// @Success      200  {object}  dto.ListStoreWorkingSchedulesResponse
// @Failure      500  {object}  utils.Response
// @Security     BearerAuth
// @Router       /admin/calendar/schedules [get]
func (h *CalendarHandler) ListSchedules(c *gin.Context) {
	var storeIDPtr *uint
	if v, ok := parseOptionalUintQuery(c, "store_id"); ok {
		storeIDPtr = &v
	}
	schedules, err := h.service.ListSchedules(c.Request.Context(), storeIDPtr)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list working schedules")
		return
	}
	c.JSON(http.StatusOK, dto.ListStoreWorkingSchedulesResponse{
		BaseResponse: dto.BaseResponse{Success: true, Message: "working schedules retrieved", Code: http.StatusOK},
		Data:         schedules,
	})
}

// CreateSchedule godoc
// @Summary      Create a store working schedule
// @Tags         admin-calendar
// @Accept       json
// @Produce      json
// @Param        request  body  dto.CreateStoreWorkingScheduleRequest  true  "Schedule payload"
// @Success      201  {object}  utils.Response{data=dto.StoreWorkingScheduleResponse}
// @Failure      400  {object}  utils.Response
// @Security     BearerAuth
// @Router       /admin/calendar/schedules [post]
func (h *CalendarHandler) CreateSchedule(c *gin.Context) {
	var req dto.CreateStoreWorkingScheduleRequest
	if !utils.BindAndValidate(c, &req, h.validate) {
		return
	}
	sched, err := h.service.CreateSchedule(c.Request.Context(), &req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to create working schedule")
		return
	}
	utils.CreatedResponse(c, "working schedule created successfully", sched)
}

// UpdateSchedule godoc
// @Summary      Update a store working schedule
// @Tags         admin-calendar
// @Accept       json
// @Produce      json
// @Param        id       path  int                                     true  "Schedule id"
// @Param        request  body  dto.UpdateStoreWorkingScheduleRequest  true  "Update payload"
// @Success      200  {object}  utils.Response{data=dto.StoreWorkingScheduleResponse}
// @Failure      400  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Security     BearerAuth
// @Router       /admin/calendar/schedules/{id} [put]
func (h *CalendarHandler) UpdateSchedule(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateStoreWorkingScheduleRequest
	if !utils.BindAndValidate(c, &req, h.validate) {
		return
	}
	sched, err := h.service.UpdateSchedule(c.Request.Context(), id, &req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to update working schedule")
		return
	}
	utils.SuccessResponse(c, "working schedule updated successfully", sched)
}

// ─── Vendor off days ────────────────────────────────────────────────────────

// ListVendorOffDays godoc
// @Summary      List vendor off days
// @Tags         admin-calendar
// @Produce      json
// @Param        page       query  int     false  "Page number"    default(1)
// @Param        limit      query  int     false  "Items per page" default(20)
// @Param        vendor_id  query  int     false  "Filter by vendor"
// @Param        status     query  string  false  "draft|published"
// @Param        off_type   query  string  false  "Filter by off type"
// @Success      200  {object}  dto.VendorOffDayListResponse
// @Failure      500  {object}  utils.Response
// @Security     BearerAuth
// @Router       /admin/calendar/vendor-off-days [get]
func (h *CalendarHandler) ListVendorOffDays(c *gin.Context) {
	var req dto.ListVendorOffDaysRequest
	if !utils.BindAndValidateQuery(c, &req, h.validate) {
		return
	}
	offDays, total, err := h.service.ListOffDays(c.Request.Context(), &req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list vendor off days")
		return
	}
	c.JSON(http.StatusOK, dto.VendorOffDayListResponse{
		BaseResponse: dto.BaseResponse{Success: true, Message: "vendor off days retrieved", Code: http.StatusOK},
		Data: dto.VendorOffDayListData{
			OffDays: offDays,
			Total:   total,
			Page:    req.Page,
			Limit:   req.Limit,
		},
	})
}

// CreateVendorOffDay godoc
// @Summary      Create a vendor off day
// @Tags         admin-calendar
// @Accept       json
// @Produce      json
// @Param        request  body  dto.CreateVendorOffDayRequest  true  "Off day payload"
// @Success      201  {object}  utils.Response{data=dto.VendorOffDayResponse}
// @Failure      400  {object}  utils.Response
// @Security     BearerAuth
// @Router       /admin/calendar/vendor-off-days [post]
func (h *CalendarHandler) CreateVendorOffDay(c *gin.Context) {
	var req dto.CreateVendorOffDayRequest
	if !utils.BindAndValidate(c, &req, h.validate) {
		return
	}
	offDay, err := h.service.CreateOffDay(c.Request.Context(), &req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to create vendor off day")
		return
	}
	utils.CreatedResponse(c, "vendor off day created successfully", offDay)
}

// UpdateVendorOffDay godoc
// @Summary      Update a vendor off day
// @Tags         admin-calendar
// @Accept       json
// @Produce      json
// @Param        id       path  int                            true  "Off day id"
// @Param        request  body  dto.UpdateVendorOffDayRequest  true  "Update payload"
// @Success      200  {object}  utils.Response{data=dto.VendorOffDayResponse}
// @Failure      400  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Security     BearerAuth
// @Router       /admin/calendar/vendor-off-days/{id} [put]
func (h *CalendarHandler) UpdateVendorOffDay(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateVendorOffDayRequest
	if !utils.BindAndValidate(c, &req, h.validate) {
		return
	}
	offDay, err := h.service.UpdateOffDay(c.Request.Context(), id, &req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to update vendor off day")
		return
	}
	utils.SuccessResponse(c, "vendor off day updated successfully", offDay)
}

// DeleteVendorOffDay godoc
// @Summary      Delete a vendor off day
// @Tags         admin-calendar
// @Produce      json
// @Param        id  path  int  true  "Off day id"
// @Success      200  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Security     BearerAuth
// @Router       /admin/calendar/vendor-off-days/{id} [delete]
func (h *CalendarHandler) DeleteVendorOffDay(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.service.DeleteOffDay(c.Request.Context(), id); err != nil {
		utils.HandleServiceError(c, err, "failed to delete vendor off day")
		return
	}
	utils.SuccessResponse(c, "vendor off day deleted successfully", nil)
}

// ─── Delivery calendar rules ────────────────────────────────────────────────

// ListRules godoc
// @Summary      List delivery calendar rules
// @Tags         admin-calendar
// @Produce      json
// @Success      200  {object}  dto.ListDeliveryCalendarRulesResponse
// @Failure      500  {object}  utils.Response
// @Security     BearerAuth
// @Router       /admin/calendar/rules [get]
func (h *CalendarHandler) ListRules(c *gin.Context) {
	rules, err := h.service.ListRules(c.Request.Context())
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list delivery calendar rules")
		return
	}
	c.JSON(http.StatusOK, dto.ListDeliveryCalendarRulesResponse{
		BaseResponse: dto.BaseResponse{Success: true, Message: "rules retrieved", Code: http.StatusOK},
		Data:         rules,
	})
}

// UpdateRules godoc
// @Summary      Bulk update delivery calendar rules
// @Description  Toggles the enabled flag for one or more rule keys.
// @Tags         admin-calendar
// @Accept       json
// @Produce      json
// @Param        request  body  dto.UpdateDeliveryCalendarRulesRequest  true  "Rule toggles"
// @Success      200  {object}  dto.ListDeliveryCalendarRulesResponse
// @Failure      400  {object}  utils.Response
// @Security     BearerAuth
// @Router       /admin/calendar/rules [put]
func (h *CalendarHandler) UpdateRules(c *gin.Context) {
	var req dto.UpdateDeliveryCalendarRulesRequest
	if !utils.BindAndValidate(c, &req, h.validate) {
		return
	}
	rules, err := h.service.UpdateRules(c.Request.Context(), &req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to update delivery calendar rules")
		return
	}
	c.JSON(http.StatusOK, dto.ListDeliveryCalendarRulesResponse{
		BaseResponse: dto.BaseResponse{Success: true, Message: "rules updated successfully", Code: http.StatusOK},
		Data:         rules,
	})
}

// ─── Delivery simulator ─────────────────────────────────────────────────────

// Simulate godoc
// @Summary      Simulate earliest delivery date
// @Description  Runs the delivery calculator against vendor schedule, holidays, off days, and courier rules.
// @Tags         admin-calendar
// @Accept       json
// @Produce      json
// @Param        request  body  dto.SimulateDeliveryRequest  true  "Simulation input"
// @Success      200  {object}  utils.Response{data=dto.SimulateDeliveryResponse}
// @Failure      400  {object}  utils.Response
// @Security     BearerAuth
// @Router       /admin/calendar/simulate [post]
func (h *CalendarHandler) Simulate(c *gin.Context) {
	var req dto.SimulateDeliveryRequest
	if !utils.BindAndValidate(c, &req, h.validate) {
		return
	}
	result, err := h.service.Simulate(c.Request.Context(), &req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to simulate delivery")
		return
	}
	utils.SuccessResponse(c, "delivery simulation computed", result)
}

// parseOptionalUintQuery reads an optional uint query param without writing an error response.
func parseOptionalUintQuery(c *gin.Context, key string) (uint, bool) {
	raw := c.Query(key)
	if raw == "" {
		return 0, false
	}
	v, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, false
	}
	return uint(v), true
}

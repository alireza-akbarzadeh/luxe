package dto

import "time"

// ─── Store working schedules ───────────────────────────────────────────────

// CreateStoreWorkingScheduleRequest creates a store's working schedule.
type CreateStoreWorkingScheduleRequest struct {
	StoreID             uint            `json:"store_id" binding:"required,gt=0"`
	Timezone            *string         `json:"timezone" binding:"omitempty,max=64"`
	WorkingDays         map[string]bool `json:"working_days" binding:"omitempty"`
	OpenTime            *string         `json:"open_time" binding:"omitempty,max=5"`
	CloseTime           *string         `json:"close_time" binding:"omitempty,max=5"`
	BreakStart          *string         `json:"break_start" binding:"omitempty,max=5"`
	BreakEnd            *string         `json:"break_end" binding:"omitempty,max=5"`
	MaxOrdersPerDay     *int            `json:"max_orders_per_day" binding:"omitempty,gte=0"`
	MaxDeliveriesPerDay *int            `json:"max_deliveries_per_day" binding:"omitempty,gte=0"`
	PrepLeadHours       *int            `json:"prep_lead_hours" binding:"omitempty,gte=0"`
	ProcessingLeadHours *int            `json:"processing_lead_hours" binding:"omitempty,gte=0"`
	DeliveryBufferHours *int            `json:"delivery_buffer_hours" binding:"omitempty,gte=0"`
}

// UpdateStoreWorkingScheduleRequest updates an existing working schedule.
type UpdateStoreWorkingScheduleRequest struct {
	Timezone            *string         `json:"timezone" binding:"omitempty,max=64"`
	WorkingDays         map[string]bool `json:"working_days" binding:"omitempty"`
	OpenTime            *string         `json:"open_time" binding:"omitempty,max=5"`
	CloseTime           *string         `json:"close_time" binding:"omitempty,max=5"`
	BreakStart          *string         `json:"break_start" binding:"omitempty,max=5"`
	BreakEnd            *string         `json:"break_end" binding:"omitempty,max=5"`
	MaxOrdersPerDay     *int            `json:"max_orders_per_day" binding:"omitempty,gte=0"`
	MaxDeliveriesPerDay *int            `json:"max_deliveries_per_day" binding:"omitempty,gte=0"`
	PrepLeadHours       *int            `json:"prep_lead_hours" binding:"omitempty,gte=0"`
	ProcessingLeadHours *int            `json:"processing_lead_hours" binding:"omitempty,gte=0"`
	DeliveryBufferHours *int            `json:"delivery_buffer_hours" binding:"omitempty,gte=0"`
}

// StoreWorkingScheduleResponse is the API shape for a store working schedule.
type StoreWorkingScheduleResponse struct {
	ID                  uint            `json:"id"`
	StoreID             uint            `json:"store_id"`
	Timezone            string          `json:"timezone"`
	WorkingDays         map[string]bool `json:"working_days"`
	OpenTime            string          `json:"open_time"`
	CloseTime           string          `json:"close_time"`
	BreakStart          *string         `json:"break_start,omitempty"`
	BreakEnd            *string         `json:"break_end,omitempty"`
	MaxOrdersPerDay     int             `json:"max_orders_per_day"`
	MaxDeliveriesPerDay int             `json:"max_deliveries_per_day"`
	PrepLeadHours       int             `json:"prep_lead_hours"`
	ProcessingLeadHours int             `json:"processing_lead_hours"`
	DeliveryBufferHours int             `json:"delivery_buffer_hours"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
}

// ListStoreWorkingSchedulesResponse wraps a list of schedules.
type ListStoreWorkingSchedulesResponse struct {
	BaseResponse
	Data []StoreWorkingScheduleResponse `json:"data"`
}

// ─── Store holidays ─────────────────────────────────────────────────────────

// CreateStoreHolidayRequest creates a holiday/closure window.
type CreateStoreHolidayRequest struct {
	Name           string  `json:"name" binding:"required,min=2,max=255"`
	Description    string  `json:"description" binding:"omitempty"`
	HolidayType    string  `json:"holiday_type" binding:"required,oneof=national regional store vendor"`
	StartDate      string  `json:"start_date" binding:"required"`
	EndDate        string  `json:"end_date" binding:"required"`
	IsRecurring    bool    `json:"is_recurring"`
	RecurrenceRule *string `json:"recurrence_rule" binding:"omitempty,max=50"`
	ApplyTo        string  `json:"apply_to" binding:"required,oneof=all stores vendor region"`
	StoreIDs       []uint  `json:"store_ids" binding:"omitempty"`
	VendorID       *uint   `json:"vendor_id" binding:"omitempty,gt=0"`
	Region         *string `json:"region" binding:"omitempty,max=100"`
	Priority       int     `json:"priority"`
	Status         *string `json:"status" binding:"omitempty,oneof=draft published"`
	Notes          string  `json:"notes" binding:"omitempty"`
}

// UpdateStoreHolidayRequest updates an existing holiday.
type UpdateStoreHolidayRequest struct {
	Name           *string `json:"name" binding:"omitempty,min=2,max=255"`
	Description    *string `json:"description" binding:"omitempty"`
	HolidayType    *string `json:"holiday_type" binding:"omitempty,oneof=national regional store vendor"`
	StartDate      *string `json:"start_date" binding:"omitempty"`
	EndDate        *string `json:"end_date" binding:"omitempty"`
	IsRecurring    *bool   `json:"is_recurring"`
	RecurrenceRule *string `json:"recurrence_rule" binding:"omitempty,max=50"`
	ApplyTo        *string `json:"apply_to" binding:"omitempty,oneof=all stores vendor region"`
	StoreIDs       []uint  `json:"store_ids" binding:"omitempty"`
	VendorID       *uint   `json:"vendor_id" binding:"omitempty,gt=0"`
	Region         *string `json:"region" binding:"omitempty,max=100"`
	Priority       *int    `json:"priority"`
	Status         *string `json:"status" binding:"omitempty,oneof=draft published"`
	Notes          *string `json:"notes" binding:"omitempty"`
}

// ListStoreHolidaysRequest filters admin holiday lists.
type ListStoreHolidaysRequest struct {
	Page        int    `form:"page,default=1"`
	Limit       int    `form:"limit,default=20"`
	Search      string `form:"search"`
	HolidayType string `form:"holiday_type"`
	Status      string `form:"status"`
	StoreID     uint   `form:"store_id"`
	Region      string `form:"region"`
	Year        int    `form:"year"`
	Month       int    `form:"month"`
}

// StoreHolidayResponse is the API shape for a holiday.
type StoreHolidayResponse struct {
	ID             uint      `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	HolidayType    string    `json:"holiday_type"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	IsRecurring    bool      `json:"is_recurring"`
	RecurrenceRule *string   `json:"recurrence_rule,omitempty"`
	ApplyTo        string    `json:"apply_to"`
	StoreIDs       []uint    `json:"store_ids,omitempty"`
	VendorID       *uint     `json:"vendor_id,omitempty"`
	Region         *string   `json:"region,omitempty"`
	Priority       int       `json:"priority"`
	Status         string    `json:"status"`
	Notes          string    `json:"notes"`
	CreatedBy      *uint     `json:"created_by,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// StoreHolidayListData is the paginated holiday list payload.
type StoreHolidayListData struct {
	Holidays []StoreHolidayResponse `json:"holidays"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	Limit    int                    `json:"limit"`
}

// StoreHolidayListResponse wraps a paginated holiday list.
type StoreHolidayListResponse struct {
	BaseResponse
	Data StoreHolidayListData `json:"data"`
}

// ─── Vendor off days ────────────────────────────────────────────────────────

// CreateVendorOffDayRequest creates a vendor-initiated off day.
type CreateVendorOffDayRequest struct {
	VendorID  uint    `json:"vendor_id" binding:"required,gt=0"`
	Title     string  `json:"title" binding:"required,min=2,max=255"`
	OffType   string  `json:"off_type" binding:"required,oneof=vacation inventory_count maintenance emergency_close personal_leave"`
	StartDate string  `json:"start_date" binding:"required"`
	EndDate   string  `json:"end_date" binding:"required"`
	Notes     string  `json:"notes" binding:"omitempty"`
	Status    *string `json:"status" binding:"omitempty,oneof=draft published"`
}

// UpdateVendorOffDayRequest updates an existing vendor off day.
type UpdateVendorOffDayRequest struct {
	Title     *string `json:"title" binding:"omitempty,min=2,max=255"`
	OffType   *string `json:"off_type" binding:"omitempty,oneof=vacation inventory_count maintenance emergency_close personal_leave"`
	StartDate *string `json:"start_date" binding:"omitempty"`
	EndDate   *string `json:"end_date" binding:"omitempty"`
	Notes     *string `json:"notes" binding:"omitempty"`
	Status    *string `json:"status" binding:"omitempty,oneof=draft published"`
}

// ListVendorOffDaysRequest filters vendor off day lists.
type ListVendorOffDaysRequest struct {
	Page     int    `form:"page,default=1"`
	Limit    int    `form:"limit,default=20"`
	VendorID uint   `form:"vendor_id"`
	Status   string `form:"status"`
	OffType  string `form:"off_type"`
}

// VendorOffDayResponse is the API shape for a vendor off day.
type VendorOffDayResponse struct {
	ID        uint      `json:"id"`
	VendorID  uint      `json:"vendor_id"`
	Title     string    `json:"title"`
	OffType   string    `json:"off_type"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Notes     string    `json:"notes"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// VendorOffDayListData is the paginated vendor off day list payload.
type VendorOffDayListData struct {
	OffDays []VendorOffDayResponse `json:"off_days"`
	Total   int64                  `json:"total"`
	Page    int                    `json:"page"`
	Limit   int                    `json:"limit"`
}

// VendorOffDayListResponse wraps a paginated vendor off day list.
type VendorOffDayListResponse struct {
	BaseResponse
	Data VendorOffDayListData `json:"data"`
}

// ─── Delivery calendar rules ────────────────────────────────────────────────

// DeliveryCalendarRuleResponse is the API shape for a calendar rule toggle.
type DeliveryCalendarRuleResponse struct {
	ID          uint   `json:"id"`
	RuleKey     string `json:"rule_key"`
	Enabled     bool   `json:"enabled"`
	Label       string `json:"label"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}

// ListDeliveryCalendarRulesResponse wraps the full rule list.
type ListDeliveryCalendarRulesResponse struct {
	BaseResponse
	Data []DeliveryCalendarRuleResponse `json:"data"`
}

// UpdateDeliveryCalendarRuleItem toggles a single rule by key.
type UpdateDeliveryCalendarRuleItem struct {
	RuleKey string `json:"rule_key" binding:"required"`
	Enabled bool   `json:"enabled"`
}

// UpdateDeliveryCalendarRulesRequest bulk-updates rule enabled flags.
type UpdateDeliveryCalendarRulesRequest struct {
	Rules []UpdateDeliveryCalendarRuleItem `json:"rules" binding:"required,min=1,dive"`
}

// ─── Summary / KPIs ─────────────────────────────────────────────────────────

// CalendarSummaryResponse holds admin dashboard KPIs.
type CalendarSummaryResponse struct {
	TotalStores            int64   `json:"total_stores"`
	ActiveStores           int64   `json:"active_stores"`
	ClosedToday            int64   `json:"closed_today"`
	UpcomingHolidays       int64   `json:"upcoming_holidays"`
	NextDeliveryDelayDays  int     `json:"next_delivery_delay_days"`
	WorkingCapacityPercent float64 `json:"working_capacity_percent"`
}

// ─── Calendar events (month grid) ───────────────────────────────────────────

// ListCalendarEventsRequest filters the month calendar grid.
type ListCalendarEventsRequest struct {
	Year    int    `form:"year" binding:"required"`
	Month   int    `form:"month" binding:"required,min=1,max=12"`
	StoreID uint   `form:"store_id"`
	Region  string `form:"region"`
	Status  string `form:"status"`
}

// CalendarDayEventResponse describes a single day's status/badges on the calendar grid.
type CalendarDayEventResponse struct {
	Date       string   `json:"date"`
	DayType    string   `json:"day_type"`
	Badges     []string `json:"badges,omitempty"`
	HolidayIDs []uint   `json:"holiday_ids,omitempty"`
	OffDayIDs  []uint   `json:"off_day_ids,omitempty"`
}

// ListCalendarEventsResponse wraps the month calendar grid.
type ListCalendarEventsResponse struct {
	BaseResponse
	Data []CalendarDayEventResponse `json:"data"`
}

// CalendarDayDetailResponse is the drawer detail for a single day.
type CalendarDayDetailResponse struct {
	Date         string                        `json:"date"`
	DayType      string                        `json:"day_type"`
	IsWorkingDay bool                          `json:"is_working_day"`
	Holidays     []StoreHolidayResponse        `json:"holidays,omitempty"`
	OffDays      []VendorOffDayResponse        `json:"off_days,omitempty"`
	Schedule     *StoreWorkingScheduleResponse `json:"schedule,omitempty"`
}

// UpcomingEventResponse is a single upcoming holiday/off day entry.
type UpcomingEventResponse struct {
	Type      string    `json:"type"`
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Scope     string    `json:"scope,omitempty"`
}

// ListUpcomingEventsResponse wraps upcoming holidays/off days.
type ListUpcomingEventsResponse struct {
	BaseResponse
	Data []UpcomingEventResponse `json:"data"`
}

// StoreStatusTodayResponse reports whether a store is open today.
type StoreStatusTodayResponse struct {
	StoreID        uint   `json:"store_id"`
	StoreName      string `json:"store_name"`
	IsWorkingToday bool   `json:"is_working_today"`
	Reason         string `json:"reason,omitempty"`
}

// ListStoresStatusTodayResponse wraps today's status for all stores.
type ListStoresStatusTodayResponse struct {
	BaseResponse
	Data []StoreStatusTodayResponse `json:"data"`
}

// ─── Delivery simulator ─────────────────────────────────────────────────────

// SimulateDeliveryRequest is the delivery calculator input.
type SimulateDeliveryRequest struct {
	StoreID        uint    `json:"store_id" binding:"required,gt=0"`
	VendorID       *uint   `json:"vendor_id" binding:"omitempty,gt=0"`
	City           *string `json:"city" binding:"omitempty,max=100"`
	Region         *string `json:"region" binding:"omitempty,max=100"`
	OrderDate      string  `json:"order_date" binding:"required"`
	ShippingMethod string  `json:"shipping_method" binding:"required,oneof=standard express pickup"`
	ShippingDays   int     `json:"shipping_days" binding:"gte=0"`
}

// DeliveryTimelineStep is one step in the delivery calculator's timeline.
type DeliveryTimelineStep struct {
	Step  string    `json:"step"`
	Date  time.Time `json:"date"`
	Days  int       `json:"days"`
	Label string    `json:"label"`
}

// SimulateDeliveryResponse is the delivery calculator output.
type SimulateDeliveryResponse struct {
	ProcessingStart time.Time              `json:"processing_start"`
	VendorAvailable bool                   `json:"vendor_available"`
	SkippedHolidays []string               `json:"skipped_holidays"`
	SkippedWeekends []string               `json:"skipped_weekends"`
	DeliveryDate    time.Time              `json:"delivery_date"`
	DelayReason     string                 `json:"delay_reason,omitempty"`
	Timeline        []DeliveryTimelineStep `json:"timeline"`
}

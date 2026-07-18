package calendar

import (
	"encoding/json"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/datatypes"
)

// dayKeys maps Go's time.Weekday to the store working-day JSON keys.
var dayKeys = map[time.Weekday]string{
	time.Monday:    "mon",
	time.Tuesday:   "tue",
	time.Wednesday: "wed",
	time.Thursday:  "thu",
	time.Friday:    "fri",
	time.Saturday:  "sat",
	time.Sunday:    "sun",
}

func defaultWorkingDays() map[string]bool {
	return map[string]bool{
		"mon": true, "tue": true, "wed": true, "thu": true, "fri": true,
		"sat": false, "sun": false,
	}
}

// parseWorkingDays decodes the jsonb working_days column, falling back to defaults.
func parseWorkingDays(raw datatypes.JSON) map[string]bool {
	days := defaultWorkingDays()
	if len(raw) == 0 {
		return days
	}
	var parsed map[string]bool
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return days
	}
	for k, v := range parsed {
		days[k] = v
	}
	return days
}

func encodeWorkingDays(days map[string]bool) (datatypes.JSON, error) {
	if days == nil {
		days = defaultWorkingDays()
	}
	b, err := json.Marshal(days)
	if err != nil {
		return nil, err
	}
	return datatypes.JSON(b), nil
}

// isWorkingDay reports whether the given date is a working day per the schedule.
func isWorkingDay(days map[string]bool, date time.Time) bool {
	key, ok := dayKeys[date.Weekday()]
	if !ok {
		return true
	}
	working, ok := days[key]
	if !ok {
		return true
	}
	return working
}

// ─── Store working schedule mapping ─────────────────────────────────────────

func toScheduleResponse(s *models.StoreWorkingSchedule) *dto.StoreWorkingScheduleResponse {
	return &dto.StoreWorkingScheduleResponse{
		ID:                  s.ID,
		StoreID:             s.StoreID,
		Timezone:            s.Timezone,
		WorkingDays:         parseWorkingDays(s.WorkingDays),
		OpenTime:            s.OpenTime,
		CloseTime:           s.CloseTime,
		BreakStart:          s.BreakStart,
		BreakEnd:            s.BreakEnd,
		MaxOrdersPerDay:     s.MaxOrdersPerDay,
		MaxDeliveriesPerDay: s.MaxDeliveriesPerDay,
		PrepLeadHours:       s.PrepLeadHours,
		ProcessingLeadHours: s.ProcessingLeadHours,
		DeliveryBufferHours: s.DeliveryBufferHours,
		CreatedAt:           s.CreatedAt,
		UpdatedAt:           s.UpdatedAt,
	}
}

func buildScheduleCreateModel(req *dto.CreateStoreWorkingScheduleRequest) (*models.StoreWorkingSchedule, error) {
	days, err := encodeWorkingDays(req.WorkingDays)
	if err != nil {
		return nil, err
	}
	sched := &models.StoreWorkingSchedule{
		StoreID:             req.StoreID,
		Timezone:            "UTC",
		WorkingDays:         days,
		OpenTime:            "09:00",
		CloseTime:           "18:00",
		ProcessingLeadHours: 24,
	}
	if req.Timezone != nil && *req.Timezone != "" {
		sched.Timezone = *req.Timezone
	}
	if req.OpenTime != nil && *req.OpenTime != "" {
		sched.OpenTime = *req.OpenTime
	}
	if req.CloseTime != nil && *req.CloseTime != "" {
		sched.CloseTime = *req.CloseTime
	}
	sched.BreakStart = req.BreakStart
	sched.BreakEnd = req.BreakEnd
	if req.MaxOrdersPerDay != nil {
		sched.MaxOrdersPerDay = *req.MaxOrdersPerDay
	}
	if req.MaxDeliveriesPerDay != nil {
		sched.MaxDeliveriesPerDay = *req.MaxDeliveriesPerDay
	}
	if req.PrepLeadHours != nil {
		sched.PrepLeadHours = *req.PrepLeadHours
	}
	if req.ProcessingLeadHours != nil {
		sched.ProcessingLeadHours = *req.ProcessingLeadHours
	}
	if req.DeliveryBufferHours != nil {
		sched.DeliveryBufferHours = *req.DeliveryBufferHours
	}
	return sched, nil
}

func applyScheduleUpdate(sched *models.StoreWorkingSchedule, req *dto.UpdateStoreWorkingScheduleRequest) error {
	if req.Timezone != nil {
		sched.Timezone = *req.Timezone
	}
	if req.WorkingDays != nil {
		days, err := encodeWorkingDays(req.WorkingDays)
		if err != nil {
			return err
		}
		sched.WorkingDays = days
	}
	if req.OpenTime != nil {
		sched.OpenTime = *req.OpenTime
	}
	if req.CloseTime != nil {
		sched.CloseTime = *req.CloseTime
	}
	if req.BreakStart != nil {
		sched.BreakStart = req.BreakStart
	}
	if req.BreakEnd != nil {
		sched.BreakEnd = req.BreakEnd
	}
	if req.MaxOrdersPerDay != nil {
		sched.MaxOrdersPerDay = *req.MaxOrdersPerDay
	}
	if req.MaxDeliveriesPerDay != nil {
		sched.MaxDeliveriesPerDay = *req.MaxDeliveriesPerDay
	}
	if req.PrepLeadHours != nil {
		sched.PrepLeadHours = *req.PrepLeadHours
	}
	if req.ProcessingLeadHours != nil {
		sched.ProcessingLeadHours = *req.ProcessingLeadHours
	}
	if req.DeliveryBufferHours != nil {
		sched.DeliveryBufferHours = *req.DeliveryBufferHours
	}
	return nil
}

// ─── Store holiday mapping ───────────────────────────────────────────────────

func toHolidayResponse(h *models.StoreHoliday, storeIDs []uint) *dto.StoreHolidayResponse {
	return &dto.StoreHolidayResponse{
		ID:             h.ID,
		Name:           h.Name,
		Description:    h.Description,
		HolidayType:    h.HolidayType,
		StartDate:      h.StartDate,
		EndDate:        h.EndDate,
		IsRecurring:    h.IsRecurring,
		RecurrenceRule: h.RecurrenceRule,
		ApplyTo:        h.ApplyTo,
		StoreIDs:       storeIDs,
		VendorID:       h.VendorID,
		Region:         h.Region,
		Priority:       h.Priority,
		Status:         h.Status,
		Notes:          h.Notes,
		CreatedBy:      h.CreatedBy,
		CreatedAt:      h.CreatedAt,
		UpdatedAt:      h.UpdatedAt,
	}
}

func buildHolidayCreateModel(req *dto.CreateStoreHolidayRequest, createdBy *uint) (*models.StoreHoliday, error) {
	start, err := parseDate(req.StartDate)
	if err != nil {
		return nil, err
	}
	end, err := parseDate(req.EndDate)
	if err != nil {
		return nil, err
	}
	status := "draft"
	if req.Status != nil && *req.Status != "" {
		status = *req.Status
	}
	return &models.StoreHoliday{
		Name:           req.Name,
		Description:    req.Description,
		HolidayType:    req.HolidayType,
		StartDate:      start,
		EndDate:        end,
		IsRecurring:    req.IsRecurring,
		RecurrenceRule: req.RecurrenceRule,
		ApplyTo:        req.ApplyTo,
		VendorID:       req.VendorID,
		Region:         req.Region,
		Priority:       req.Priority,
		Status:         status,
		Notes:          req.Notes,
		CreatedBy:      createdBy,
	}, nil
}

func applyHolidayUpdate(h *models.StoreHoliday, req *dto.UpdateStoreHolidayRequest) error {
	if req.Name != nil {
		h.Name = *req.Name
	}
	if req.Description != nil {
		h.Description = *req.Description
	}
	if req.HolidayType != nil {
		h.HolidayType = *req.HolidayType
	}
	if req.StartDate != nil {
		start, err := parseDate(*req.StartDate)
		if err != nil {
			return err
		}
		h.StartDate = start
	}
	if req.EndDate != nil {
		end, err := parseDate(*req.EndDate)
		if err != nil {
			return err
		}
		h.EndDate = end
	}
	if req.IsRecurring != nil {
		h.IsRecurring = *req.IsRecurring
	}
	if req.RecurrenceRule != nil {
		h.RecurrenceRule = req.RecurrenceRule
	}
	if req.ApplyTo != nil {
		h.ApplyTo = *req.ApplyTo
	}
	if req.VendorID != nil {
		h.VendorID = req.VendorID
	}
	if req.Region != nil {
		h.Region = req.Region
	}
	if req.Priority != nil {
		h.Priority = *req.Priority
	}
	if req.Status != nil {
		h.Status = *req.Status
	}
	if req.Notes != nil {
		h.Notes = *req.Notes
	}
	return nil
}

// ─── Vendor off day mapping ──────────────────────────────────────────────────

func toOffDayResponse(o *models.VendorOffDay) *dto.VendorOffDayResponse {
	return &dto.VendorOffDayResponse{
		ID:        o.ID,
		VendorID:  o.VendorID,
		Title:     o.Title,
		OffType:   o.OffType,
		StartDate: o.StartDate,
		EndDate:   o.EndDate,
		Notes:     o.Notes,
		Status:    o.Status,
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
	}
}

func buildOffDayCreateModel(req *dto.CreateVendorOffDayRequest) (*models.VendorOffDay, error) {
	start, err := parseDate(req.StartDate)
	if err != nil {
		return nil, err
	}
	end, err := parseDate(req.EndDate)
	if err != nil {
		return nil, err
	}
	status := "draft"
	if req.Status != nil && *req.Status != "" {
		status = *req.Status
	}
	return &models.VendorOffDay{
		VendorID:  req.VendorID,
		Title:     req.Title,
		OffType:   req.OffType,
		StartDate: start,
		EndDate:   end,
		Notes:     req.Notes,
		Status:    status,
	}, nil
}

func applyOffDayUpdate(o *models.VendorOffDay, req *dto.UpdateVendorOffDayRequest) error {
	if req.Title != nil {
		o.Title = *req.Title
	}
	if req.OffType != nil {
		o.OffType = *req.OffType
	}
	if req.StartDate != nil {
		start, err := parseDate(*req.StartDate)
		if err != nil {
			return err
		}
		o.StartDate = start
	}
	if req.EndDate != nil {
		end, err := parseDate(*req.EndDate)
		if err != nil {
			return err
		}
		o.EndDate = end
	}
	if req.Notes != nil {
		o.Notes = *req.Notes
	}
	if req.Status != nil {
		o.Status = *req.Status
	}
	return nil
}

// ─── Rule mapping ────────────────────────────────────────────────────────────

func toRuleResponse(r *models.DeliveryCalendarRule) dto.DeliveryCalendarRuleResponse {
	return dto.DeliveryCalendarRuleResponse{
		ID:          r.ID,
		RuleKey:     r.RuleKey,
		Enabled:     r.Enabled,
		Label:       r.Label,
		Description: r.Description,
		SortOrder:   r.SortOrder,
	}
}

// parseDate accepts RFC3339 timestamps or plain YYYY-MM-DD dates.
func parseDate(value string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, nil
	}
	return time.Parse("2006-01-02", value)
}

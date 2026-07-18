package calendar

import (
	"context"
	"fmt"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

// GetSummary computes the admin dashboard KPIs for the store working calendar.
func (s *Service) GetSummary(ctx context.Context) (*dto.CalendarSummaryResponse, error) {
	totalStores, err := s.storeRepo.CountAll(ctx)
	if err != nil {
		return nil, err
	}
	activeStores, err := s.storeRepo.CountByStatus(ctx, constants.StoreStatusActive)
	if err != nil {
		return nil, err
	}

	today := startOfDay(time.Now())
	stores, err := s.storeRepo.ListActiveLite(ctx)
	if err != nil {
		return nil, err
	}
	holidaysToday, err := s.holidayRepo.ListForRange(ctx, today, today.AddDate(0, 0, 1), nil, nil)
	if err != nil {
		return nil, err
	}
	offDaysToday, err := s.offDayRepo.ListForRange(ctx, nil, today, today.AddDate(0, 0, 1))
	if err != nil {
		return nil, err
	}

	var closedToday int64
	for _, store := range stores {
		open, _ := s.isStoreOpenOnDate(ctx, store.ID, today, holidaysToday, offDaysToday)
		if !open {
			closedToday++
		}
	}

	upcomingHolidays, err := s.holidayRepo.CountUpcoming(ctx, today)
	if err != nil {
		return nil, err
	}

	nextDelayDays := 0
	if len(stores) > 0 {
		result, err := s.CalculateEarliestDelivery(ctx, CalculateEarliestDeliveryInput{
			StoreID:        stores[0].ID,
			OrderDate:      time.Now(),
			ShippingMethod: "standard",
			ShippingDays:   0,
		})
		if err == nil {
			nextDelayDays = int(startOfDay(result.DeliveryDate).Sub(today).Hours() / 24)
		}
	}

	workingCapacityPercent := 100.0
	if len(stores) > 0 {
		workingCapacityPercent = float64(int64(len(stores))-closedToday) / float64(len(stores)) * 100
	}

	return &dto.CalendarSummaryResponse{
		TotalStores:            totalStores,
		ActiveStores:           activeStores,
		ClosedToday:            closedToday,
		UpcomingHolidays:       upcomingHolidays,
		NextDeliveryDelayDays:  nextDelayDays,
		WorkingCapacityPercent: workingCapacityPercent,
	}, nil
}

// isStoreOpenOnDate reports whether a store is open on the given date given
// pre-fetched holidays/off days, and a human-readable reason when closed.
func (s *Service) isStoreOpenOnDate(
	ctx context.Context,
	storeID uint,
	date time.Time,
	holidays []models.StoreHoliday,
	offDays []models.VendorOffDay,
) (bool, string) {
	if h := holidayForDate(holidays, date, "national", "regional", "store", "vendor"); h != nil {
		return false, "Holiday: " + h.Name
	}
	if od := offDayForDate(offDays, date); od != nil {
		return false, "Vendor off day: " + od.Title
	}
	sched, err := s.scheduleRepo.GetByStoreID(ctx, storeID)
	if err == nil {
		days := parseWorkingDays(sched.WorkingDays)
		if !isWorkingDay(days, date) {
			return false, "Closed per working schedule"
		}
	}
	return true, ""
}

// ListEvents builds the month calendar grid used by the admin dashboard.
func (s *Service) ListEvents(ctx context.Context, req *dto.ListCalendarEventsRequest) ([]dto.CalendarDayEventResponse, error) {
	monthStart := time.Date(req.Year, time.Month(req.Month), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0)

	var storeIDPtr *uint
	if req.StoreID > 0 {
		storeIDPtr = &req.StoreID
	}
	var regionPtr *string
	if req.Region != "" {
		regionPtr = &req.Region
	}

	holidays, err := s.holidayRepo.ListForRange(ctx, monthStart, monthEnd, storeIDPtr, regionPtr)
	if err != nil {
		return nil, err
	}
	var vendorIDPtr *uint
	if req.StoreID > 0 {
		vendorIDPtr = &req.StoreID
	}
	offDays, err := s.offDayRepo.ListForRange(ctx, vendorIDPtr, monthStart, monthEnd)
	if err != nil {
		return nil, err
	}

	var workingDays map[string]bool
	if req.StoreID > 0 {
		if sched, err := s.scheduleRepo.GetByStoreID(ctx, req.StoreID); err == nil {
			workingDays = parseWorkingDays(sched.WorkingDays)
		}
	}

	events := make([]dto.CalendarDayEventResponse, 0, 31)
	for d := monthStart; d.Before(monthEnd); d = d.AddDate(0, 0, 1) {
		event := dto.CalendarDayEventResponse{Date: d.Format("2006-01-02"), DayType: "working"}

		for i := range holidays {
			h := &holidays[i]
			if req.Status != "" && h.Status != req.Status {
				continue
			}
			if dateWithinRange(d, h.StartDate, h.EndDate) {
				event.DayType = "holiday"
				event.Badges = append(event.Badges, h.HolidayType+"_holiday")
				event.HolidayIDs = append(event.HolidayIDs, h.ID)
			}
		}
		for i := range offDays {
			o := &offDays[i]
			if dateWithinRange(d, o.StartDate, o.EndDate) {
				if event.DayType == "working" {
					event.DayType = "off_day"
				}
				event.Badges = append(event.Badges, o.OffType)
				event.OffDayIDs = append(event.OffDayIDs, o.ID)
			}
		}
		if event.DayType == "working" {
			if isWeekend(d) {
				event.DayType = "weekend"
			} else if workingDays != nil && !isWorkingDay(workingDays, d) {
				event.DayType = "closed"
			}
		}

		events = append(events, event)
	}
	return events, nil
}

// GetDayDetail returns the drawer detail payload for a single calendar day.
func (s *Service) GetDayDetail(ctx context.Context, dateStr string, storeID uint) (*dto.CalendarDayDetailResponse, error) {
	date, err := parseDate(dateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid date: %w", err)
	}
	dayStart := startOfDay(date)
	dayEnd := dayStart.AddDate(0, 0, 1)

	var storeIDPtr *uint
	if storeID > 0 {
		storeIDPtr = &storeID
	}
	holidays, err := s.holidayRepo.ListForRange(ctx, dayStart, dayEnd, storeIDPtr, nil)
	if err != nil {
		return nil, err
	}
	offDays, err := s.offDayRepo.ListForRange(ctx, storeIDPtr, dayStart, dayEnd)
	if err != nil {
		return nil, err
	}

	holidayResp := make([]dto.StoreHolidayResponse, 0, len(holidays))
	for i := range holidays {
		holidayResp = append(holidayResp, *toHolidayResponse(&holidays[i], storeIDsFromModel(&holidays[i])))
	}
	offDayResp := make([]dto.VendorOffDayResponse, 0, len(offDays))
	for i := range offDays {
		offDayResp = append(offDayResp, *toOffDayResponse(&offDays[i]))
	}

	dayType := "working"
	isWorking := !isWeekend(date)
	var scheduleResp *dto.StoreWorkingScheduleResponse
	if storeID > 0 {
		if sched, err := s.scheduleRepo.GetByStoreID(ctx, storeID); err == nil {
			scheduleResp = toScheduleResponse(sched)
			isWorking = isWorkingDay(parseWorkingDays(sched.WorkingDays), date)
		}
	}
	if len(holidays) > 0 {
		dayType = "holiday"
		isWorking = false
	} else if len(offDays) > 0 {
		dayType = "off_day"
		isWorking = false
	} else if isWeekend(date) {
		dayType = "weekend"
	} else if !isWorking {
		dayType = "closed"
	}

	return &dto.CalendarDayDetailResponse{
		Date:         dayStart.Format("2006-01-02"),
		DayType:      dayType,
		IsWorkingDay: isWorking,
		Holidays:     holidayResp,
		OffDays:      offDayResp,
		Schedule:     scheduleResp,
	}, nil
}

// ListUpcomingEvents merges upcoming holidays and vendor off days for the admin dashboard feed.
func (s *Service) ListUpcomingEvents(ctx context.Context, limit int) ([]dto.UpcomingEventResponse, error) {
	if limit <= 0 {
		limit = 10
	}
	now := startOfDay(time.Now())

	holidays, err := s.holidayRepo.ListUpcoming(ctx, now, limit)
	if err != nil {
		return nil, err
	}
	offDays, err := s.offDayRepo.ListUpcoming(ctx, now, limit)
	if err != nil {
		return nil, err
	}

	events := make([]dto.UpcomingEventResponse, 0, len(holidays)+len(offDays))
	for i := range holidays {
		h := &holidays[i]
		events = append(events, dto.UpcomingEventResponse{
			Type: "holiday", ID: h.ID, Title: h.Name,
			StartDate: h.StartDate, EndDate: h.EndDate, Scope: h.HolidayType,
		})
	}
	for i := range offDays {
		o := &offDays[i]
		events = append(events, dto.UpcomingEventResponse{
			Type: "off_day", ID: o.ID, Title: o.Title,
			StartDate: o.StartDate, EndDate: o.EndDate, Scope: o.OffType,
		})
	}

	sortEventsByStartDate(events)
	if len(events) > limit {
		events = events[:limit]
	}
	return events, nil
}

func sortEventsByStartDate(events []dto.UpcomingEventResponse) {
	for i := 1; i < len(events); i++ {
		for j := i; j > 0 && events[j].StartDate.Before(events[j-1].StartDate); j-- {
			events[j], events[j-1] = events[j-1], events[j]
		}
	}
}

// ListStoresStatusToday reports open/closed status for every active store today.
func (s *Service) ListStoresStatusToday(ctx context.Context) ([]dto.StoreStatusTodayResponse, error) {
	stores, err := s.storeRepo.ListActiveLite(ctx)
	if err != nil {
		return nil, err
	}
	today := startOfDay(time.Now())
	holidaysToday, err := s.holidayRepo.ListForRange(ctx, today, today.AddDate(0, 0, 1), nil, nil)
	if err != nil {
		return nil, err
	}
	offDaysToday, err := s.offDayRepo.ListForRange(ctx, nil, today, today.AddDate(0, 0, 1))
	if err != nil {
		return nil, err
	}

	resp := make([]dto.StoreStatusTodayResponse, 0, len(stores))
	for _, store := range stores {
		open, reason := s.isStoreOpenOnDate(ctx, store.ID, today, holidaysToday, offDaysToday)
		resp = append(resp, dto.StoreStatusTodayResponse{
			StoreID:        store.ID,
			StoreName:      store.Name,
			IsWorkingToday: open,
			Reason:         reason,
		})
	}
	return resp, nil
}

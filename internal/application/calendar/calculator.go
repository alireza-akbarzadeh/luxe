package calendar

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

// maxCalculatorIterations bounds the day-by-day advance loop so a
// misconfigured vendor (e.g. permanently closed) cannot hang the calculator.
const maxCalculatorIterations = 120

// CalculateEarliestDeliveryInput is the delivery calculator's input.
//
// Pipeline (in order, exactly as required by the store working calendar spec):
//
//	order date
//	→ add vendor processing time
//	→ vendor working schedule check
//	→ skip vendor off days
//	→ skip store holidays
//	→ skip national holidays
//	→ skip regional holidays
//	→ warehouse/store working days
//	→ courier availability (skip weekends unless express rule)
//	→ first valid processing day
//	→ add shipping duration days
//	→ earliest delivery date
type CalculateEarliestDeliveryInput struct {
	StoreID        uint
	VendorID       uint
	City           string
	Region         string
	OrderDate      time.Time
	ShippingMethod string
	ShippingDays   int
}

// TimelineStep is one stage in the delivery calculator's pipeline, used by the
// admin UI to visualize how the earliest delivery date was derived.
type TimelineStep struct {
	Step  string    `json:"step"`
	Date  time.Time `json:"date"`
	Days  int       `json:"days"`
	Label string    `json:"label"`
}

// CalculateEarliestDeliveryResult is the delivery calculator's output.
type CalculateEarliestDeliveryResult struct {
	ProcessingStart time.Time
	VendorAvailable bool
	SkippedHolidays []string
	SkippedWeekends []string
	DeliveryDate    time.Time
	DelayReason     string
	Timeline        []TimelineStep
}

// resolution accumulates per-category skip counts/dates while the day-by-day
// loop advances toward the first valid processing day.
type resolution struct {
	skippedHolidayDates []string // vendor schedule + vendor off day + any holiday closures
	skippedWeekendDates []string
	vendorScheduleDays  int
	vendorOffDayDays    int
	storeHolidayDays    int
	nationalHolidayDays int
	regionalHolidayDays int
	storeScheduleDays   int
	weekendDays         int
}

// CalculateEarliestDelivery computes the earliest possible delivery date for an
// order, walking day-by-day past vendor/store closures, holidays, and weekends
// per the enabled delivery calendar rules.
func (s *Service) CalculateEarliestDelivery(ctx context.Context, input CalculateEarliestDeliveryInput) (*CalculateEarliestDeliveryResult, error) {
	vendorID := input.VendorID
	if vendorID == 0 {
		vendorID = input.StoreID
	}

	rulesEnabled, err := s.ruleRepo.EnabledMap(ctx)
	if err != nil {
		return nil, err
	}

	vendorSchedule, err := s.getScheduleOrDefault(ctx, vendorID)
	if err != nil {
		return nil, err
	}
	var storeSchedule *models.StoreWorkingSchedule
	if vendorID != input.StoreID {
		storeSchedule, err = s.getScheduleOrDefault(ctx, input.StoreID)
		if err != nil {
			return nil, err
		}
	}

	processingStart := input.OrderDate.Add(time.Duration(vendorSchedule.ProcessingLeadHours) * time.Hour)

	windowStart := startOfDay(processingStart)
	windowEnd := windowStart.AddDate(0, 0, maxCalculatorIterations+1)

	var regionPtr *string
	if input.Region != "" {
		regionPtr = &input.Region
	}
	storeIDForHolidays := input.StoreID
	holidays, err := s.holidayRepo.ListForRange(ctx, windowStart, windowEnd, &storeIDForHolidays, regionPtr)
	if err != nil {
		return nil, err
	}
	offDays, err := s.offDayRepo.ListForRange(ctx, &vendorID, windowStart, windowEnd)
	if err != nil {
		return nil, err
	}

	vendorDays := parseWorkingDays(vendorSchedule.WorkingDays)
	var storeDays map[string]bool
	if storeSchedule != nil {
		storeDays = parseWorkingDays(storeSchedule.WorkingDays)
	}

	isPickup := input.ShippingMethod == constants.ShippingMethodPickup
	isExpress := input.ShippingMethod == constants.ShippingMethodExpress
	pickupIgnoresRules := isPickup && rulesEnabled[constants.RuleKeyPickupIgnoreDeliveryRules]

	res := &resolution{}
	cursor := windowStart
	vendorAvailable := true
	delayReason := ""

	iterations := 0
	for {
		if iterations >= maxCalculatorIterations {
			vendorAvailable = false
			delayReason = fmt.Sprintf("no valid processing day found within %d days; check vendor schedule and closures", maxCalculatorIterations)
			break
		}
		iterations++

		if skipDay(cursor, vendorDays, storeDays, holidays, offDays, rulesEnabled, isExpress, pickupIgnoresRules, res) {
			cursor = cursor.AddDate(0, 0, 1)
			continue
		}
		break
	}
	validProcessingDay := cursor

	timeline := buildTimeline(input, vendorSchedule, processingStart, validProcessingDay, res, isPickup, vendorAvailable, delayReason)

	deliveryDate := validProcessingDay
	if !isPickup {
		deliveryDate = validProcessingDay.AddDate(0, 0, input.ShippingDays)
	}
	timeline = append(timeline, TimelineStep{
		Step:  "earliest_delivery_date",
		Date:  deliveryDate,
		Days:  0,
		Label: "Earliest delivery date: " + deliveryDate.Format("2006-01-02"),
	})

	return &CalculateEarliestDeliveryResult{
		ProcessingStart: processingStart,
		VendorAvailable: vendorAvailable,
		SkippedHolidays: res.skippedHolidayDates,
		SkippedWeekends: res.skippedWeekendDates,
		DeliveryDate:    deliveryDate,
		DelayReason:     delayReason,
		Timeline:        timeline,
	}, nil
}

// skipDay reports whether the candidate date must be skipped, and records
// which category caused the skip. Checks run in the exact spec order.
func skipDay(
	date time.Time,
	vendorDays, storeDays map[string]bool,
	holidays []models.StoreHoliday,
	offDays []models.VendorOffDay,
	rulesEnabled map[string]bool,
	isExpress, pickupIgnoresRules bool,
	res *resolution,
) bool {
	dateStr := date.Format("2006-01-02")

	// 1. vendor working schedule check
	if rulesEnabled[constants.RuleKeyVendorClosedSkip] && !isWorkingDay(vendorDays, date) {
		res.vendorScheduleDays++
		res.skippedHolidayDates = append(res.skippedHolidayDates, dateStr)
		return true
	}

	// 2. skip vendor off days (maintenance/emergency close gated by maintenance_block; other
	// off day types such as vacation/personal_leave always block since there is no dedicated toggle).
	if od := offDayForDate(offDays, date); od != nil {
		toggleGated := od.OffType == constants.OffDayTypeMaintenance || od.OffType == constants.OffDayTypeEmergencyClose
		if !toggleGated || rulesEnabled[constants.RuleKeyMaintenanceBlock] {
			res.vendorOffDayDays++
			res.skippedHolidayDates = append(res.skippedHolidayDates, dateStr)
			return true
		}
	}

	// 3. skip store holidays (holiday_type = store/vendor closures)
	if rulesEnabled[constants.RuleKeyVendorClosedSkip] {
		if h := holidayForDate(holidays, date, constants.HolidayTypeStore, constants.HolidayTypeVendor); h != nil {
			res.storeHolidayDays++
			res.skippedHolidayDates = append(res.skippedHolidayDates, dateStr)
			return true
		}
	}

	// 4. skip national holidays
	if rulesEnabled[constants.RuleKeyNationalHolidaySkip] {
		if h := holidayForDate(holidays, date, constants.HolidayTypeNational); h != nil {
			res.nationalHolidayDays++
			res.skippedHolidayDates = append(res.skippedHolidayDates, dateStr)
			return true
		}
	}

	// 5. skip regional holidays
	if rulesEnabled[constants.RuleKeyRegionalHolidaySkip] {
		if h := holidayForDate(holidays, date, constants.HolidayTypeRegional); h != nil {
			res.regionalHolidayDays++
			res.skippedHolidayDates = append(res.skippedHolidayDates, dateStr)
			return true
		}
	}

	// 6. warehouse/store working days (only meaningful when vendor and store differ)
	if storeDays != nil && !isWorkingDay(storeDays, date) {
		res.storeScheduleDays++
		res.skippedHolidayDates = append(res.skippedHolidayDates, dateStr)
		return true
	}

	// 7. courier availability — skip weekends unless express ignores them or shipping is pickup
	if !pickupIgnoresRules && isWeekend(date) {
		weekendAllowed := isExpress && rulesEnabled[constants.RuleKeyExpressIgnoreWeekend]
		if !weekendAllowed {
			res.weekendDays++
			res.skippedWeekendDates = append(res.skippedWeekendDates, dateStr)
			return true
		}
	}

	return false
}

func (s *Service) getScheduleOrDefault(ctx context.Context, storeID uint) (*models.StoreWorkingSchedule, error) {
	sched, err := s.scheduleRepo.GetByStoreID(ctx, storeID)
	if err == nil {
		return sched, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	days, _ := encodeWorkingDays(defaultWorkingDays())
	return &models.StoreWorkingSchedule{
		StoreID:             storeID,
		Timezone:            "UTC",
		WorkingDays:         days,
		OpenTime:            "09:00",
		CloseTime:           "18:00",
		ProcessingLeadHours: 24,
	}, nil
}

func holidayForDate(holidays []models.StoreHoliday, date time.Time, types ...string) *models.StoreHoliday {
	typeSet := make(map[string]struct{}, len(types))
	for _, t := range types {
		typeSet[t] = struct{}{}
	}
	for i := range holidays {
		h := &holidays[i]
		if _, ok := typeSet[h.HolidayType]; !ok {
			continue
		}
		if dateWithinRange(date, h.StartDate, h.EndDate) {
			return h
		}
	}
	return nil
}

func offDayForDate(offDays []models.VendorOffDay, date time.Time) *models.VendorOffDay {
	for i := range offDays {
		o := &offDays[i]
		if dateWithinRange(date, o.StartDate, o.EndDate) {
			return o
		}
	}
	return nil
}

func dateWithinRange(date, start, end time.Time) bool {
	day := startOfDay(date)
	return !day.Before(startOfDay(start)) && day.Before(startOfDay(end).AddDate(0, 0, 1))
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func isWeekend(t time.Time) bool {
	return t.Weekday() == time.Saturday || t.Weekday() == time.Sunday
}

func buildTimeline(
	input CalculateEarliestDeliveryInput,
	vendorSchedule *models.StoreWorkingSchedule,
	processingStart, validDay time.Time,
	res *resolution,
	isPickup, vendorAvailable bool,
	delayReason string,
) []TimelineStep {
	timeline := []TimelineStep{
		{Step: "order_date", Date: input.OrderDate, Days: 0, Label: "Order placed"},
		{
			Step:  "vendor_processing",
			Date:  processingStart,
			Days:  vendorSchedule.ProcessingLeadHours,
			Label: fmt.Sprintf("Vendor processing time added (%dh)", vendorSchedule.ProcessingLeadHours),
		},
	}

	appendIfSkipped := func(step string, days int, label string) {
		if days > 0 {
			timeline = append(timeline, TimelineStep{Step: step, Date: validDay, Days: days, Label: label})
		}
	}

	appendIfSkipped("vendor_schedule_check", res.vendorScheduleDays, fmt.Sprintf("Skipped %d day(s) vendor was closed per working schedule", res.vendorScheduleDays))
	appendIfSkipped("skip_vendor_off_days", res.vendorOffDayDays, fmt.Sprintf("Skipped %d day(s) for vendor off days", res.vendorOffDayDays))
	appendIfSkipped("skip_store_holidays", res.storeHolidayDays, fmt.Sprintf("Skipped %d day(s) for store/vendor holidays", res.storeHolidayDays))
	appendIfSkipped("skip_national_holidays", res.nationalHolidayDays, fmt.Sprintf("Skipped %d day(s) for national holidays", res.nationalHolidayDays))
	appendIfSkipped("skip_regional_holidays", res.regionalHolidayDays, fmt.Sprintf("Skipped %d day(s) for regional holidays", res.regionalHolidayDays))
	appendIfSkipped("warehouse_store_working_days", res.storeScheduleDays, fmt.Sprintf("Skipped %d day(s) store/warehouse was closed", res.storeScheduleDays))

	courierLabel := fmt.Sprintf("Skipped %d weekend day(s) for courier availability", res.weekendDays)
	if isPickup {
		courierLabel = "Pickup — courier availability step bypassed"
	}
	timeline = append(timeline, TimelineStep{Step: "courier_availability", Date: validDay, Days: res.weekendDays, Label: courierLabel})

	firstValidLabel := "First valid processing day: " + validDay.Format("2006-01-02")
	if !vendorAvailable {
		firstValidLabel = delayReason
	}
	timeline = append(timeline, TimelineStep{Step: "first_valid_processing_day", Date: validDay, Days: 0, Label: firstValidLabel})

	shippingDays := input.ShippingDays
	shippingLabel := fmt.Sprintf("Shipping duration added (%d day(s))", shippingDays)
	if isPickup {
		shippingDays = 0
		shippingLabel = "Pickup — no shipping duration added"
	}
	timeline = append(timeline, TimelineStep{Step: "add_shipping_duration", Date: validDay, Days: shippingDays, Label: shippingLabel})

	return timeline
}

// ─── Simulate wraps the calculator with DTO conversion for the HTTP layer ──

// Simulate runs the delivery calculator for the admin simulator UI.
func (s *Service) Simulate(ctx context.Context, req *dto.SimulateDeliveryRequest) (*dto.SimulateDeliveryResponse, error) {
	orderDate, err := parseDate(req.OrderDate)
	if err != nil {
		return nil, utils.ErrBadRequest("invalid order_date; use RFC3339 or YYYY-MM-DD")
	}

	input := CalculateEarliestDeliveryInput{
		StoreID:        req.StoreID,
		OrderDate:      orderDate,
		ShippingMethod: req.ShippingMethod,
		ShippingDays:   req.ShippingDays,
	}
	if req.VendorID != nil {
		input.VendorID = *req.VendorID
	}
	if req.City != nil {
		input.City = *req.City
	}
	if req.Region != nil {
		input.Region = *req.Region
	}

	result, err := s.CalculateEarliestDelivery(ctx, input)
	if err != nil {
		return nil, err
	}

	timeline := make([]dto.DeliveryTimelineStep, 0, len(result.Timeline))
	for _, step := range result.Timeline {
		timeline = append(timeline, dto.DeliveryTimelineStep{
			Step:  step.Step,
			Date:  step.Date,
			Days:  step.Days,
			Label: step.Label,
		})
	}

	return &dto.SimulateDeliveryResponse{
		ProcessingStart: result.ProcessingStart,
		VendorAvailable: result.VendorAvailable,
		SkippedHolidays: result.SkippedHolidays,
		SkippedWeekends: result.SkippedWeekends,
		DeliveryDate:    result.DeliveryDate,
		DelayReason:     result.DelayReason,
		Timeline:        timeline,
	}, nil
}

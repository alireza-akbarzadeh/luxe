// Package calendar holds domain validation rules for store working schedules,
// holidays, vendor off days, and delivery calendar rule toggles.
package calendar

import (
	"errors"
	"strings"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
)

var (
	ErrInvalidName        = errors.New("name is required")
	ErrInvalidDateRange   = errors.New("end date must be on or after start date")
	ErrInvalidHolidayType = errors.New("invalid holiday type")
	ErrInvalidApplyTo     = errors.New("invalid apply_to scope")
	ErrInvalidOffType     = errors.New("invalid off day type")
	ErrInvalidStatus      = errors.New("invalid status")
	ErrInvalidTimeOfDay   = errors.New("time must be in HH:MM 24h format")
)

var validHolidayTypes = map[string]struct{}{
	constants.HolidayTypeNational: {},
	constants.HolidayTypeRegional: {},
	constants.HolidayTypeStore:    {},
	constants.HolidayTypeVendor:   {},
}

var validApplyTo = map[string]struct{}{
	constants.HolidayApplyToAll:    {},
	constants.HolidayApplyToStores: {},
	constants.HolidayApplyToVendor: {},
	constants.HolidayApplyToRegion: {},
}

var validOffTypes = map[string]struct{}{
	constants.OffDayTypeVacation:       {},
	constants.OffDayTypeInventoryCount: {},
	constants.OffDayTypeMaintenance:    {},
	constants.OffDayTypeEmergencyClose: {},
	constants.OffDayTypePersonalLeave:  {},
}

var validStatuses = map[string]struct{}{
	constants.CalendarStatusDraft:     {},
	constants.CalendarStatusPublished: {},
}

// Service holds calendar domain rules.
type Service struct{}

// NewService creates a calendar domain service.
func NewService() *Service { return &Service{} }

// ValidateName ensures a display name is present.
func (s *Service) ValidateName(name string) error {
	if strings.TrimSpace(name) == "" {
		return ErrInvalidName
	}
	return nil
}

// ValidateDateRange ensures the end date is not before the start date.
func (s *Service) ValidateDateRange(start, end time.Time) error {
	if end.Before(start) {
		return ErrInvalidDateRange
	}
	return nil
}

// ValidateHolidayType ensures the holiday scope is a known value.
func (s *Service) ValidateHolidayType(holidayType string) error {
	if _, ok := validHolidayTypes[holidayType]; !ok {
		return ErrInvalidHolidayType
	}
	return nil
}

// ValidateApplyTo ensures the apply-to target is a known value.
func (s *Service) ValidateApplyTo(applyTo string) error {
	if _, ok := validApplyTo[applyTo]; !ok {
		return ErrInvalidApplyTo
	}
	return nil
}

// ValidateOffType ensures the vendor off day type is a known value.
func (s *Service) ValidateOffType(offType string) error {
	if _, ok := validOffTypes[offType]; !ok {
		return ErrInvalidOffType
	}
	return nil
}

// ValidateStatus ensures the draft/published status is a known value.
func (s *Service) ValidateStatus(status string) error {
	if _, ok := validStatuses[status]; !ok {
		return ErrInvalidStatus
	}
	return nil
}

// ValidateTimeOfDay ensures a HH:MM string is well-formed (24h clock).
func (s *Service) ValidateTimeOfDay(value string) error {
	if value == "" {
		return nil
	}
	if _, err := time.Parse("15:04", value); err != nil {
		return ErrInvalidTimeOfDay
	}
	return nil
}

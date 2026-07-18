// Package calendar implements the store working calendar and delivery
// availability module: working schedules, holidays, vendor off days,
// admin-togglable delivery rules, and the earliest-delivery-date calculator.
package calendar

import (
	domaincalendar "github.com/alireza-akbarzadeh/luxe/internal/domain/calendar"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"gorm.io/gorm"
)

// Service handles calendar HTTP-oriented use cases (DTO mapping, validation, orchestration).
type Service struct {
	domain *domaincalendar.Service

	scheduleRepo *postgres.StoreWorkingScheduleRepository
	holidayRepo  *postgres.StoreHolidayRepository
	offDayRepo   *postgres.VendorOffDayRepository
	ruleRepo     *postgres.DeliveryCalendarRuleRepository
	storeRepo    *postgres.StoreRepository
}

// NewService wires calendar application use cases.
func NewService(db *gorm.DB) *Service {
	return &Service{
		domain:       domaincalendar.NewService(),
		scheduleRepo: postgres.NewStoreWorkingScheduleRepository(db),
		holidayRepo:  postgres.NewStoreHolidayRepository(db),
		offDayRepo:   postgres.NewVendorOffDayRepository(db),
		ruleRepo:     postgres.NewDeliveryCalendarRuleRepository(db),
		storeRepo:    postgres.NewStoreRepository(db),
	}
}

package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// StoreWorkingSchedule defines a store's open/closed days, hours, and processing
// lead times used by the delivery calculator to compute earliest delivery dates.
type StoreWorkingSchedule struct {
	ID                  uint           `gorm:"primaryKey" json:"id"`
	StoreID             uint           `gorm:"not null;uniqueIndex" json:"store_id"`
	Timezone            string         `gorm:"size:64;not null;default:'UTC'" json:"timezone"`
	WorkingDays         datatypes.JSON `gorm:"type:jsonb;not null" json:"working_days"`
	OpenTime            string         `gorm:"size:5;not null;default:'09:00'" json:"open_time"`
	CloseTime           string         `gorm:"size:5;not null;default:'18:00'" json:"close_time"`
	BreakStart          *string        `gorm:"size:5" json:"break_start,omitempty"`
	BreakEnd            *string        `gorm:"size:5" json:"break_end,omitempty"`
	MaxOrdersPerDay     int            `gorm:"not null;default:0" json:"max_orders_per_day"`
	MaxDeliveriesPerDay int            `gorm:"not null;default:0" json:"max_deliveries_per_day"`
	PrepLeadHours       int            `gorm:"not null;default:0" json:"prep_lead_hours"`
	ProcessingLeadHours int            `gorm:"not null;default:24" json:"processing_lead_hours"`
	DeliveryBufferHours int            `gorm:"not null;default:0" json:"delivery_buffer_hours"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`

	Store *Store `gorm:"foreignKey:StoreID" json:"store,omitempty"`
}

func (StoreWorkingSchedule) TableName() string {
	return "store_working_schedules"
}

// StoreHoliday is a closure window (national/regional/store/vendor scoped) that the
// delivery calculator skips over when computing the earliest processing day.
type StoreHoliday struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	Name           string         `gorm:"size:255;not null" json:"name"`
	Description    string         `gorm:"type:text;not null;default:''" json:"description"`
	HolidayType    string         `gorm:"size:20;not null;default:'store'" json:"holiday_type"`
	StartDate      time.Time      `gorm:"not null" json:"start_date"`
	EndDate        time.Time      `gorm:"not null" json:"end_date"`
	IsRecurring    bool           `gorm:"not null;default:false" json:"is_recurring"`
	RecurrenceRule *string        `gorm:"size:50" json:"recurrence_rule,omitempty"`
	ApplyTo        string         `gorm:"size:20;not null;default:'stores'" json:"apply_to"`
	VendorID       *uint          `gorm:"index" json:"vendor_id,omitempty"`
	Region         *string        `gorm:"size:100" json:"region,omitempty"`
	Priority       int            `gorm:"not null;default:0" json:"priority"`
	Status         string         `gorm:"size:20;not null;default:'draft'" json:"status"`
	Notes          string         `gorm:"type:text;not null;default:''" json:"notes"`
	CreatedBy      *uint          `json:"created_by,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	Stores []Store `gorm:"many2many:store_holiday_stores;joinForeignKey:HolidayID;joinReferences:StoreID" json:"stores,omitempty"`
}

func (StoreHoliday) TableName() string {
	return "store_holidays"
}

// StoreHolidayStore is the join row scoping a holiday to a specific store
// when StoreHoliday.ApplyTo == "stores".
type StoreHolidayStore struct {
	HolidayID uint `gorm:"primaryKey;column:holiday_id" json:"holiday_id"`
	StoreID   uint `gorm:"primaryKey;column:store_id" json:"store_id"`
}

func (StoreHolidayStore) TableName() string {
	return "store_holiday_stores"
}

// VendorOffDay is a vendor/store-initiated closure (vacation, maintenance, etc.)
// distinct from platform-managed holidays.
type VendorOffDay struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	VendorID  uint           `gorm:"not null;index" json:"vendor_id"`
	Title     string         `gorm:"size:255;not null" json:"title"`
	OffType   string         `gorm:"size:30;not null;default:'vacation'" json:"off_type"`
	StartDate time.Time      `gorm:"not null" json:"start_date"`
	EndDate   time.Time      `gorm:"not null" json:"end_date"`
	Notes     string         `gorm:"type:text;not null;default:''" json:"notes"`
	Status    string         `gorm:"size:20;not null;default:'draft'" json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Vendor *Store `gorm:"foreignKey:VendorID" json:"vendor,omitempty"`
}

func (VendorOffDay) TableName() string {
	return "vendor_off_days"
}

// DeliveryCalendarRule is an admin-togglable rule the delivery calculator consults
// (e.g. whether express shipments ignore weekend skips).
type DeliveryCalendarRule struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	RuleKey     string    `gorm:"size:50;not null;uniqueIndex" json:"rule_key"`
	Enabled     bool      `gorm:"not null;default:true" json:"enabled"`
	Label       string    `gorm:"size:255;not null" json:"label"`
	Description string    `gorm:"type:text;not null;default:''" json:"description"`
	SortOrder   int       `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (DeliveryCalendarRule) TableName() string {
	return "delivery_calendar_rules"
}

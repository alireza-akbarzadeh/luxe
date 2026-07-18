package constants

// Store holiday types — scope of a closure window.
const (
	HolidayTypeNational = "national"
	HolidayTypeRegional = "regional"
	HolidayTypeStore    = "store"
	HolidayTypeVendor   = "vendor"
)

// Store holiday apply-to targets.
const (
	HolidayApplyToAll    = "all"
	HolidayApplyToStores = "stores"
	HolidayApplyToVendor = "vendor"
	HolidayApplyToRegion = "region"
)

// Store holiday / vendor off day statuses.
const (
	CalendarStatusDraft     = "draft"
	CalendarStatusPublished = "published"
)

// Vendor off day types.
const (
	OffDayTypeVacation       = "vacation"
	OffDayTypeInventoryCount = "inventory_count"
	OffDayTypeMaintenance    = "maintenance"
	OffDayTypeEmergencyClose = "emergency_close"
	OffDayTypePersonalLeave  = "personal_leave"
)

// Delivery calendar rule keys — admin-togglable rules consulted by the delivery calculator.
const (
	RuleKeyVendorClosedSkip          = "vendor_closed_skip"
	RuleKeyNationalHolidaySkip       = "national_holiday_skip"
	RuleKeyRegionalHolidaySkip       = "regional_holiday_skip"
	RuleKeyCapacityFullNextDay       = "capacity_full_next_day"
	RuleKeyMaintenanceBlock          = "maintenance_block"
	RuleKeyExpressIgnoreWeekend      = "express_ignore_weekend"
	RuleKeyPickupIgnoreDeliveryRules = "pickup_ignore_delivery_rules"
)

// Shipping methods accepted by the delivery calculator/simulator.
const (
	ShippingMethodStandard = "standard"
	ShippingMethodExpress  = "express"
	ShippingMethodPickup   = "pickup"
)

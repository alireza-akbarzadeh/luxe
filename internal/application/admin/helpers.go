package admin

import (
	"math"
)

func dashboardPeriodDays(period string) int {
	switch period {
	case "7d":
		return 7
	case "90d":
		return 90
	default:
		return 30
	}
}

func dashboardKPI(current, previous float64) KPI {
	change := 0.0
	if previous > 0 {
		change = ((current - previous) / previous) * 100
	} else if current > 0 {
		change = 100
	}
	return KPI{
		Value:         current,
		PreviousValue: previous,
		ChangePercent: math.Round(change*10) / 10,
	}
}

// KPI mirrors dto.AdminDashboardKPI for application-layer calculations.
type KPI struct {
	Value         float64
	PreviousValue float64
	ChangePercent float64
}

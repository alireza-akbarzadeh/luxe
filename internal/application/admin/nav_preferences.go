package admin

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

// GetDashboardHealth returns platform health metrics for the admin dashboard.
func (q *Queries) GetDashboardHealth(ctx context.Context) (*dto.AdminDashboardHealth, error) {
	platform, err := q.GetStats(ctx)
	if err != nil {
		return nil, err
	}
	health, err := buildPlatformHealth(ctx, q, *platform)
	if err != nil {
		return nil, err
	}
	return &health, nil
}

// ExportDashboardCSV builds a CSV export for the dashboard revenue series.
func (q *Queries) ExportDashboardCSV(ctx context.Context, filters dto.AdminDashboardExportFilters) ([]byte, error) {
	overview, err := q.GetDashboardOverview(ctx, dto.AdminDashboardFilters{Period: filters.Period})
	if err != nil {
		return nil, err
	}
	if filters.Period == "" {
		filters.Period = overview.Period
	}

	buf := &bytes.Buffer{}
	writer := csv.NewWriter(buf)
	_ = writer.Write([]string{"date", "revenue", "orders", "avg_order_value"})

	for _, row := range overview.RevenueSeries {
		aov := 0.0
		if row.Orders > 0 {
			aov = row.Revenue / float64(row.Orders)
		}
		_ = writer.Write([]string{
			row.Date,
			fmt.Sprintf("%.2f", row.Revenue),
			strconv.FormatInt(row.Orders, 10),
			fmt.Sprintf("%.2f", aov),
		})
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, utils.ErrInternal(err)
	}

	return buf.Bytes(), nil
}

// NavPreferences returns navigation preferences for an admin user.
func (q *Queries) NavPreferences(ctx context.Context, userID uint) (*dto.AdminNavPreferencesResponse, error) {
	return q.navRepo.GetByUserID(ctx, userID)
}

// SaveNavPreferences upserts navigation preferences for an admin user.
func (c *Commands) SaveNavPreferences(ctx context.Context, userID uint, req dto.UpdateAdminNavPreferencesRequest) (*dto.AdminNavPreferencesResponse, error) {
	if req.Favorites == nil {
		req.Favorites = []string{}
	}
	if req.Recent == nil {
		req.Recent = []dto.AdminNavRecentPage{}
	}
	if len(req.Recent) > 10 {
		req.Recent = req.Recent[:10]
	}
	return c.navRepo.Upsert(ctx, userID, req)
}

// AppendRecentPage records a recently visited admin page for a user.
func (c *Commands) AppendRecentPage(ctx context.Context, userID uint, page dto.AdminNavRecentPage) (*dto.AdminNavPreferencesResponse, error) {
	current, err := c.navRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	page.VisitedAt = time.Now().UTC()
	filtered := make([]dto.AdminNavRecentPage, 0, len(current.Recent)+1)
	for _, item := range current.Recent {
		if item.Href != page.Href {
			filtered = append(filtered, item)
		}
	}
	filtered = append([]dto.AdminNavRecentPage{page}, filtered...)
	if len(filtered) > 10 {
		filtered = filtered[:10]
	}

	return c.navRepo.Upsert(ctx, userID, dto.UpdateAdminNavPreferencesRequest{
		Favorites: current.Favorites,
		Recent:    filtered,
	})
}

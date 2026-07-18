package dto

import "time"

// AdminNavRecentPage is a recently visited admin page entry.
type AdminNavRecentPage struct {
	Href      string    `json:"href" validate:"required"`
	Label     string    `json:"label" validate:"required"`
	VisitedAt time.Time `json:"visited_at"`
}

// AdminNavPreferencesResponse is the persisted nav preferences for an admin user.
type AdminNavPreferencesResponse struct {
	Favorites []string             `json:"favorites"`
	Recent    []AdminNavRecentPage `json:"recent"`
}

// UpdateAdminNavPreferencesRequest updates favorites and/or recent pages.
type UpdateAdminNavPreferencesRequest struct {
	Favorites []string             `json:"favorites"`
	Recent    []AdminNavRecentPage `json:"recent"`
}

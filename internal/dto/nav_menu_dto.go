package dto

import (
	"encoding/json"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

type NavItemResponse struct {
	Type     string         `json:"type"`
	Label    string         `json:"label"`
	Href     *string        `json:"href,omitempty"`
	Badge    *string        `json:"badge,omitempty"`
	ViewAll  *ViewAll       `json:"viewAll,omitempty"`
	Columns  []Column       `json:"columns,omitempty"`
	Featured []FeaturedItem `json:"featured,omitempty"`
}

type ViewAll struct {
	Label string `json:"label"`
	Href  string `json:"href"`
}

type Column struct {
	Title string `json:"title"`
	Links []Link `json:"links"`
}

type Link struct {
	Title string `json:"title"`
	Href  string `json:"href"`
}

type FeaturedItem struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Href        string  `json:"href"`
	Image       string  `json:"image"`
	Badge       *string `json:"badge,omitempty"`
}

type UpsertNavMenuRequest struct {
	Label    string         `json:"label" binding:"required"`
	Type     string         `json:"type" binding:"required,oneof=mega link"`
	Href     *string        `json:"href,omitempty"`
	Badge    *string        `json:"badge,omitempty"`
	ViewAll  *ViewAll       `json:"viewAll,omitempty"`
	Columns  []Column       `json:"columns,omitempty"`
	Featured []FeaturedItem `json:"featured,omitempty"`
	Order    int            `json:"order"`
}

// ToNavItemResponse Helper to convert model to DTO (flatten JSONB)
func ToNavItemResponse(m *models.NavMenu) (*NavItemResponse, error) {
	resp := &NavItemResponse{
		Type:  m.Type,
		Label: m.Label,
		Href:  m.Href,
		Badge: m.Badge,
	}
	if m.ViewAll != nil {
		var va ViewAll
		if err := json.Unmarshal(m.ViewAll, &va); err == nil {
			resp.ViewAll = &va
		}
	}
	if m.Columns != nil {
		var cols []Column
		if err := json.Unmarshal(m.Columns, &cols); err == nil {
			resp.Columns = cols
		}
	}
	if m.Featured != nil {
		var feat []FeaturedItem
		if err := json.Unmarshal(m.Featured, &feat); err == nil {
			resp.Featured = feat
		}
	}
	return resp, nil
}

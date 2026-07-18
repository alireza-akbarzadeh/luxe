package dto

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/i18n"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/datatypes"
)

type NavItemResponse struct {
	ID        uint              `json:"id"`
	Type      string            `json:"type"`
	Label     string            `json:"label"`
	LabelI18n i18n.LocalizedMap `json:"labelI18n,omitempty"`
	Href      *string           `json:"href,omitempty"`
	Badge     *string           `json:"badge,omitempty"`
	BadgeI18n i18n.LocalizedMap `json:"badgeI18n,omitempty"`
	Order     int               `json:"order"`
	ViewAll   *ViewAll          `json:"viewAll,omitempty"`
	Columns   []Column          `json:"columns,omitempty"`
	Featured  []FeaturedItem    `json:"featured,omitempty"`
}

type ReorderNavMenuItem struct {
	ID    uint `json:"id" binding:"required"`
	Order int  `json:"order" binding:"gte=0"`
}

type ReorderNavMenusRequest struct {
	Items []ReorderNavMenuItem `json:"items" binding:"required,min=1,dive"`
}

type ViewAll struct {
	Label     string            `json:"label"`
	LabelI18n i18n.LocalizedMap `json:"labelI18n,omitempty"`
	Href      string            `json:"href"`
}

type Column struct {
	Title     string            `json:"title"`
	TitleI18n i18n.LocalizedMap `json:"titleI18n,omitempty"`
	Links     []Link            `json:"links"`
}

type Link struct {
	Title     string            `json:"title"`
	TitleI18n i18n.LocalizedMap `json:"titleI18n,omitempty"`
	Href      string            `json:"href"`
}

type FeaturedItem struct {
	Title           string            `json:"title"`
	TitleI18n       i18n.LocalizedMap `json:"titleI18n,omitempty"`
	Description     string            `json:"description"`
	DescriptionI18n i18n.LocalizedMap `json:"descriptionI18n,omitempty"`
	Href            string            `json:"href"`
	Image           string            `json:"image"`
	Badge           *string           `json:"badge,omitempty"`
	BadgeI18n       i18n.LocalizedMap `json:"badgeI18n,omitempty"`
}

type UpsertNavMenuRequest struct {
	Label     string            `json:"label" binding:"required"`
	LabelI18n i18n.LocalizedMap `json:"labelI18n,omitempty"`
	Type      string            `json:"type" binding:"required,oneof=mega link"`
	Href      *string           `json:"href,omitempty"`
	Badge     *string           `json:"badge,omitempty"`
	BadgeI18n i18n.LocalizedMap `json:"badgeI18n,omitempty"`
	ViewAll   *ViewAll          `json:"viewAll,omitempty"`
	Columns   []Column          `json:"columns,omitempty"`
	Featured  []FeaturedItem    `json:"featured,omitempty"`
	Order     int               `json:"order"`
}

// ToNavItemResponse converts a nav menu model to a locale-aware API response.
func ToNavItemResponse(ctx context.Context, m *models.NavMenu) (*NavItemResponse, error) {
	labelMap := decodeLabelI18n(m.LabelI18n, m.Label)
	badgeMap := decodeBadgeI18n(m.BadgeI18n, m.Badge)

	resp := &NavItemResponse{
		ID:        m.ID,
		Type:      m.Type,
		Label:     labelMap.Resolve(ctx, m.Label),
		LabelI18n: labelMap,
		Href:      m.Href,
		Order:     m.SortOrder,
		Columns:   decodeColumns(ctx, m.Columns),
		Featured:  decodeFeatured(ctx, m.Featured),
	}

	if m.Badge != nil || len(badgeMap) > 0 {
		resolved := badgeMap.Resolve(ctx, derefString(m.Badge))
		resp.Badge = &resolved
	}
	resp.BadgeI18n = badgeMap

	resp.ViewAll = decodeViewAll(ctx, m.ViewAll)

	return resp, nil
}

// LabelI18nFromRequest merges request translations with any existing DB values.
func LabelI18nFromRequest(req *UpsertNavMenuRequest, existing datatypes.JSON) datatypes.JSON {
	current := decodeLabelI18n(existing, req.Label)
	merged := i18n.MergeLocalized(current, req.LabelI18n, req.Label)
	return datatypes.JSON(i18n.MarshalLocalizedField(merged))
}

// BadgeI18nFromRequest merges badge translations with any existing DB values.
func BadgeI18nFromRequest(req *UpsertNavMenuRequest, existing datatypes.JSON) datatypes.JSON {
	current := decodeBadgeI18n(existing, req.Badge)
	merged := i18n.MergeLocalized(current, req.BadgeI18n, derefString(req.Badge))
	return datatypes.JSON(i18n.MarshalLocalizedField(merged))
}

// NavMenuJSONFromRequest encodes nested mega-menu JSON for persistence.
func NavMenuJSONFromRequest(req *UpsertNavMenuRequest) (viewAll, columns, featured datatypes.JSON) {
	if req.ViewAll != nil {
		viewAll = encodeViewAll(req.ViewAll)
	}
	columns = encodeColumns(req.Columns)
	featured = encodeFeatured(req.Featured)
	return viewAll, columns, featured
}

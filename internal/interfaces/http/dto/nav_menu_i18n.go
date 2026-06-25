package dto

import (
	"context"
	"encoding/json"

	"github.com/alireza-akbarzadeh/luxe/internal/i18n"
	"gorm.io/datatypes"
)

type storedViewAll struct {
	Label json.RawMessage `json:"label"`
	Href  string          `json:"href"`
}

type storedLink struct {
	Title json.RawMessage `json:"title"`
	Href  string          `json:"href"`
}

type storedColumn struct {
	Title json.RawMessage `json:"title"`
	Links []storedLink    `json:"links"`
}

type storedFeatured struct {
	Title       json.RawMessage `json:"title"`
	Description json.RawMessage `json:"description,omitempty"`
	Href        string          `json:"href"`
	Image       string          `json:"image"`
	Badge       json.RawMessage `json:"badge,omitempty"`
}

func encodeLocalizedField(fallback string, translations i18n.LocalizedMap) json.RawMessage {
	return i18n.MarshalLocalizedField(i18n.MergeLocalized(nil, translations, fallback))
}

func decodeLocalizedField(raw json.RawMessage) i18n.LocalizedMap {
	m, _ := i18n.ParseLocalizedField(raw)
	return m
}

func encodeViewAll(input *ViewAll) datatypes.JSON {
	if input == nil {
		return nil
	}
	stored := storedViewAll{
		Label: encodeLocalizedField(input.Label, input.LabelI18n),
		Href:  input.Href,
	}
	raw, _ := json.Marshal(stored)
	return raw
}

func decodeViewAll(ctx context.Context, raw datatypes.JSON) *ViewAll {
	if len(raw) == 0 {
		return nil
	}

	var stored storedViewAll
	if err := json.Unmarshal(raw, &stored); err != nil {
		// Legacy shape: {"label":"View all","href":"/shop"}
		var legacy ViewAll
		if err := json.Unmarshal(raw, &legacy); err != nil {
			return nil
		}
		labelMap := i18n.LocalizedMap{i18n.DefaultLocale: legacy.Label}
		return &ViewAll{
			Label:     labelMap.Resolve(ctx, legacy.Label),
			LabelI18n: labelMap,
			Href:      legacy.Href,
		}
	}

	labelMap, fallback := i18n.ParseLocalizedField(stored.Label)
	return &ViewAll{
		Label:     labelMap.Resolve(ctx, fallback),
		LabelI18n: labelMap,
		Href:      stored.Href,
	}
}

func encodeColumns(columns []Column) datatypes.JSON {
	if len(columns) == 0 {
		return nil
	}
	stored := make([]storedColumn, 0, len(columns))
	for _, column := range columns {
		item := storedColumn{
			Title: encodeLocalizedField(column.Title, column.TitleI18n),
			Links: make([]storedLink, 0, len(column.Links)),
		}
		for _, link := range column.Links {
			item.Links = append(item.Links, storedLink{
				Title: encodeLocalizedField(link.Title, link.TitleI18n),
				Href:  link.Href,
			})
		}
		stored = append(stored, item)
	}
	raw, _ := json.Marshal(stored)
	return raw
}

func decodeColumns(ctx context.Context, raw datatypes.JSON) []Column {
	if len(raw) == 0 {
		return nil
	}

	var stored []storedColumn
	if err := json.Unmarshal(raw, &stored); err != nil {
		// Legacy plain-string columns.
		var legacy []Column
		if err := json.Unmarshal(raw, &legacy); err != nil {
			return nil
		}
		out := make([]Column, 0, len(legacy))
		for _, column := range legacy {
			titleMap := i18n.LocalizedMap{i18n.DefaultLocale: column.Title}
			outColumn := Column{
				Title:     titleMap.Resolve(ctx, column.Title),
				TitleI18n: titleMap,
				Links:     make([]Link, 0, len(column.Links)),
			}
			for _, link := range column.Links {
				linkMap := i18n.LocalizedMap{i18n.DefaultLocale: link.Title}
				outColumn.Links = append(outColumn.Links, Link{
					Title:     linkMap.Resolve(ctx, link.Title),
					TitleI18n: linkMap,
					Href:      link.Href,
				})
			}
			out = append(out, outColumn)
		}
		return out
	}

	out := make([]Column, 0, len(stored))
	for _, column := range stored {
		titleMap, titleFallback := i18n.ParseLocalizedField(column.Title)
		outColumn := Column{
			Title:     titleMap.Resolve(ctx, titleFallback),
			TitleI18n: titleMap,
			Links:     make([]Link, 0, len(column.Links)),
		}
		for _, link := range column.Links {
			linkMap, linkFallback := i18n.ParseLocalizedField(link.Title)
			outColumn.Links = append(outColumn.Links, Link{
				Title:     linkMap.Resolve(ctx, linkFallback),
				TitleI18n: linkMap,
				Href:      link.Href,
			})
		}
		out = append(out, outColumn)
	}
	return out
}

func encodeFeatured(items []FeaturedItem) datatypes.JSON {
	if len(items) == 0 {
		return nil
	}
	stored := make([]storedFeatured, 0, len(items))
	for _, item := range items {
		stored = append(stored, storedFeatured{
			Title:       encodeLocalizedField(item.Title, item.TitleI18n),
			Description: encodeLocalizedField(item.Description, item.DescriptionI18n),
			Href:        item.Href,
			Image:       item.Image,
			Badge:       encodeLocalizedField(derefString(item.Badge), item.BadgeI18n),
		})
	}
	raw, _ := json.Marshal(stored)
	return raw
}

func decodeFeatured(ctx context.Context, raw datatypes.JSON) []FeaturedItem {
	if len(raw) == 0 {
		return nil
	}

	var stored []storedFeatured
	if err := json.Unmarshal(raw, &stored); err != nil {
		var legacy []FeaturedItem
		if err := json.Unmarshal(raw, &legacy); err != nil {
			return nil
		}
		out := make([]FeaturedItem, 0, len(legacy))
		for _, item := range legacy {
			titleMap := i18n.LocalizedMap{i18n.DefaultLocale: item.Title}
			descMap := i18n.LocalizedMap{i18n.DefaultLocale: item.Description}
			var badgeMap i18n.LocalizedMap
			if item.Badge != nil {
				badgeMap = i18n.LocalizedMap{i18n.DefaultLocale: *item.Badge}
			}
			out = append(out, FeaturedItem{
				Title:           titleMap.Resolve(ctx, item.Title),
				TitleI18n:       titleMap,
				Description:     descMap.Resolve(ctx, item.Description),
				DescriptionI18n: descMap,
				Href:            item.Href,
				Image:           item.Image,
				Badge:           localizedOptionalString(ctx, badgeMap, item.Badge),
				BadgeI18n:       badgeMap,
			})
		}
		return out
	}

	out := make([]FeaturedItem, 0, len(stored))
	for _, item := range stored {
		titleMap, titleFallback := i18n.ParseLocalizedField(item.Title)
		descMap, descFallback := i18n.ParseLocalizedField(item.Description)
		badgeMap, badgeFallback := i18n.ParseLocalizedField(item.Badge)
		out = append(out, FeaturedItem{
			Title:           titleMap.Resolve(ctx, titleFallback),
			TitleI18n:       titleMap,
			Description:     descMap.Resolve(ctx, descFallback),
			DescriptionI18n: descMap,
			Href:            item.Href,
			Image:           item.Image,
			Badge:           localizedOptionalString(ctx, badgeMap, optionalString(badgeFallback)),
			BadgeI18n:       badgeMap,
		})
	}
	return out
}

func localizedOptionalString(ctx context.Context, m i18n.LocalizedMap, fallback *string) *string {
	fb := derefString(fallback)
	if len(m) == 0 && fb == "" {
		return nil
	}
	resolved := m.Resolve(ctx, fb)
	return &resolved
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func decodeLabelI18n(raw datatypes.JSON, fallback string) i18n.LocalizedMap {
	if len(raw) == 0 {
		if fallback == "" {
			return nil
		}
		return i18n.LocalizedMap{i18n.DefaultLocale: fallback}
	}
	m, _ := i18n.ParseLocalizedField(json.RawMessage(raw))
	if len(m) == 0 && fallback != "" {
		return i18n.LocalizedMap{i18n.DefaultLocale: fallback}
	}
	return m
}

func decodeBadgeI18n(raw datatypes.JSON, fallback *string) i18n.LocalizedMap {
	fb := derefString(fallback)
	if len(raw) == 0 {
		if fb == "" {
			return nil
		}
		return i18n.LocalizedMap{i18n.DefaultLocale: fb}
	}
	m, _ := i18n.ParseLocalizedField(json.RawMessage(raw))
	if len(m) == 0 && fb != "" {
		return i18n.LocalizedMap{i18n.DefaultLocale: fb}
	}
	return m
}

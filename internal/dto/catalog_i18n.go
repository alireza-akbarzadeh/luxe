package dto

import (
	"context"
	"encoding/json"

	"github.com/alireza-akbarzadeh/luxe/internal/i18n"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/datatypes"
)

// DecodeCatalogNameI18n parses stored JSONB name translations.
func DecodeCatalogNameI18n(raw datatypes.JSON, fallback string) i18n.LocalizedMap {
	return decodeCatalogNameI18n(raw, fallback)
}

func decodeCatalogNameI18n(raw datatypes.JSON, fallback string) i18n.LocalizedMap {
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

func decodeCatalogDescriptionI18n(raw datatypes.JSON, fallback string) i18n.LocalizedMap {
	return decodeCatalogNameI18n(raw, fallback)
}

// EncodeCatalogI18n merges incoming locale map into stored JSONB.
func EncodeCatalogI18n(existing datatypes.JSON, incoming i18n.LocalizedMap, fallback string) datatypes.JSON {
	return encodeCatalogI18n(existing, incoming, fallback)
}

func encodeCatalogI18n(existing datatypes.JSON, incoming i18n.LocalizedMap, fallback string) datatypes.JSON {
	current := decodeCatalogNameI18n(existing, fallback)
	merged := i18n.MergeLocalized(current, incoming, fallback)
	return datatypes.JSON(i18n.MarshalLocalizedField(merged))
}

// EncodeSearchAliases stores search synonym list as JSONB.
func EncodeSearchAliases(aliases []string) datatypes.JSON {
	return encodeSearchAliases(aliases)
}

func encodeSearchAliases(aliases []string) datatypes.JSON {
	if len(aliases) == 0 {
		return datatypes.JSON([]byte("[]"))
	}
	raw, err := json.Marshal(aliases)
	if err != nil {
		return datatypes.JSON([]byte("[]"))
	}
	return raw
}

// MergeSearchAliases appends unique aliases to existing JSONB.
func MergeSearchAliases(existing datatypes.JSON, incoming []string) datatypes.JSON {
	return mergeSearchAliases(existing, incoming)
}

func mergeSearchAliases(existing datatypes.JSON, incoming []string) datatypes.JSON {
	if len(incoming) == 0 {
		return existing
	}
	current := i18n.ParseSearchAliases(json.RawMessage(existing))
	seen := make(map[string]struct{}, len(current)+len(incoming))
	merged := make([]string, 0, len(current)+len(incoming))
	for _, alias := range append(current, incoming...) {
		if _, ok := seen[alias]; ok {
			continue
		}
		seen[alias] = struct{}{}
		merged = append(merged, alias)
	}
	return encodeSearchAliases(merged)
}

// BuildProductSearchDocument assembles cross-locale search text for a product row.
func BuildProductSearchDocument(p *models.Product, category *models.Category) string {
	nameMap := decodeCatalogNameI18n(p.NameI18n, p.Name)
	descMap := decodeCatalogDescriptionI18n(p.DescriptionI18n, p.Description)
	parts := append([]string{}, i18n.CollectLocalizedTexts(nameMap)...)
	parts = append(parts, i18n.CollectLocalizedTexts(descMap)...)
	parts = append(parts, p.Name, p.Description, p.SKU, p.Barcode, p.Slug)
	if category != nil {
		catNameMap := decodeCatalogNameI18n(category.NameI18n, category.Name)
		parts = append(parts, i18n.CollectLocalizedTexts(catNameMap)...)
		parts = append(parts, category.Name)
	}
	parts = append(parts, i18n.ParseSearchAliases(json.RawMessage(p.SearchAliases))...)
	return i18n.BuildSearchDocument(parts...)
}

// BuildCategorySearchDocument assembles cross-locale search text for a category row.
func BuildCategorySearchDocument(c *models.Category) string {
	nameMap := decodeCatalogNameI18n(c.NameI18n, c.Name)
	descMap := decodeCatalogDescriptionI18n(c.DescriptionI18n, c.Description)
	parts := append([]string{}, i18n.CollectLocalizedTexts(nameMap)...)
	parts = append(parts, i18n.CollectLocalizedTexts(descMap)...)
	parts = append(parts, c.Name, c.Description, c.Slug)
	return i18n.BuildSearchDocument(parts...)
}

// LocalizeCategoryModel resolves name/description on a category for API output.
func LocalizeCategoryModel(ctx context.Context, c *models.Category) {
	if c == nil {
		return
	}
	nameMap := decodeCatalogNameI18n(c.NameI18n, c.Name)
	descMap := decodeCatalogDescriptionI18n(c.DescriptionI18n, c.Description)
	c.Name = nameMap.Resolve(ctx, c.Name)
	c.Description = descMap.Resolve(ctx, c.Description)
}

// LocalizeCategoryModels resolves a slice of categories in place.
func LocalizeCategoryModels(ctx context.Context, categories []models.Category) {
	for i := range categories {
		LocalizeCategoryModel(ctx, &categories[i])
	}
}

// resolvedCategoryResponse maps a category with locale-aware fields.
func resolvedCategoryResponse(ctx context.Context, c *models.Category) CategoryResponse {
	if c == nil || c.ID == 0 {
		return CategoryResponse{}
	}
	nameMap := decodeCatalogNameI18n(c.NameI18n, c.Name)
	descMap := decodeCatalogDescriptionI18n(c.DescriptionI18n, c.Description)
	return CategoryResponse{
		ID:             c.ID,
		Name:           nameMap.Resolve(ctx, c.Name),
		NameI18n:       nameMap,
		Slug:           c.Slug,
		Description:    descMap.Resolve(ctx, c.Description),
		DescriptionI18n: descMap,
		Level:          c.Level,
		Path:           c.Path,
		IsActive:       c.IsActive,
		ParentID:       c.ParentID,
	}
}

// applyProductI18nFields sets resolved name/description on ProductResponse.
func applyProductI18nFields(ctx context.Context, p *models.Product, r *ProductResponse) {
	nameMap := decodeCatalogNameI18n(p.NameI18n, p.Name)
	descMap := decodeCatalogDescriptionI18n(p.DescriptionI18n, p.Description)
	r.Name = nameMap.Resolve(ctx, p.Name)
	r.NameI18n = nameMap
	r.Description = descMap.Resolve(ctx, p.Description)
	r.DescriptionI18n = descMap
}

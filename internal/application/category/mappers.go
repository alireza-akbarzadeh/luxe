package category

import (
	"errors"
	"regexp"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

var ErrHasChildren = errors.New("category has children")

var slugSanitizer = regexp.MustCompile(`[^a-z0-9]+`)

func generateSlug(name string) string {
	slug := strings.ToLower(strings.TrimSpace(name))
	slug = slugSanitizer.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "item"
	}
	return slug
}

// BuildCreateModel maps a create DTO to a category model.
func BuildCreateModel(req dto.CreateCategoryRequest, slug string) *models.Category {
	category := &models.Category{
		Name:        req.Name,
		Slug:        slug,
		Description: req.Description,
		ParentID:    req.ParentID,
		IsActive:    req.IsActive,
	}
	category.NameI18n = dto.EncodeCatalogI18n(category.NameI18n, req.NameI18n, category.Name)
	category.DescriptionI18n = dto.EncodeCatalogI18n(category.DescriptionI18n, req.DescriptionI18n, category.Description)
	return category
}

// ApplyUpdateDTO mutates a loaded category from an update DTO.
func ApplyUpdateDTO(category *models.Category, req dto.UpdateCategoryRequest, newSlug string) {
	if req.Name != nil {
		category.Name = *req.Name
		if newSlug != "" {
			category.Slug = newSlug
		}
	}
	if req.Slug != nil && *req.Slug != "" {
		category.Slug = newSlug
	}
	if req.Description != nil {
		category.Description = *req.Description
	}
	if req.ParentID != nil {
		category.ParentID = req.ParentID
	}
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}
	if req.Name != nil || len(req.NameI18n) > 0 {
		category.NameI18n = dto.EncodeCatalogI18n(category.NameI18n, req.NameI18n, category.Name)
	}
	if req.Description != nil || len(req.DescriptionI18n) > 0 {
		category.DescriptionI18n = dto.EncodeCatalogI18n(category.DescriptionI18n, req.DescriptionI18n, category.Description)
	}
	category.SearchDocument = dto.BuildCategorySearchDocument(category)
}

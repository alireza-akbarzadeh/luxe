package collection

import (
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

func parseScheduleTime(raw *string) (*time.Time, error) {
	if raw == nil {
		return nil, nil
	}
	trimmed := *raw
	if trimmed == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, trimmed)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

// BuildCreateModel maps a create DTO to a collection model.
func BuildCreateModel(req *dto.CreateCollectionRequest, slug string) (*models.Collection, error) {
	status := req.Status
	if status == "" {
		status = "draft"
	}
	collectionType := req.CollectionType
	if collectionType == "" {
		collectionType = "smart"
	}
	href := req.Href
	if href == "" {
		href = "/shop"
	}
	cta := req.CTALabel
	if cta == "" {
		cta = "Shop collection"
	}
	startsAt, err := parseScheduleTime(req.StartsAt)
	if err != nil {
		return nil, err
	}
	endsAt, err := parseScheduleTime(req.EndsAt)
	if err != nil {
		return nil, err
	}
	return &models.Collection{
		Slug:              slug,
		Eyebrow:           req.Eyebrow,
		Title:             req.Title,
		Description:       req.Description,
		Href:              href,
		ImageURL:          req.ImageURL,
		CTALabel:          cta,
		SortOrder:         req.SortOrder,
		Status:            status,
		CollectionType:    collectionType,
		StartsAt:          startsAt,
		EndsAt:            endsAt,
		PreviewSort:       req.PreviewSort,
		PreviewIsNew:      req.PreviewIsNew,
		PreviewCategoryID: req.PreviewCategoryID,
		Theme:             req.Theme,
	}, nil
}

// ApplyUpdateDTO mutates a loaded collection from an update DTO.
func ApplyUpdateDTO(collection *models.Collection, req *dto.UpdateCollectionRequest, newSlug string) error {
	if newSlug != "" {
		collection.Slug = newSlug
	}
	if req.Eyebrow != nil {
		collection.Eyebrow = *req.Eyebrow
	}
	if req.Title != nil {
		collection.Title = *req.Title
	}
	if req.Description != nil {
		collection.Description = *req.Description
	}
	if req.Href != nil {
		collection.Href = *req.Href
	}
	if req.ImageURL != nil {
		collection.ImageURL = *req.ImageURL
	}
	if req.CTALabel != nil {
		collection.CTALabel = *req.CTALabel
	}
	if req.SortOrder != nil {
		collection.SortOrder = *req.SortOrder
	}
	if req.Status != nil {
		collection.Status = *req.Status
	}
	if req.CollectionType != nil {
		collection.CollectionType = *req.CollectionType
	}
	if req.StartsAt != nil {
		startsAt, err := parseScheduleTime(req.StartsAt)
		if err != nil {
			return err
		}
		collection.StartsAt = startsAt
	}
	if req.EndsAt != nil {
		endsAt, err := parseScheduleTime(req.EndsAt)
		if err != nil {
			return err
		}
		collection.EndsAt = endsAt
	}
	if req.PreviewSort != nil {
		collection.PreviewSort = *req.PreviewSort
	}
	if req.PreviewIsNew != nil {
		collection.PreviewIsNew = req.PreviewIsNew
	}
	if req.PreviewCategoryID != nil {
		collection.PreviewCategoryID = req.PreviewCategoryID
	}
	if req.Theme != nil {
		collection.Theme = *req.Theme
	}
	return nil
}

// ProductIDsFromModel returns ordered product ids from preloaded collection products.
func ProductIDsFromModel(c *models.Collection) []uint {
	if len(c.Products) == 0 {
		return nil
	}
	ids := make([]uint, 0, len(c.Products))
	for _, item := range c.Products {
		ids = append(ids, item.ProductID)
	}
	return ids
}

// ToResponse maps a collection model to an API response.
func ToResponse(c *models.Collection) *dto.CollectionResponse {
	collectionType := c.CollectionType
	if collectionType == "" {
		collectionType = "smart"
	}
	resp := &dto.CollectionResponse{
		ID:                c.ID,
		Slug:              c.Slug,
		Eyebrow:           c.Eyebrow,
		Title:             c.Title,
		Description:       c.Description,
		Href:              c.Href,
		ImageURL:          c.ImageURL,
		CTALabel:          c.CTALabel,
		SortOrder:         c.SortOrder,
		Status:            c.Status,
		CollectionType:    collectionType,
		StartsAt:          c.StartsAt,
		EndsAt:            c.EndsAt,
		ProductIDs:        ProductIDsFromModel(c),
		PreviewSort:       c.PreviewSort,
		PreviewIsNew:      c.PreviewIsNew,
		PreviewCategoryID: c.PreviewCategoryID,
		Theme:             c.Theme,
		CreatedAt:         c.CreatedAt,
		UpdatedAt:         c.UpdatedAt,
	}
	if c.WorkflowState != nil {
		resp.WorkflowState = dto.ToStateView(c.WorkflowState)
	}
	return resp
}

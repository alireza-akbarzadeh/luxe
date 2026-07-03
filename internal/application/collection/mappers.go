package collection

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

// BuildCreateModel maps a create DTO to a collection model.
func BuildCreateModel(req *dto.CreateCollectionRequest, slug string) *models.Collection {
	status := req.Status
	if status == "" {
		status = "draft"
	}
	href := req.Href
	if href == "" {
		href = "/shop"
	}
	cta := req.CTALabel
	if cta == "" {
		cta = "Shop collection"
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
		PreviewSort:       req.PreviewSort,
		PreviewIsNew:      req.PreviewIsNew,
		PreviewCategoryID: req.PreviewCategoryID,
		Theme:             req.Theme,
	}
}

// ApplyUpdateDTO mutates a loaded collection from an update DTO.
func ApplyUpdateDTO(collection *models.Collection, req *dto.UpdateCollectionRequest, newSlug string) {
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
}

// ToResponse maps a collection model to an API response.
func ToResponse(c *models.Collection) *dto.CollectionResponse {
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

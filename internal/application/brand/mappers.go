package brand

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

// BuildCreateModel maps a create DTO to a brand model.
func BuildCreateModel(req *dto.CreateBrandRequest) *models.Brand {
	status := "draft"
	if req.Status != nil {
		status = *req.Status
	}
	return &models.Brand{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		LogoURL:     req.LogoURL,
		Status:      status,
	}
}

// ApplyUpdateDTO mutates a loaded brand from an update DTO.
func ApplyUpdateDTO(brand *models.Brand, req *dto.UpdateBrandRequest) {
	if req.Name != nil {
		brand.Name = *req.Name
	}
	if req.Slug != nil {
		brand.Slug = *req.Slug
	}
	if req.Description != nil {
		brand.Description = req.Description
	}
	if req.LogoURL != nil {
		brand.LogoURL = req.LogoURL
	}
	if req.Status != nil {
		brand.Status = *req.Status
	}
}

// ToResponse maps a brand model to an API response.
func ToResponse(b *models.Brand) *dto.BrandResponse {
	resp := &dto.BrandResponse{
		ID:          b.ID,
		Name:        b.Name,
		Slug:        b.Slug,
		Description: b.Description,
		LogoURL:     b.LogoURL,
		Status:      b.Status,
		CreatedAt:   b.CreatedAt,
		UpdatedAt:   b.UpdatedAt,
	}
	if b.WorkflowState != nil {
		resp.WorkflowState = dto.ToStateView(b.WorkflowState)
	}
	return resp
}

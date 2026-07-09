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
	featuredSort := 0
	if req.FeaturedSortOrder != nil {
		featuredSort = *req.FeaturedSortOrder
	}
	isFeatured := false
	if req.IsFeatured != nil {
		isFeatured = *req.IsFeatured
	}
	metaTitle := ""
	if req.MetaTitle != nil {
		metaTitle = *req.MetaTitle
	}
	metaDescription := ""
	if req.MetaDescription != nil {
		metaDescription = *req.MetaDescription
	}
	return &models.Brand{
		Name:              req.Name,
		Slug:              req.Slug,
		Description:       req.Description,
		LogoURL:           req.LogoURL,
		Status:            status,
		IsFeatured:        isFeatured,
		FeaturedSortOrder: featuredSort,
		MetaTitle:         metaTitle,
		MetaDescription:   metaDescription,
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
	if req.IsFeatured != nil {
		brand.IsFeatured = *req.IsFeatured
	}
	if req.FeaturedSortOrder != nil {
		brand.FeaturedSortOrder = *req.FeaturedSortOrder
	}
	if req.MetaTitle != nil {
		brand.MetaTitle = *req.MetaTitle
	}
	if req.MetaDescription != nil {
		brand.MetaDescription = *req.MetaDescription
	}
}

// ToResponse maps a brand model to an API response.
func ToResponse(b *models.Brand, productCount ...int64) *dto.BrandResponse {
	resp := &dto.BrandResponse{
		ID:                b.ID,
		Name:              b.Name,
		Slug:              b.Slug,
		Description:       b.Description,
		LogoURL:           b.LogoURL,
		Status:            b.Status,
		IsFeatured:        b.IsFeatured,
		FeaturedSortOrder: b.FeaturedSortOrder,
		MetaTitle:         b.MetaTitle,
		MetaDescription:   b.MetaDescription,
		CreatedAt:         b.CreatedAt,
		UpdatedAt:         b.UpdatedAt,
	}
	if len(productCount) > 0 {
		count := productCount[0]
		resp.ProductCount = &count
	}
	if b.WorkflowState != nil {
		resp.WorkflowState = dto.ToStateView(b.WorkflowState)
	}
	return resp
}

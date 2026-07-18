package catalog

import (
	domain "github.com/alireza-akbarzadeh/luxe/internal/domain/catalog"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

// BuildAttributes converts DTO attribute inputs into model attributes.
func BuildAttributes(inputs []dto.ProductAttributeInput) []models.ProductAttribute {
	attrs := make([]models.ProductAttribute, 0, len(inputs))
	for _, a := range inputs {
		attrs = append(attrs, models.ProductAttribute{
			Name:   a.Name,
			Values: a.Values,
		})
	}
	return attrs
}

// BuildCreateModel maps a create DTO to a GORM product model.
func BuildCreateModel(req dto.CreateProductRequest, slug string) *models.Product {
	product := &models.Product{
		Name:              req.Name,
		Slug:              slug,
		Description:       req.Description,
		Price:             req.Price,
		CompareAtPrice:    req.CompareAtPrice,
		Cost:              req.Cost,
		SKU:               req.SKU,
		Barcode:           req.Barcode,
		Stock:             req.Stock,
		LowStockThreshold: req.LowStockThreshold,
		Weight:            req.Weight,
		IsDigital:         req.IsDigital,
		CategoryID:        req.CategoryID,
		BrandID:           req.BrandID,
		Images:            req.Images,
		Status:            req.Status,
		MetaTitle:         req.MetaTitle,
		MetaDescription:   req.MetaDescription,
		IsNew:             false,
		Rating:            0,
		ReviewsCount:      0,
		Colors:            req.Colors,
		Sizes:             req.Sizes,
		Attributes:        BuildAttributes(req.Attributes),
	}
	if req.IsNew != nil {
		product.IsNew = *req.IsNew
	}
	if req.StoreID != nil {
		product.StoreID = *req.StoreID
	}
	if req.TrackInventory != nil {
		product.TrackInventory = *req.TrackInventory
	}
	if req.WarehouseLocation != "" {
		product.WarehouseLocation = req.WarehouseLocation
	}
	if req.AllowBackorder != nil {
		product.AllowBackorder = *req.AllowBackorder
	}
	if req.Visibility != "" {
		product.Visibility = req.Visibility
	}
	if len(req.Tags) > 0 {
		product.Tags = req.Tags
	}
	if len(req.Channels) > 0 {
		product.Channels = req.Channels
	}
	if req.PublishedAt != nil {
		product.PublishedAt = req.PublishedAt
	}
	if product.Status == "" {
		product.Status = "draft"
	}
	if product.LowStockThreshold == 0 {
		product.LowStockThreshold = 5
	}
	product.NameI18n = dto.EncodeCatalogI18n(product.NameI18n, req.NameI18n, product.Name)
	product.DescriptionI18n = dto.EncodeCatalogI18n(product.DescriptionI18n, req.DescriptionI18n, product.Description)
	product.SearchAliases = dto.EncodeSearchAliases(req.SearchAliases)
	return product
}

// BuildBulkCreateModel maps a create DTO to a model for bulk import (minimal fields).
func BuildBulkCreateModel(req dto.CreateProductRequest, slug string) *models.Product {
	p := BuildCreateModel(req, slug)
	p.Attributes = BuildAttributes(req.Attributes)
	return p
}

// CreateProductInputFromDTO maps HTTP create input to domain validation input.
func CreateProductInputFromDTO(req dto.CreateProductRequest) domain.CreateProductInput {
	return domain.CreateProductInput{
		Name:        req.Name,
		Description: req.Description,
		PriceCents:  int64(req.Price * 100),
		Currency:    "USD",
		SKU:         req.SKU,
		Stock:       req.Stock,
		CategoryID:  req.CategoryID,
		BrandID:     req.BrandID,
	}
}

// ApplyUpdateDTO mutates a loaded product from an update DTO.
func ApplyUpdateDTO(product *models.Product, req dto.UpdateProductRequest, newSlug string) {
	if req.Name != nil {
		product.Name = *req.Name
		if newSlug != "" {
			product.Slug = newSlug
		}
	}
	if req.StoreID != nil {
		product.StoreID = *req.StoreID
	}
	if req.Description != nil {
		product.Description = *req.Description
	}
	if req.Price != nil {
		product.Price = *req.Price
	}
	if req.CompareAtPrice != nil {
		product.CompareAtPrice = req.CompareAtPrice
	}
	if req.Cost != nil {
		product.Cost = req.Cost
	}
	if req.SKU != nil {
		product.SKU = *req.SKU
	}
	if req.Barcode != nil {
		product.Barcode = *req.Barcode
	}
	if req.LowStockThreshold != nil {
		product.LowStockThreshold = *req.LowStockThreshold
	}
	if req.Weight != nil {
		product.Weight = req.Weight
	}
	if req.IsDigital != nil {
		product.IsDigital = *req.IsDigital
	}
	if req.CategoryID != nil {
		product.CategoryID = req.CategoryID
	}
	if req.BrandID != nil {
		product.BrandID = req.BrandID
	}
	if req.Images != nil {
		product.Images = *req.Images
	}
	if req.Status != nil {
		product.Status = *req.Status
	}
	if req.MetaTitle != nil {
		product.MetaTitle = *req.MetaTitle
	}
	if req.MetaDescription != nil {
		product.MetaDescription = *req.MetaDescription
	}
	if req.IsNew != nil {
		product.IsNew = *req.IsNew
	}
	if req.Colors != nil {
		product.Colors = *req.Colors
	}
	if req.Sizes != nil {
		product.Sizes = *req.Sizes
	}
	if req.TrackInventory != nil {
		product.TrackInventory = *req.TrackInventory
	}
	if req.WarehouseLocation != nil {
		product.WarehouseLocation = *req.WarehouseLocation
	}
	if req.AllowBackorder != nil {
		product.AllowBackorder = *req.AllowBackorder
	}
	if req.Visibility != nil {
		product.Visibility = *req.Visibility
	}
	if req.Tags != nil {
		product.Tags = *req.Tags
	}
	if req.Channels != nil {
		product.Channels = *req.Channels
	}
	if req.PublishedAt != nil {
		product.PublishedAt = req.PublishedAt
	}
	if req.Name != nil || len(req.NameI18n) > 0 {
		product.NameI18n = dto.EncodeCatalogI18n(product.NameI18n, req.NameI18n, product.Name)
	}
	if req.Description != nil || len(req.DescriptionI18n) > 0 {
		product.DescriptionI18n = dto.EncodeCatalogI18n(product.DescriptionI18n, req.DescriptionI18n, product.Description)
	}
	if req.SearchAliases != nil {
		product.SearchAliases = dto.MergeSearchAliases(product.SearchAliases, *req.SearchAliases)
	}
}

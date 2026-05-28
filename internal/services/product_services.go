package services

import (
	"errors"
	"fmt"

	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"gorm.io/gorm"
)

type ProductServiceInterface interface {
	List(limit, offset int, filters dto.ProductListFilters) ([]*models.Product, int64, error)
	BulkCreate(products []dto.CreateProductRequest) ([]*models.Product, error)
	GetByID(id uint) (*models.Product, error)
	GetBySlug(slug string) (*models.Product, error)
	Create(req dto.CreateProductRequest) (*models.Product, error)
	Update(productID uint, req dto.UpdateProductRequest) (*models.Product, error)
	Delete(id uint) error
	BulkDelete(productIDs []uint) error
	CheckLowStockAndAlert() error
	GetRelated(productID uint, limit int) ([]*models.Product, error)
	GetSuggestions(productIDs []uint, limit int) ([]*models.Product, error)
	GetByStoreID(storeID uint, limit, offset int, filters dto.ProductListFilters) ([]*models.Product, int64, error)
}

type productService struct {
	db *gorm.DB
}

func NewProductService(db *gorm.DB) ProductServiceInterface {
	return &productService{db: db}
}

// UniqSlug ensureUniqueSlug checks and modifies slug to be unique.
func (s *productService) UniqSlug(baseSlug string, excludeID uint) string {
	slug := baseSlug
	counter := 1
	for {
		var count int64
		query := s.db.Model(&models.Product{}).Where("slug = ?", slug)
		if excludeID > 0 {
			query = query.Where("id != ?", excludeID)
		}
		query.Count(&count)
		if count == 0 {
			break
		}
		slug = fmt.Sprintf("%s-%d", baseSlug, counter)
		counter++
	}
	return slug
}

func (s *productService) Create(req dto.CreateProductRequest) (*models.Product, error) {
	baseSlug := generateSlug(req.Name)
	slug := s.UniqSlug(baseSlug, 0)
	product := models.Product{
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
		Images:            req.Images,
		Status:            req.Status,
		MetaTitle:         req.MetaTitle,
		MetaDescription:   req.MetaDescription,
		IsNew:             false,
		Rating:            0.0,
		ReviewsCount:      0,
		Colors:            req.Colors,
		Sizes:             req.Sizes,
	}
	if req.IsNew != nil {
		product.IsNew = *req.IsNew
	}
	if req.StoreID != nil {
		product.StoreID = *req.StoreID
	}
	if product.Status == "" {
		product.Status = "draft"
	}
	if product.LowStockThreshold == 0 {
		product.LowStockThreshold = 5
	}
	if req.IsNew != nil {
		product.IsNew = *req.IsNew
	}

	if err := s.db.Create(product).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}
	return &product, nil
}

// GetByID Retrieve product by id
func (s *productService) GetByID(id uint) (*models.Product, error) {
	var product models.Product
	if err := s.db.Preload("Category").Preload("Store").First(&product, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("product not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return &product, nil
}

// GetBySlug Retrieve product by slug
func (s *productService) GetBySlug(slug string) (*models.Product, error) {
	var product models.Product
	if err := s.db.Preload("Category").Preload("Store").Where("slug = ?", slug).First(&product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("product not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return &product, nil
}

// Update product

func (s *productService) Update(id uint, req dto.UpdateProductRequest) (*models.Product, error) {
	product, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		product.Name = *req.Name
		baseSlug := generateSlug(*req.Name)
		product.Slug = s.UniqSlug(baseSlug, id)
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
		var existing models.Product
		if err := s.db.Where("sku = ? AND id != ?", *req.SKU, id).First(&existing).Error; err == nil {
			return nil, utils.ErrConflict("SKU already exists")
		}
		product.SKU = *req.SKU
	}
	if req.Barcode != nil {
		product.Barcode = *req.Barcode
	}
	if req.Stock != nil {
		product.Stock = *req.Stock
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
	if req.Images != nil {
		product.Images = *req.Images
	}
	if req.Status != nil {
		product.Status = *req.Status
	}
	if req.MetaTitle != nil {
		product.MetaTitle = *req.MetaTitle
	}
	if req.IsNew != nil {
		product.IsNew = *req.IsNew
	}

	if req.MetaDescription != nil {
		product.MetaDescription = *req.MetaDescription
	}
	if req.Colors != nil {
		product.Colors = *req.Colors
	}
	if req.Sizes != nil {
		product.Sizes = *req.Sizes
	}

	if err := s.db.Save(product).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}
	return product, nil
}

// Delete product
func (s *productService) Delete(id uint) error {
	result := s.db.Delete(&models.Product{}, id)
	if result.Error != nil {
		return utils.ErrInternal(result.Error)
	}
	if result.RowsAffected == 0 {
		return utils.ErrNotFound("product not found")
	}
	return nil
}

// List retrieve list of product
func (s *productService) List(limit, offset int, filters dto.ProductListFilters) ([]*models.Product, int64, error) {
	const maxLimit = 100
	if limit <= 0 {
		limit = 10
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	if offset < 0 {
		offset = 0
	}

	query := s.db.Model(&models.Product{}).Order("id DESC")

	// String filters
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.Name != "" {
		// Case-insensitive partial match for product name
		query = query.Where("LOWER(name) LIKE LOWER(?)", "%"+filters.Name+"%")
	}
	if filters.StoreID != nil && *filters.StoreID != 0 {
		query = query.Where("store_id = ?", *filters.StoreID)
	}
	if filters.SKU != "" {
		// Partial match for SKU (usually exact but can be partial)
		query = query.Where("sku LIKE ?", "%"+filters.SKU+"%")
	}

	// Numeric filters
	if filters.CategoryID != 0 {
		query = query.Where("category_id = ?", filters.CategoryID)
	}
	if filters.MinPrice != 0 {
		query = query.Where("price >= ?", filters.MinPrice)
	}
	if filters.MaxPrice != 0 {
		query = query.Where("price <= ?", filters.MaxPrice)
	}
	if filters.MinRating != 0 {
		query = query.Where("rating >= ?", filters.MinRating)
	}
	if filters.MaxRating != 0 {
		query = query.Where("rating <= ?", filters.MaxRating)
	}
	if filters.MinReviews != 0 {
		query = query.Where("reviews_count >= ?", filters.MinReviews)
	}
	if filters.MaxReviews != 0 {
		query = query.Where("reviews_count <= ?", filters.MaxReviews)
	}

	// Boolean filters (handle nil pointers)
	if filters.IsDigital != nil {
		query = query.Where("is_digital = ?", *filters.IsDigital)
	}
	if filters.IsNew != nil {
		query = query.Where("is_new = ?", *filters.IsNew)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count products: %w", err)
	}

	var products []*models.Product
	if err := query.Limit(limit).Offset(offset).Preload("Category").Find(&products).Error; err != nil {
		return nil, 0, fmt.Errorf("find products: %w", err)
	}

	return products, total, nil
}

// BulkCreate create multiple product
// BulkCreate create multiple product
func (s *productService) BulkCreate(products []dto.CreateProductRequest) ([]*models.Product, error) {
	if len(products) == 0 {
		return nil, utils.ErrBadRequest("no products provided")
	}
	var createdProducts []*models.Product

	err := s.db.Transaction(func(tx *gorm.DB) error {
		for _, p := range products {
			baseSlug := generateSlug(p.Name)
			slug := s.UniqSlug(baseSlug, 0)

			// Build model from DTO
			req := &models.Product{
				Name:              p.Name,
				Slug:              slug,
				Description:       p.Description,
				Price:             p.Price,
				CompareAtPrice:    p.CompareAtPrice,
				Cost:              p.Cost,
				SKU:               p.SKU,
				Barcode:           p.Barcode,
				Stock:             p.Stock,
				LowStockThreshold: p.LowStockThreshold,
				Weight:            p.Weight,
				IsDigital:         p.IsDigital,
				CategoryID:        p.CategoryID,
				Images:            p.Images,
				Status:            p.Status,
				MetaTitle:         p.MetaTitle,
				MetaDescription:   p.MetaDescription,
				Colors:            p.Colors,
				Sizes:             p.Sizes,
				IsNew:             false, // default
				Rating:            0,
				ReviewsCount:      0,
			}

			// Apply optional fields
			if p.StoreID != nil {
				req.StoreID = *p.StoreID
			}
			if p.IsNew != nil {
				req.IsNew = *p.IsNew
			}

			// Set defaults on the MODEL, not on the DTO
			if req.Status == "" {
				req.Status = "draft"
			}
			if req.LowStockThreshold == 0 {
				req.LowStockThreshold = 5
			}

			// Create the product
			if err := tx.Create(req).Error; err != nil {
				return err
			}
			createdProducts = append(createdProducts, req)
		}
		return nil
	})
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	return createdProducts, nil
}

// BulkDelete remove multiple product with the give ids
func (s *productService) BulkDelete(productIDs []uint) error {
	if len(productIDs) == 0 {
		return utils.ErrBadRequest("no product IDs provided")
	}
	rest := s.db.Where("id IN ?", productIDs).Delete(&models.Product{})
	if rest.Error != nil {
		return utils.ErrInternal(rest.Error)
	}
	if rest.RowsAffected == 0 {
		return utils.ErrNotFound("products not found")
	}
	return nil
}

// CheckLowStockAndAlert scans for products with stock <= low_stock_threshold
// and logs a warning for each. Returns an error if the database query fails.
func (s *productService) CheckLowStockAndAlert() error {
	var products []models.Product
	err := s.db.Where("stock <= low_stock_threshold AND status = ?", "active").
		Find(&products).Error
	if err != nil {
		return utils.ErrInternal(err)
	}

	if len(products) == 0 {
		utils.Log.Info("Low stock check: no products below threshold")
		return nil
	}

	// Log each low‑stock product (you can replace with email or notification)
	for _, p := range products {
		utils.Log.Warnf("LOW STOCK ALERT: Product ID=%d, Name=%s, Stock=%d, Threshold=%d",
			p.ID, p.Name, p.Stock, p.LowStockThreshold)
	}

	// Optional: also send a summary email to the admin
	// s.sendLowStockEmail(products)

	return nil
}

func (s *productService) GetRelated(productID uint, limit int) ([]*models.Product, error) {
	var product models.Product

	err := s.db.First(&product, productID).Error
	if err != nil {
		return nil, err
	}

	var related []*models.Product

	err = s.db.
		Preload("Category").
		Where("category_id = ? AND id != ?", product.CategoryID, productID).
		Order("rating DESC, reviews_count DESC").
		Limit(limit).
		Find(&related).Error
	if err != nil {
		return nil, err
	}

	return related, nil
}

func (s *productService) GetSuggestions(productIDs []uint, limit int) ([]*models.Product, error) {
	if limit == 0 {
		limit = 4
	}

	// Get categories of cart items
	var categoryIDs []uint
	s.db.Model(&models.Product{}).
		Where("id IN ?", productIDs).
		Distinct("category_id").
		Pluck("category_id", &categoryIDs)

	// Fetch products from those categories, excluding cart items
	var suggestions []*models.Product
	err := s.db.
		Preload("Category").
		Where("category_id IN ? AND id NOT IN ?", categoryIDs, productIDs).
		Order("rating DESC, reviews_count DESC, created_at DESC").
		Limit(limit).
		Find(&suggestions).Error

	return suggestions, err
}

func (s *productService) GetByStoreID(storeID uint, limit, offset int, filters dto.ProductListFilters) ([]*models.Product, int64, error) {
	filters.StoreID = &storeID
	return s.List(limit, offset, filters)
}

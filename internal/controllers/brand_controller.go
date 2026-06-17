// Package controllers provides HTTP handlers for the application.
package controllers

import (
	"errors"

	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/services"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// BrandController handles brand-related HTTP requests.
type BrandController struct {
	brandService services.BrandServiceInterface
	validate     *validator.Validate
}

// NewBrandController creates a new BrandController.
func NewBrandController(brandService services.BrandServiceInterface) *BrandController {
	return &BrandController{
		brandService: brandService,
		validate:     validator.New(),
	}
}

// CreateBrand godoc
// @Summary      Create a new brand
// @Description  Creates a brand with the provided details. Defaults status to "draft" if not supplied.
// @Tags         brands
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateBrandRequest true "Brand creation payload"
// @Success      201  {object}  utils.Response{data=dto.BrandResponse}  "Brand created"
// @Failure      400  {object}  utils.Response  "Validation error"
// @Failure      500  {object}  utils.Response  "Internal server error"
// @Router       /brands [post]
func (ctrl *BrandController) CreateBrand(c *gin.Context) {
	var req dto.CreateBrandRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	brand, err := ctrl.brandService.Create(c.Request.Context(), &req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to create brand")
		return
	}

	utils.CreatedResponse(c, "brand created successfully", brand)
}

// GetBrand godoc
// @Summary      Get a brand by ID
// @Description  Returns a single brand by its unique identifier.
// @Tags         brands
// @Produce      json
// @Param        id   path      int  true  "Brand ID"
// @Success      200  {object}  utils.Response{data=dto.BrandResponse}  "Brand found"
// @Failure      400  {object}  utils.Response  "Invalid ID"
// @Failure      404  {object}  utils.Response  "Brand not found"
// @Failure      500  {object}  utils.Response  "Internal server error"
// @Router       /brands/{id} [get]
func (ctrl *BrandController) GetBrand(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	brand, err := ctrl.brandService.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			utils.NotFoundResponse(c, "brand not found")
			return
		}
		utils.HandleServiceError(c, err, "failed to retrieve brand")
		return
	}

	utils.SuccessResponse(c, "brand retrieved", brand)
}

// ListBrands godoc
// @Summary      List all brands
// @Description  Returns a paginated list of brands with optional search and status filtering.
// @Tags         brands
// @Produce      json
// @Param        page    query     int     false  "Page number"           default(1)
// @Param        limit   query     int     false  "Items per page"        default(20)
// @Param        search  query     string  false  "Search by name or slug"
// @Param        status  query     string  false  "Filter by status"
// @Success      200     {object}  utils.Response{data=[]dto.BrandResponse}  "Brand list"
// @Failure      500     {object}  utils.Response  "Internal server error"
// @Router       /brands [get]
func (ctrl *BrandController) ListBrands(c *gin.Context) {
	var req dto.ListBrandsRequest
	if !utils.BindAndValidateQuery(c, &req, ctrl.validate) {
		return
	}

	brands, total, err := ctrl.brandService.List(c.Request.Context(), &req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list brands")
		return
	}

	// Optionally include total count in a more structured response
	// For now we return the slice directly as Data.
	utils.SuccessResponse(c, "brands retrieved", brands)
	// If you want total returned, you can wrap: gin.H{"items": brands, "total": total}
	_ = total // suppress unused variable; adjust as needed
}

// UpdateBrand godoc
// @Summary      Update a brand
// @Description  Partially updates a brand. Only supplied fields are changed.
// @Tags         brands
// @Accept       json
// @Produce      json
// @Param        id       path      int                     true  "Brand ID"
// @Param        request  body      dto.UpdateBrandRequest  true  "Brand update payload"
// @Success      200      {object}  utils.Response{data=dto.BrandResponse}  "Brand updated"
// @Failure      400      {object}  utils.Response  "Validation error"
// @Failure      404      {object}  utils.Response  "Brand not found"
// @Failure      500      {object}  utils.Response  "Internal server error"
// @Router       /brands/{id} [put]
func (ctrl *BrandController) UpdateBrand(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateBrandRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	brand, err := ctrl.brandService.Update(c.Request.Context(), id, &req)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			utils.NotFoundResponse(c, "brand not found")
			return
		}
		utils.HandleServiceError(c, err, "failed to update brand")
		return
	}

	utils.SuccessResponse(c, "brand updated", brand)
}

// DeleteBrand godoc
// @Summary      Delete a brand
// @Description  Permanently deletes a brand by ID.
// @Tags         brands
// @Produce      json
// @Param        id   path      int  true  "Brand ID"
// @Success      200  {object}  utils.Response  "Brand deleted"
// @Failure      400  {object}  utils.Response  "Invalid ID"
// @Failure      404  {object}  utils.Response  "Brand not found"
// @Failure      500  {object}  utils.Response  "Internal server error"
// @Router       /brands/{id} [delete]
func (ctrl *BrandController) DeleteBrand(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	err := ctrl.brandService.Delete(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			utils.NotFoundResponse(c, "brand not found")
			return
		}
		utils.HandleServiceError(c, err, "failed to delete brand")
		return
	}

	utils.SuccessResponse(c, "brand deleted", nil)
}

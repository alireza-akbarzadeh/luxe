package handlers

import (
	"net/http"
	"strconv"

	"github.com/alireza-akbarzadeh/luxe/internal/application/category"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type CategoryHandler struct {
	categoryService *category.Service
	validate        *validator.Validate
}

func NewCategoryHandler(categoryService *category.Service) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
		validate:        validator.New(),
	}
}

// Create creates a new category (admin only).
// @Summary      Create a new category
// @Description  Creates a new product category. Only accessible by users with the "admin" role.
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CreateCategoryRequest true "Category creation data"
// @Success      201 {object} dto.CategorySingleResponse
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/categories [post]
func (ctrl *CategoryHandler) Create(c *gin.Context) {
	var req dto.CreateCategoryRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	category, err := ctrl.categoryService.Create(req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to create category")
		return
	}
	dto.LocalizeCategoryModel(c.Request.Context(), category)
	resp := dto.CategorySingleResponse{
		BaseResponse: dto.BaseResponse{
			Success: true,
			Message: constants.MsgCreateSuccess,
			Code:    http.StatusCreated,
		},
		Data: dto.CategoryData{Category: *category},
	}
	c.JSON(http.StatusCreated, resp)
}

// Update updates an existing category (admin only).
// @Summary      Update a category
// @Description  Updates an existing category by ID. Only accessible by users with the "admin" role.
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path      int                           true  "Category ID"
// @Param        request body      dto.UpdateCategoryRequest true  "Category update data"
// @Success      200     {object}  dto.CategorySingleResponse
// @Failure      400     {object}  utils.Response
// @Failure      401     {object}  utils.Response
// @Failure      403     {object}  utils.Response
// @Failure      404     {object}  utils.Response
// @Failure      500     {object}  utils.Response
// @Router       /admin/categories/{id} [put]
func (ctrl *CategoryHandler) Update(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateCategoryRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	category, err := ctrl.categoryService.Update(id, req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to update category")
		return
	}
	dto.LocalizeCategoryModel(c.Request.Context(), category)
	resp := dto.CategorySingleResponse{
		BaseResponse: dto.BaseResponse{
			Success: true,
			Message: constants.MsgUpdateSuccess,
			Code:    http.StatusOK,
		},
		Data: dto.CategoryData{Category: *category},
	}
	c.JSON(http.StatusOK, resp)
}

// Delete deletes a category (admin only).
// @Summary      Delete a category
// @Description  Deletes a category by ID. Only accessible by users with the "admin" role.
// @Description  Categories with child categories cannot be deleted – delete children first.
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Category ID"
// @Success      200  {object}  dto.EmptyResponse
// @Failure      400  {object}  utils.Response
// @Failure      401  {object}  utils.Response
// @Failure      403  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /admin/categories/{id} [delete]
func (ctrl *CategoryHandler) Delete(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	if err := ctrl.categoryService.Delete(id); err != nil {
		utils.HandleServiceError(c, err, "failed to delete category")
		return
	}
	resp := dto.EmptyResponse{
		BaseResponse: dto.BaseResponse{
			Success: true,
			Message: constants.MsgDeleteSuccess,
			Code:    http.StatusOK,
		},
	}
	c.JSON(http.StatusOK, resp)
}

// GetOne retrieves a single category by ID or slug (public).
// @Summary      Get a category by ID or slug
// @Description  Returns a single category. Accepts either a numeric ID or a URL slug.
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Param        identifier   path      string  true  "Category ID (numeric) or slug (string)"
// @Success      200          {object}  dto.CategorySingleResponse
// @Failure      400          {object}  utils.Response
// @Failure      404          {object}  utils.Response
// @Failure      500          {object}  utils.Response
// @Router       /categories/{identifier} [get]
func (ctrl *CategoryHandler) GetOne(c *gin.Context) {
	identifier := c.Param("identifier")
	id, err := strconv.ParseUint(identifier, 10, 64)
	var category *models.Category
	if err == nil {
		category, err = ctrl.categoryService.GetByID(uint(id))
	} else {
		category, err = ctrl.categoryService.GetBySlug(identifier)
	}
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch category")
		return
	}
	dto.LocalizeCategoryModel(c.Request.Context(), category)
	resp := dto.CategorySingleResponse{
		BaseResponse: dto.BaseResponse{
			Success: true,
			Message: constants.MsgFetchSuccess,
			Code:    http.StatusOK,
		},
		Data: dto.CategoryData{Category: *category},
	}
	c.JSON(http.StatusOK, resp)
}

// List returns a paginated list of categories (public).
// @Summary      List categories
// @Description  Returns a paginated list of categories with optional filtering.
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Param        limit       query     int     false  "Items per page"  default(20)  minimum(1)  maximum(100)
// @Param        offset      query     int     false  "Offset (skip number of items)"  default(0)  minimum(0)
// @Param        is_active   query     bool    false  "Filter by active status (true/false)"
// @Param        parent_id   query     int     false  "Filter by parent category ID"
// @Param        search   	 query     string  false  "Filter by name  the name "
// @Param        sort        query     string   false  "Sort order (popular, name)"
// @Success      200         {object}  dto.CategoryListResponse
// @Failure      400         {object}  utils.Response
// @Failure      500         {object}  utils.Response
// @Router       /categories [get]
func (ctrl *CategoryHandler) List(c *gin.Context) {
	var req dto.CategoryListFilters
	if !utils.BindAndValidateQuery(c, &req, ctrl.validate) {
		return
	}

	categories, total, err := ctrl.categoryService.List(req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list categories")
		return
	}
	dto.LocalizeCategoryModels(c.Request.Context(), categories)

	resp := dto.CategoryListResponse{
		BaseResponse: dto.BaseResponse{
			Success: true,
			Message: constants.MsgFetchSuccess,
			Code:    http.StatusOK,
		},
		Data: dto.CategoryListData{
			Categories: categories,
			Total:      total,
			Limit:      req.Limit,
			Offset:     req.Offset,
		},
	}
	c.JSON(http.StatusOK, resp)
}

// BulkCreate creates multiple categories in one request (admin only).
// @Summary      Bulk create categories
// @Description  Creates multiple categories at once. Only accessible by users with the "admin" role.
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body []dto.CreateCategoryRequest true "Array of categories to create"
// @Success      201 {object} dto.BulkCreateCategoryResponse
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/categories/bulk [post]
func (ctrl *CategoryHandler) BulkCreate(c *gin.Context) {
	var reqs []dto.CreateCategoryRequest
	if !utils.BindAndValidate(c, &reqs, ctrl.validate) {
		return
	}
	if len(reqs) == 0 {
		utils.ErrorResponse(c, 400, "no categories provided")
		return
	}
	categories, err := ctrl.categoryService.BulkCreate(reqs)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to bulk create categories")
		return
	}
	for _, cat := range categories {
		dto.LocalizeCategoryModel(c.Request.Context(), cat)
	}
	resp := dto.BulkCreateCategoryResponse{
		BaseResponse: dto.BaseResponse{
			Success: true,
			Message: "categories created successfully",
			Code:    http.StatusCreated,
		},
		Data: dto.BulkCategoryData{Categories: categories},
	}
	c.JSON(http.StatusCreated, resp)
}

// BulkDelete removes multiple categories (admin only).
// @Summary      Bulk delete categories
// @Description  Deletes multiple categories by their IDs. Only accessible by users with the "admin" role.
// @Description  Cannot delete categories that have child categories – delete children first.
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body object true "Category IDs to delete" SchemaExample({"ids":[1,2,3]})
// @Success      200 {object} dto.EmptyResponse
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/categories/bulk [delete]
func (ctrl *CategoryHandler) BulkDelete(c *gin.Context) {
	var req category.BulkDeleteRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	if err := ctrl.categoryService.BulkDelete(req.IDs); err != nil {
		utils.HandleServiceError(c, err, "failed to bulk delete categories")
		return
	}
	resp := dto.EmptyResponse{
		BaseResponse: dto.BaseResponse{
			Success: true,
			Message: "categories deleted successfully",
			Code:    http.StatusOK,
		},
	}
	c.JSON(http.StatusOK, resp)
}

// GetCategoryByID retrieves a single category by ID (admin only).
// @Summary      Get a category by ID (admin)
// @Description  Returns a single category by its numeric ID. Only accessible by users with the "admin" role.
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Category ID"
// @Success      200  {object}  dto.CategorySingleResponse
// @Failure      400  {object}  utils.Response
// @Failure      401  {object}  utils.Response
// @Failure      403  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /admin/categories/{id} [get]
func (ctrl *CategoryHandler) GetCategoryByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, 400, "invalid category ID")
		return
	}

	category, err := ctrl.categoryService.GetByID(uint(id))
	if err != nil {
		utils.HandleServiceError(c, err, "failed to find category")
		return
	}
	dto.LocalizeCategoryModel(c.Request.Context(), category)

	// Return same DTO format as other admin endpoints
	resp := dto.CategorySingleResponse{
		BaseResponse: dto.BaseResponse{
			Success: true,
			Message: constants.MsgFetchSuccess,
			Code:    http.StatusOK,
		},
		Data: dto.CategoryData{Category: *category},
	}
	c.JSON(http.StatusOK, resp)
}

package controllers

import (
	"net/http"
	"strconv"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/middleware"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/services"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ProductController struct {
	productService  services.ProductServiceInterface
	userLikeService services.UsertLikeServiceInterface
	pdpService      services.PdpServiceInterface
	validate        *validator.Validate
}

func NewProductController(ps services.ProductServiceInterface, uls services.UsertLikeServiceInterface, pdp services.PdpServiceInterface) *ProductController {
	return &ProductController{
		productService:  ps,
		userLikeService: uls,
		pdpService:      pdp,
		validate:        validator.New(),
	}
}

// Create product (admin only)
// @Summary      Create product
// @Description  Add a new product to the catalog
// @Tags         Products
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CreateProductRequest true "Product details"
// @Success      201 {object} utils.Response{data=dto.ProductResponse}
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      409 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /products [post]
func (ctrl *ProductController) Create(c *gin.Context) {
	var req dto.CreateProductRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	product, err := ctrl.productService.Create(req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to create product")
		return
	}
	_ = ctrl.pdpService.RecordPriceSnapshot(product)
	utils.CreatedResponse(c, constants.MsgCreateSuccess, dto.ToProductResponse(*product))
}

// Update product (admin only)
// @Summary      Update product
// @Description  Update an existing product by its ID
// @Tags         Products
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path int                       true "Product ID"
// @Param        request body dto.UpdateProductRequest true "Product update data"
// @Success      200 {object} utils.Response{data=dto.ProductResponse}
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Failure      409 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /products/{id} [put]
func (ctrl *ProductController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid product id")
		return
	}
	var req dto.UpdateProductRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	existing, err := ctrl.productService.GetByID(uint(id))
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch product")
		return
	}
	oldStock := existing.Stock

	product, err := ctrl.productService.Update(uint(id), req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to update product")
		return
	}
	_ = ctrl.pdpService.RecordPriceSnapshot(product)
	if oldStock == 0 && product.Stock > 0 {
		_ = ctrl.pdpService.NotifyBackInStock(product.ID, product.Name, product.Slug)
	}
	utils.SuccessResponse(c, constants.MsgUpdateSuccess, dto.ToProductResponse(*product))
}

// Delete product (admin only)
// @Summary      Delete product
// @Description  Soft delete a product by its ID
// @Tags         Products
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path int true "Product ID"
// @Success      200 {object} utils.Response
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /products/{id} [delete]
func (ctrl *ProductController) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid product id")
		return
	}
	if err := ctrl.productService.Delete(uint(id)); err != nil {
		utils.HandleServiceError(c, err, "failed to delete product")
		return
	}
	utils.SuccessResponse(c, constants.MsgDeleteSuccess, nil)
}

// GetOne returns a single product by ID or slug, with `is_liked` flag.
// @Summary      Get product by ID or slug
// @Description  Fetch a single product using either its numeric ID or slug
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        id   path string true "Product identifier (ID or slug)"
// @Success      200 {object} utils.Response{data=object{product=dto.ProductResponse,is_liked=bool,stock_subscribed=bool}}
// @Failure      400 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /products/{id} [get]
func (ctrl *ProductController) GetOne(c *gin.Context) {
	identifier := c.Param("id")
	id, parseErr := strconv.ParseUint(identifier, 10, 64)

	var product *models.Product
	var err error
	if parseErr == nil {
		product, err = ctrl.productService.GetByID(uint(id))
	} else {
		product, err = ctrl.productService.GetBySlug(identifier)
	}
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch product")
		return
	}

	isLiked := false
	stockSubscribed := false
	if userID, ok := middleware.GetUserID(c); ok {
		liked, err := ctrl.userLikeService.IsLikedByUser(userID, product.ID)
		if err == nil {
			isLiked = liked
		}
		subscribed, err := ctrl.pdpService.IsStockSubscribed(userID, product.ID)
		if err == nil {
			stockSubscribed = subscribed
		}
	}

	utils.SuccessResponse(c, constants.MsgFetchSuccess, gin.H{
		"product":           dto.ToProductResponse(*product),
		"is_liked":          isLiked,
		"stock_subscribed":  stockSubscribed,
	})
}

// List returns paginated products with filters, each with `is_liked`.
// @Summary      List products
// @Description  Get products with pagination and optional filters
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        limit       query int    false "Items per page (default 20, max 100)" default(20)
// @Param        offset      query int    false "Number of items to skip"              default(0)
// @Param        status      query string false "Product status"                       Enums(active,draft,archived)
// @Param        name        query string false "Filter by product name"
// @Param        sku         query string false "Filter by SKU"
// @Param        category_id query int    false "Filter by category ID"
// @Param        brand_id    query int    false "Filter by brand ID"
// @Param        min_price   query number false "Minimum price"
// @Param        max_price   query number false "Maximum price"
// @Param        min_rating  query number false "Minimum rating (0–5)"
// @Param        max_rating  query number false "Maximum rating"
// @Param        min_reviews query int    false "Minimum review count"
// @Param        max_reviews query int    false "Maximum review count"
// @Param        is_digital  query bool   false "Digital products only"
// @Param        is_new      query bool   false "New products only"
// @Param        sort        query string false "Sort order" Enums(rating_desc,rating_asc,newest,reviews_desc,price_asc,price_desc)
// @Success 200 {object} utils.Response{data=object{products=[]dto.ProductWithLike,total=int,limit=int,offset=int}}
// @Failure      400 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /products [get]
func (ctrl *ProductController) List(c *gin.Context) {
	limit := constants.DefaultLimit
	if l, err := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(constants.DefaultLimit))); err == nil && l >= constants.MinLimit {
		limit = l
	}
	if limit > constants.MaxLimit {
		limit = constants.MaxLimit
	}

	offset := constants.MinOffset
	if o, err := strconv.Atoi(c.DefaultQuery("offset", strconv.Itoa(constants.MinOffset))); err == nil && o >= constants.MinOffset {
		offset = o
	}

	var filters dto.ProductListFilters
	if !utils.BindAndValidateQuery(c, &filters, ctrl.validate) {
		return
	}

	products, total, err := ctrl.productService.List(limit, offset, filters)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list products")
		return
	}

	// Build liked map for authenticated user
	likedMap := make(map[uint]bool)
	if userID, ok := middleware.GetUserID(c); ok {
		likedIDs, err := ctrl.userLikeService.GetUserLikedProductIDs(userID)
		if err == nil {
			for _, id := range likedIDs {
				likedMap[id] = true
			}
		}
	}

	productsWithLike := make([]dto.ProductWithLike, len(products))
	for i, p := range products {
		productsWithLike[i] = dto.ProductWithLike{
			ProductResponse: dto.ToProductResponse(*p),
			IsLiked:         likedMap[p.ID],
		}
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, dto.ProductListData{
		Products: productsWithLike,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	})
}

// BulkCreate creates multiple products (admin only)
// @Summary      Bulk create products
// @Description  Create multiple products in a single request (admin only)
// @Tags         Products
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body []dto.CreateProductRequest true "Array of products"
// @Success      201 {object} utils.Response{data=[]dto.ProductResponse}
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Router       /products/bulk [post]
func (ctrl *ProductController) BulkCreate(c *gin.Context) {
	var reqs []dto.CreateProductRequest
	if !utils.BindAndValidate(c, &reqs, ctrl.validate) {
		return
	}
	if len(reqs) == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "no products provided")
		return
	}
	products, err := ctrl.productService.BulkCreate(reqs)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to bulk create products")
		return
	}
	responses := make([]dto.ProductResponse, len(products))
	for i, p := range products {
		responses[i] = dto.ToProductResponse(*p)
	}
	utils.CreatedResponse(c, "products created successfully", responses)
}

// BulkDelete deletes multiple products (admin only)
// @Summary      Bulk delete products
// @Description  Soft delete multiple products by their IDs (admin only)
// @Tags         Products
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.BulkDeleteProductsRequest true "Product IDs to delete"
// @Success      200 {object} utils.Response
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Router       /products/bulk [delete]
func (ctrl *ProductController) BulkDelete(c *gin.Context) {
	var req dto.BulkDeleteProductsRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	if err := ctrl.productService.BulkDelete(req.ProductIDs); err != nil {
		utils.HandleServiceError(c, err, "failed to delete products")
		return
	}
	utils.SuccessResponse(c, "products deleted successfully", nil)
}

// GetRelated returns related products (public)
// @Summary      Get related products
// @Description  Fetch products from the same category, ordered by rating and review count
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        id    path  int true  "Product ID"
// @Param        limit query int false "Number of related products (default 4, max 10)"
// @Success      200 {object} utils.Response{data=[]dto.ProductResponse}
// @Failure      400 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /products/{id}/related [get]
func (ctrl *ProductController) GetRelated(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid product id")
		return
	}

	limit := 4
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "4")); err == nil && l > 0 {
		if l > 10 {
			limit = 10
		} else {
			limit = l
		}
	}

	products, err := ctrl.productService.GetRelated(uint(id), limit)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch related products")
		return
	}
	responses := make([]dto.ProductResponse, len(products))
	for i, p := range products {
		responses[i] = dto.ToProductResponse(*p)
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, responses)
}

// GetProductSuggestions returns product recommendations based on a list of product IDs.
// @Summary      Get smart product suggestions
// @Description  Returns products from the same categories as the provided product IDs, excluding the products themselves. Useful for cart page "you might also like".
// @Tags         Products
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.SuggestionsRequest true "Product IDs and limit"
// @Success      200 {object} utils.Response{data=[]dto.ProductResponse} "Suggestions fetched"
// @Failure      400 {object} utils.Response "Invalid request"
// @Failure      500 {object} utils.Response "Internal error"
// @Router       /products/suggestions [post]
func (ctrl *ProductController) GetProductSuggestions(c *gin.Context) {
	var req dto.SuggestionsRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	suggestions, err := ctrl.productService.GetSuggestions(req.ProductIDs, req.Limit)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to get suggestions")
		return
	}

	responses := make([]dto.ProductResponse, len(suggestions))
	for i, p := range suggestions {
		responses[i] = dto.ToProductResponse(*p)
	}
	utils.SuccessResponse(c, "suggestions fetched", responses)
}

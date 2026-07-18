package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	appai "github.com/alireza-akbarzadeh/luxe/internal/application/ai"
	appcatalog "github.com/alireza-akbarzadeh/luxe/internal/application/catalog"
	appstore "github.com/alireza-akbarzadeh/luxe/internal/application/store"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type StoreHandler struct {
	commands       *appstore.Commands
	queries        *appstore.Queries
	productService *appcatalog.Service
	aiService      *appai.Service
	vendorInsights vendorDashboardAdapter
	validate       *validator.Validate
}

func NewStoreHandler(
	commands *appstore.Commands,
	queries *appstore.Queries,
	ps *appcatalog.Service,
	aiService *appai.Service,
	vendorInsights vendorDashboardAdapter,
) *StoreHandler {
	return &StoreHandler{
		commands:       commands,
		queries:        queries,
		productService: ps,
		aiService:      aiService,
		vendorInsights: vendorInsights,
		validate:       validator.New(),
	}
}

// ListStores returns paginated stores with filters.
// @Summary      List stores
// @Description  Get stores with pagination, search, location, rating, category filters
// @Tags         Stores
// @Accept       json
// @Produce      json
// @Param        limit         query int     false "Items per page (default 20, max 100)" default(20)
// @Param        offset        query int     false "Number of items to skip" default(0)
// @Param        search        query string  false "Search by name or description"
// @Param        location      query string  false "Filter by location (partial match)"
// @Param        min_rating    query number  false "Minimum rating (0-5)"
// @Param        category_slug query string  false "Filter by category slug"
// @Param        sort_by       query string  false "Sort order (rating, followers, newest)"
// @Success      200 {object} utils.Response{data=object{stores=[]dto.StoreResponse,total=int,limit=int,offset=int}}
// @Failure      400 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /stores [get]
func (ctrl *StoreHandler) ListStores(c *gin.Context) {
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

	var filters dto.StoreFilter
	if !utils.BindAndValidateQuery(c, &filters, ctrl.validate) {
		return
	}
	stores, total, err := ctrl.queries.ListStores(limit, offset, filters)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list stores")
		return
	}
	responses := make([]dto.StoreResponse, len(stores))
	for i, s := range stores {
		responses[i] = dto.ToStoreResponse(c.Request.Context(), s)
	}

	if userID, ok := middleware.GetUserID(c); ok && len(stores) > 0 {
		storeIDs := make([]uint, len(stores))
		for i, s := range stores {
			storeIDs[i] = s.ID
		}
		followed, err := ctrl.queries.GetFollowedStoreIDs(userID, storeIDs)
		if err == nil {
			for i := range responses {
				if followed[responses[i].ID] {
					v := true
					responses[i].IsFollowed = &v
				}
			}
		}
	}

	utils.SuccessResponse(c, constants.MsgFetchSuccess, gin.H{
		"stores": responses,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetStore returns a single store by slug.
// @Summary      Get store by slug
// @Description  Fetch store details including categories and stats
// @Tags         Stores
// @Accept       json
// @Produce      json
// @Param        slug path string true "Store slug"
// @Success      200 {object} utils.Response{data=dto.StoreResponse}
// @Failure      404 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /stores/{slug} [get]
func (ctrl *StoreHandler) GetStore(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "store slug is required")
		return
	}

	store, err := ctrl.queries.GetBySlug(slug)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch store")
		return
	}
	resp := dto.ToStoreResponse(c.Request.Context(), store)

	if userID, ok := middleware.GetUserID(c); ok {
		followed, err := ctrl.queries.IsFollowing(userID, store.ID)
		if err == nil {
			resp.IsFollowed = &followed
		}
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, resp)
}

// GetStoreProducts returns paginated products of a store (by slug).
// @Summary      Get products of a store
// @Description  List products belonging to a store with product filters & pagination
// @Tags         Stores
// @Accept       json
// @Produce      json
// @Param        slug          path   string true  "Store slug"
// @Param        limit         query  int    false "Items per page (default 20, max 100)"
// @Param        offset        query  int    false "Number of items to skip"
// @Param        name          query  string false "Search by product name (partial match)"
// @Param        category_id   query  int    false "Filter by category ID"
// @Param        min_price     query  number false "Minimum price"
// @Param        max_price     query  number false "Maximum price"
// @Param        min_rating    query  number false "Minimum rating"
// @Param        is_digital    query  bool   false "Digital products only"
// @Param        is_new        query  bool   false "New arrivals only"
// @Param        sort          query  string false "Sort order (price_asc,price_desc,rating_desc,newest)"
// @Success      200 {object} utils.Response{data=object{products=[]dto.ProductResponse,total=int,limit=int,offset=int}}
// @Failure      400 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /stores/{slug}/products [get]
func (ctrl *StoreHandler) GetStoreProducts(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "store slug is required")
		return
	}

	store, err := ctrl.queries.GetBySlug(slug)
	if err != nil {
		utils.HandleServiceError(c, err, "store not found")
		return
	}

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

	products, total, err := ctrl.productService.GetByStoreID(store.ID, limit, offset, filters)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch store products")
		return
	}

	responses := make([]dto.ProductResponse, len(products))
	for i, p := range products {
		responses[i] = dto.ToProductResponse(c.Request.Context(), *p)
	}

	utils.SuccessResponse(c, constants.MsgFetchSuccess, gin.H{
		"products": responses,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

// GetStoreAdmin returns a store by ID for admin management (includes non-active).
// @Summary      Get store by ID (admin)
// @Description  Fetch store details by primary key for admin edit screens
// @Tags         Stores
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Store ID"
// @Success      200 {object} utils.Response{data=dto.StoreResponse}
// @Failure      400 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/stores/{id} [get]
func (ctrl *StoreHandler) GetStoreAdmin(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid store id")
		return
	}
	store, err := ctrl.queries.GetByID(uint(id))
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch store")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, dto.ToAdminStoreResponse(c.Request.Context(), store))
}

// ListStoresAdmin returns paginated stores for admin management.
// @Summary      List stores (admin)
// @Description  Paginated store list with status filters for admin review
// @Tags         Stores
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        limit      query int    false "Items per page"
// @Param        offset     query int    false "Offset"
// @Param        search     query string false "Search name or description"
// @Param        status     query string false "Filter by status (pending, active, suspended)"
// @Param        sort_by    query string false "Sort order (newest, oldest)"
// @Success      200 {object} utils.Response{data=object{stores=[]dto.AdminStoreResponse,total=int,limit=int,offset=int}}
// @Failure      401 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/stores [get]
func (ctrl *StoreHandler) ListStoresAdmin(c *gin.Context) {
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

	var filters dto.AdminStoreFilter
	if !utils.BindAndValidateQuery(c, &filters, ctrl.validate) {
		return
	}

	stores, total, err := ctrl.queries.ListAdminStores(limit, offset, filters)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list stores")
		return
	}

	responses := make([]dto.AdminStoreResponse, len(stores))
	for i, store := range stores {
		responses[i] = dto.ToAdminStoreResponse(c.Request.Context(), store)
	}

	utils.SuccessResponse(c, constants.MsgFetchSuccess, gin.H{
		"stores": responses,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetAdminVendorKPIs returns vendor hub KPI counts for admin.
// @Summary      Vendor KPIs (admin)
// @Description  Returns store counts by status and verification for the vendors admin hub
// @Tags         Vendors
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} utils.Response{data=dto.AdminVendorKPIsResponse}
// @Failure      401 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/vendors/kpis [get]
func (ctrl *StoreHandler) GetAdminVendorKPIs(c *gin.Context) {
	kpis, err := ctrl.queries.GetAdminVendorKPIs(c.Request.Context())
	if err != nil {
		utils.HandleServiceError(c, err, "failed to get vendor KPIs")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, kpis)
}

// GetAdminVendorPerformance returns sales and operational metrics for a vendor store.
// @Summary      Vendor performance (admin)
// @Description  Returns revenue, orders, products, and chart data for a vendor store
// @Tags         Vendors
// @Produce      json
// @Security     BearerAuth
// @Param        id     path  int    true  "Store ID"
// @Param        period query string false "Period: 7d, 30d, or 90d (default 30d)"
// @Success      200 {object} utils.Response{data=dto.AdminVendorPerformanceResponse}
// @Failure      400 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/vendors/{id}/performance [get]
func (ctrl *StoreHandler) GetAdminVendorPerformance(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid store id")
		return
	}

	var filters dto.AdminVendorPerformanceFilters
	if !utils.BindAndValidateQuery(c, &filters, ctrl.validate) {
		return
	}

	days := 30
	switch filters.Period {
	case "7d":
		days = 7
	case "90d":
		days = 90
	}

	report, err := ctrl.vendorInsights.LoadAdminPerformance(c.Request.Context(), uint(id), days)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to get vendor performance")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, report)
}

// CreateStore creates a new store (admin only).
// @Summary      Create store
// @Description  Create a new store (admin only)
// @Tags         Stores
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CreateStoreRequest true "Store details"
// @Success      201 {object} utils.Response{data=dto.StoreResponse}
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/stores [post]
func (ctrl *StoreHandler) CreateStore(c *gin.Context) {
	var req dto.CreateStoreRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	store, err := ctrl.commands.Create(req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to create store")
		return
	}
	utils.CreatedResponse(c, constants.MsgCreateSuccess, dto.ToStoreResponse(c.Request.Context(), store))
}

// UpdateStore updates an existing store (admin only).
// @Summary      Update store
// @Description  Update store details (admin only)
// @Tags         Stores
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path int                       true "Store ID"
// @Param        request body dto.UpdateStoreRequest true "Updated store data"
// @Success      200 {object} utils.Response{data=dto.StoreResponse}
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/stores/{id} [put]
func (ctrl *StoreHandler) UpdateStore(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid store id")
		return
	}
	var req dto.UpdateStoreRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	store, err := ctrl.commands.Update(uint(id), req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to update store")
		return
	}
	utils.SuccessResponse(c, constants.MsgUpdateSuccess, dto.ToStoreResponse(c.Request.Context(), store))
}

// DeleteStore deletes a store (admin only).
// @Summary      Delete store
// @Description  Soft delete a store (admin only)
// @Tags         Stores
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Store ID"
// @Success      200 {object} utils.Response
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/stores/{id} [delete]
func (ctrl *StoreHandler) DeleteStore(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid store id")
		return
	}
	if err := ctrl.commands.Delete(uint(id)); err != nil {
		utils.HandleServiceError(c, err, "failed to delete store")
		return
	}
	utils.SuccessResponse(c, constants.MsgDeleteSuccess, nil)
}

// FollowStore godoc
// @Summary      Follow a store
// @Description  Follow a store (authenticated)
// @Tags         Stores
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        slug path string true "Store slug"
// @Success      200 {object} utils.Response
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Router       /stores/{slug}/follow [post]
func (ctrl *StoreHandler) FollowStore(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	slug := c.Param("slug")
	store, err := ctrl.queries.GetBySlug(slug)
	if err != nil {
		utils.HandleServiceError(c, err, "store not found")
		return
	}
	if err := ctrl.commands.FollowStore(userID, store.ID); err != nil {
		utils.HandleServiceError(c, err, "failed to follow store")
		return
	}
	utils.SuccessResponse(c, "followed store successfully", nil)
}

// UnfollowStore godoc
// @Summary      Unfollow a store
// @Description  Unfollow a store (authenticated)
// @Tags         Stores
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        slug path string true "Store slug"
// @Success      200 {object} utils.Response
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Router       /stores/{slug}/follow [delete]
func (ctrl *StoreHandler) UnfollowStore(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	slug := c.Param("slug")
	store, err := ctrl.queries.GetBySlug(slug)
	if err != nil {
		utils.HandleServiceError(c, err, "store not found")
		return
	}
	if err := ctrl.commands.UnfollowStore(userID, store.ID); err != nil {
		utils.HandleServiceError(c, err, "failed to unfollow store")
		return
	}
	utils.SuccessResponse(c, "unfollowed store successfully", nil)
}

func parseStoreReviewPagination(c *gin.Context) (limit, offset int) {
	limit, _ = strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ = strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit < 1 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func (ctrl *StoreHandler) loadStoreBySlug(c *gin.Context) (*models.Store, bool) {
	slug := c.Param("slug")
	if slug == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "store slug is required")
		return nil, false
	}
	store, err := ctrl.queries.GetBySlug(slug)
	if err != nil {
		utils.HandleServiceError(c, err, "store not found")
		return nil, false
	}
	return store, true
}

// GetStoreReviews godoc
// @Summary      List store reviews
// @Description  Paginated reviews and rating summary for a store
// @Tags         Stores
// @Produce      json
// @Param        slug   path  string true  "Store slug"
// @Param        limit  query int    false "Items per page" default(10)
// @Param        offset query int    false "Offset" default(0)
// @Success      200 {object} utils.Response
// @Router       /stores/{slug}/reviews [get]
func (ctrl *StoreHandler) GetStoreReviews(c *gin.Context) {
	store, ok := ctrl.loadStoreBySlug(c)
	if !ok {
		return
	}
	limit, offset := parseStoreReviewPagination(c)

	reviews, total, summary, err := ctrl.queries.ListStoreReviews(store.ID, limit, offset)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch store reviews")
		return
	}

	viewerID, _ := middleware.GetUserID(c)
	responseReviews := make([]dto.StoreReviewResponse, len(reviews))
	for i := range reviews {
		responseReviews[i] = dto.ToStoreReviewResponse(&reviews[i], viewerID)
	}

	utils.SuccessResponse(c, constants.MsgFetchSuccess, gin.H{
		"reviews": responseReviews,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
		"summary": summary,
	})
}

// GetMyStoreReview godoc
// @Summary      Get my store review
// @Description  Returns the authenticated user's review for a store, if any
// @Tags         Stores
// @Produce      json
// @Security     BearerAuth
// @Param        slug path string true "Store slug"
// @Success      200 {object} utils.Response
// @Router       /stores/{slug}/reviews/me [get]
func (ctrl *StoreHandler) GetMyStoreReview(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	store, ok := ctrl.loadStoreBySlug(c)
	if !ok {
		return
	}

	review, err := ctrl.queries.GetUserStoreReview(userID, store.ID)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch review")
		return
	}
	if review == nil {
		utils.SuccessResponse(c, constants.MsgFetchSuccess, nil)
		return
	}

	resp := dto.ToStoreReviewResponse(review, userID)
	utils.SuccessResponse(c, constants.MsgFetchSuccess, resp)
}

// CreateStoreReview godoc
// @Summary      Create store review
// @Description  Leave a rating and comment for a store
// @Tags         Stores
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        slug    path string true "Store slug"
// @Param        request body dto.CreateStoreReviewRequest true "Review data"
// @Success      201 {object} utils.Response
// @Router       /stores/{slug}/reviews [post]
func (ctrl *StoreHandler) CreateStoreReview(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	store, ok := ctrl.loadStoreBySlug(c)
	if !ok {
		return
	}

	var req dto.CreateStoreReviewRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	review, err := ctrl.commands.CreateStoreReview(userID, store.ID, req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to create review")
		return
	}

	utils.CreatedResponse(c, "review submitted", dto.ToStoreReviewResponse(review, userID))
}

// UpdateStoreReview godoc
// @Summary      Update store review
// @Description  Update the authenticated user's store review
// @Tags         Stores
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        slug      path string true "Store slug"
// @Param        reviewId  path int true "Review ID"
// @Param        request   body dto.UpdateStoreReviewRequest true "Updated review"
// @Success      200 {object} utils.Response
// @Router       /stores/{slug}/reviews/{reviewId} [put]
func (ctrl *StoreHandler) UpdateStoreReview(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	if _, ok := ctrl.loadStoreBySlug(c); !ok {
		return
	}

	reviewID, err := strconv.ParseUint(c.Param("reviewId"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid review id")
		return
	}

	var req dto.UpdateStoreReviewRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	review, err := ctrl.commands.UpdateStoreReview(userID, uint(reviewID), req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to update review")
		return
	}

	utils.SuccessResponse(c, constants.MsgUpdateSuccess, dto.ToStoreReviewResponse(review, userID))
}

// DeleteStoreReview godoc
// @Summary      Delete store review
// @Description  Delete the authenticated user's store review
// @Tags         Stores
// @Security     BearerAuth
// @Param        slug     path string true "Store slug"
// @Param        reviewId path int    true "Review ID"
// @Success      200 {object} utils.Response
// @Router       /stores/{slug}/reviews/{reviewId} [delete]
func (ctrl *StoreHandler) DeleteStoreReview(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	if _, ok := ctrl.loadStoreBySlug(c); !ok {
		return
	}

	reviewID, err := strconv.ParseUint(c.Param("reviewId"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid review id")
		return
	}

	if err := ctrl.commands.DeleteStoreReview(userID, uint(reviewID)); err != nil {
		utils.HandleServiceError(c, err, "failed to delete review")
		return
	}

	utils.SuccessResponse(c, constants.MsgDeleteSuccess, nil)
}

// ListVendorStores returns stores the authenticated user can manage in the vendor panel.
// @Summary      List vendor stores
// @Description  Returns stores owned by the current seller, or all stores for admins/moderators
// @Tags         Vendor
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} utils.Response{data=[]dto.StoreResponse}
// @Failure      401 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /vendor/stores [get]
func (ctrl *StoreHandler) ListVendorStores(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	role, _ := middleware.GetUserRole(c)

	stores, err := ctrl.queries.ListVendorStores(c.Request.Context(), userID, role)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list vendor stores")
		return
	}

	responses := make([]dto.VendorStoreResponse, len(stores))
	for i, store := range stores {
		responses[i] = dto.ToVendorStoreResponse(c.Request.Context(), store)
	}

	utils.SuccessResponse(c, constants.MsgFetchSuccess, responses)
}

// GetVendorStore returns a vendor-owned store with settings.
// @Summary      Get vendor store
// @Description  Fetch store details for the authenticated seller
// @Tags         Vendor
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Store ID"
// @Success      200 {object} utils.Response{data=dto.VendorStoreResponse}
// @Failure      401 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Router       /vendor/stores/{id} [get]
func (ctrl *StoreHandler) GetVendorStore(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	storeID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid store id")
		return
	}

	role, _ := middleware.GetUserRole(c)
	store, err := ctrl.queries.GetVendorStore(c.Request.Context(), uint(storeID), userID, role)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch vendor store")
		return
	}

	utils.SuccessResponse(c, constants.MsgFetchSuccess, dto.ToVendorStoreResponse(c.Request.Context(), store))
}

// UpdateVendorStore updates a vendor-owned store profile.
// @Summary      Update vendor store
// @Description  Update storefront and business metadata for the authenticated seller
// @Tags         Vendor
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Store ID"
// @Param        request body dto.VendorUpdateStoreRequest true "Updated store data"
// @Success      200 {object} utils.Response{data=dto.VendorStoreResponse}
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Router       /vendor/stores/{id} [put]
func (ctrl *StoreHandler) UpdateVendorStore(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	storeID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid store id")
		return
	}

	var req dto.VendorUpdateStoreRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	role, _ := middleware.GetUserRole(c)
	store, err := ctrl.commands.UpdateForVendor(c.Request.Context(), uint(storeID), userID, role, req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to update vendor store")
		return
	}

	utils.SuccessResponse(c, constants.MsgUpdateSuccess, dto.ToVendorStoreResponse(c.Request.Context(), store))
}

// CreateVendorStore registers a storefront for the authenticated seller.
// @Summary      Create vendor store
// @Description  Self-service seller onboarding — creates a store owned by the current user
// @Tags         Vendor
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.VendorCreateStoreRequest true "Store and business profile"
// @Success      201 {object} utils.Response{data=dto.StoreResponse}
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /vendor/stores [post]
func (ctrl *StoreHandler) CreateVendorStore(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	var req dto.VendorCreateStoreRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	store, err := ctrl.commands.CreateForVendor(c.Request.Context(), userID, req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to create vendor store")
		return
	}

	utils.CreatedResponse(c, constants.MsgCreateSuccess, dto.ToVendorStoreResponse(c.Request.Context(), store))
}

func (ctrl *StoreHandler) parseVendorProductFilters(c *gin.Context) (dto.ProductListFilters, bool) {
	var filters dto.ProductListFilters
	if !utils.BindAndValidateQuery(c, &filters, ctrl.validate) {
		return filters, false
	}
	if search := c.Query("search"); search != "" {
		filters.Search = search
	}
	return filters, true
}

// ListVendorStoreProducts returns paginated products for a vendor-owned store.
// @Summary      List vendor store products
// @Description  Returns products belonging to the given store with search and status filters.
// @Tags         Vendor
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path   int     true  "Store ID"
// @Param        limit   query  int     false "Items per page" default(20)
// @Param        offset  query  int     false "Offset" default(0)
// @Param        status  query  string  false "Product status"
// @Param        search  query  string  false "Search name, SKU, or barcode"
// @Success      200 {object} utils.Response{data=dto.VendorProductListData}
// @Failure      401 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Router       /vendor/stores/{id}/products [get]
func (ctrl *StoreHandler) ListVendorStoreProducts(c *gin.Context) {
	storeID, _, _, ok := authorizeVendorStore(c, ctrl.queries)
	if !ok {
		return
	}

	limit, offset := paginationParams(c, constants.DefaultLimit)
	filters, ok := ctrl.parseVendorProductFilters(c)
	if !ok {
		return
	}

	products, total, err := ctrl.productService.GetByStoreID(storeID, limit, offset, filters)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch vendor products")
		return
	}

	utils.SuccessResponse(c, constants.MsgFetchSuccess, dto.VendorProductListData{
		Products: dto.ToVendorProductListItems(products),
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	})
}

// GetVendorStoreProductStats returns product count summaries for a vendor store.
// @Summary      Vendor store product stats
// @Description  Returns total products, counts by status, and low-stock count for the vendor dashboard.
// @Tags         Vendor
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Store ID"
// @Success      200 {object} utils.Response{data=dto.VendorProductStatsResponse}
// @Failure      401 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Router       /vendor/stores/{id}/products/stats [get]
func (ctrl *StoreHandler) GetVendorStoreProductStats(c *gin.Context) {
	storeID, _, _, ok := authorizeVendorStore(c, ctrl.queries)
	if !ok {
		return
	}

	stats, err := ctrl.productService.GetVendorStoreProductStats(c.Request.Context(), storeID)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch vendor product stats")
		return
	}

	utils.SuccessResponse(c, constants.MsgFetchSuccess, dto.VendorProductStatsResponse{
		Total:    stats.Total,
		ByStatus: stats.ByStatus,
		LowStock: stats.LowStock,
	})
}

// GetVendorStoreAiDashboard returns AI-powered operational insights for a vendor store.
// @Summary      Vendor AI dashboard briefing
// @Description  Generates prioritized actions, alerts, and opportunities from store orders, catalog, and inventory facts
// @Tags         Vendor
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Store ID"
// @Success      200 {object} utils.Response{data=dto.AiVendorDashboardResponse}
// @Failure      401 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Failure      429 {object} utils.Response
// @Router       /vendor/stores/{id}/ai/dashboard [get]
func (ctrl *StoreHandler) GetVendorStoreAiDashboard(c *gin.Context) {
	storeID, userID, _, ok := authorizeVendorStore(c, ctrl.queries)
	if !ok {
		return
	}

	subjectKey := fmt.Sprintf("%d:%d", userID, storeID)
	result, err := ctrl.aiService.VendorDashboard(
		c.Request.Context(),
		storeID,
		subjectKey,
		ctrl.vendorInsights,
	)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to generate vendor dashboard insights")
		return
	}

	utils.SuccessResponse(c, constants.MsgFetchSuccess, result)
}

// GetVendorStoreAiSalesInsights returns AI-powered sales analytics for a vendor store.
// @Summary      Vendor AI sales insights
// @Description  Revenue metrics, daily series, top products, and AI narrative for the selected period
// @Tags         Vendor
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Store ID"
// @Param        days query int false "Analysis window in days (default 30, max 90)" default(30)
// @Success      200 {object} utils.Response{data=dto.AiVendorSalesInsightsResponse}
// @Failure      401 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Failure      429 {object} utils.Response
// @Router       /vendor/stores/{id}/ai/sales-insights [get]
func (ctrl *StoreHandler) GetVendorStoreAiSalesInsights(c *gin.Context) {
	storeID, userID, _, ok := authorizeVendorStore(c, ctrl.queries)
	if !ok {
		return
	}

	days := parseVendorAnalysisDays(c)
	if days == 0 {
		return
	}

	subjectKey := fmt.Sprintf("%d:%d", userID, storeID)
	result, err := ctrl.aiService.VendorSalesInsights(
		c.Request.Context(),
		storeID,
		days,
		subjectKey,
		ctrl.vendorInsights,
	)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to generate vendor sales insights")
		return
	}

	utils.SuccessResponse(c, constants.MsgFetchSuccess, result)
}

// GetVendorStoreAiInventoryForecast returns AI-powered inventory forecasts for a vendor store.
// @Summary      Vendor AI inventory forecast
// @Description  Stock velocity, days-until-stockout, and replenishment suggestions from catalog and sales
// @Tags         Vendor
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Store ID"
// @Param        days query int false "Analysis window in days (default 30, max 90)" default(30)
// @Success      200 {object} utils.Response{data=dto.AiVendorInventoryForecastResponse}
// @Failure      401 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Failure      429 {object} utils.Response
// @Router       /vendor/stores/{id}/ai/inventory-forecast [get]
func (ctrl *StoreHandler) GetVendorStoreAiInventoryForecast(c *gin.Context) {
	storeID, userID, _, ok := authorizeVendorStore(c, ctrl.queries)
	if !ok {
		return
	}

	days := parseVendorAnalysisDays(c)
	if days == 0 {
		return
	}

	subjectKey := fmt.Sprintf("%d:%d", userID, storeID)
	result, err := ctrl.aiService.VendorInventoryForecast(
		c.Request.Context(),
		storeID,
		days,
		subjectKey,
		ctrl.vendorInsights,
	)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to generate vendor inventory forecast")
		return
	}

	utils.SuccessResponse(c, constants.MsgFetchSuccess, result)
}

// GetVendorStoreAiPricingAssistant returns AI-powered pricing recommendations for a vendor store.
// @Summary      Vendor AI pricing assistant
// @Description  Per-SKU price actions based on demand, margin, and inventory signals
// @Tags         Vendor
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Store ID"
// @Param        days query int false "Analysis window in days (default 30, max 90)" default(30)
// @Success      200 {object} utils.Response{data=dto.AiVendorPricingAssistantResponse}
// @Failure      401 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Failure      429 {object} utils.Response
// @Router       /vendor/stores/{id}/ai/pricing-assistant [get]
func (ctrl *StoreHandler) GetVendorStoreAiPricingAssistant(c *gin.Context) {
	storeID, userID, _, ok := authorizeVendorStore(c, ctrl.queries)
	if !ok {
		return
	}

	days := parseVendorAnalysisDays(c)
	if days == 0 {
		return
	}

	subjectKey := fmt.Sprintf("%d:%d", userID, storeID)
	result, err := ctrl.aiService.VendorPricingAssistant(
		c.Request.Context(),
		storeID,
		days,
		subjectKey,
		ctrl.vendorInsights,
	)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to generate vendor pricing assistant")
		return
	}

	utils.SuccessResponse(c, constants.MsgFetchSuccess, result)
}

// GetVendorStoreAiCustomerSegments returns AI-powered customer segmentation for a vendor store.
// @Summary      Vendor AI customer segments
// @Description  Buyer segments, spend tiers, and retention campaign ideas from order history
// @Tags         Vendor
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Store ID"
// @Param        days query int false "Analysis window in days (default 365, max 730)" default(365)
// @Success      200 {object} utils.Response{data=dto.AiVendorCustomerSegmentsResponse}
// @Failure      401 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Failure      429 {object} utils.Response
// @Router       /vendor/stores/{id}/ai/customer-segments [get]
func (ctrl *StoreHandler) GetVendorStoreAiCustomerSegments(c *gin.Context) {
	storeID, userID, _, ok := authorizeVendorStore(c, ctrl.queries)
	if !ok {
		return
	}

	days := parseVendorCustomerDays(c)
	if days == 0 {
		return
	}

	subjectKey := fmt.Sprintf("%d:%d", userID, storeID)
	result, err := ctrl.aiService.VendorCustomerSegments(
		c.Request.Context(),
		storeID,
		days,
		subjectKey,
		ctrl.vendorInsights,
	)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to generate vendor customer segments")
		return
	}

	utils.SuccessResponse(c, constants.MsgFetchSuccess, result)
}

func parseVendorAnalysisDays(c *gin.Context) int {
	days := 30
	if raw := c.Query("days"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			utils.BadRequestResponse(c, "days must be a positive integer")
			return 0
		}
		days = parsed
	}
	return days
}

func parseVendorCustomerDays(c *gin.Context) int {
	days := 365
	if raw := c.Query("days"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			utils.BadRequestResponse(c, "days must be a positive integer")
			return 0
		}
		days = parsed
	}
	if days > 730 {
		days = 730
	}
	return days
}

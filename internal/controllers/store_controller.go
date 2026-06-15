package controllers

import (
	"net/http"
	"strconv"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/middleware"
	"github.com/alireza-akbarzadeh/luxe/internal/services"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type StoreController struct {
	storeService   services.StoreServiceInterface
	productService services.ProductServiceInterface
	validate       *validator.Validate
}

func NewStoreController(ss services.StoreServiceInterface, ps services.ProductServiceInterface) *StoreController {
	return &StoreController{
		storeService:   ss,
		productService: ps,
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
func (ctrl *StoreController) ListStores(c *gin.Context) {
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
	stores, total, err := ctrl.storeService.ListStores(limit, offset, filters)
	if err != nil {
		utils.HandleAppError(c, err, "failed to list stores")
		return
	}
	responses := make([]dto.StoreResponse, len(stores))
	for i, s := range stores {
		responses[i] = dto.ToStoreResponse(s)
	}

	if userID, ok := middleware.GetUserID(c); ok && len(stores) > 0 {
		storeIDs := make([]uint, len(stores))
		for i, s := range stores {
			storeIDs[i] = s.ID
		}
		followed, err := ctrl.storeService.GetFollowedStoreIDs(userID, storeIDs)
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
func (ctrl *StoreController) GetStore(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "store slug is required")
		return
	}

	store, err := ctrl.storeService.GetBySlug(slug)
	if err != nil {
		utils.HandleAppError(c, err, "failed to fetch store")
		return
	}
	resp := dto.ToStoreResponse(store)

	if userID, ok := middleware.GetUserID(c); ok {
		followed, err := ctrl.storeService.IsFollowing(userID, store.ID)
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
func (ctrl *StoreController) GetStoreProducts(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "store slug is required")
		return
	}

	store, err := ctrl.storeService.GetBySlug(slug)
	if err != nil {
		utils.HandleAppError(c, err, "store not found")
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
		utils.HandleAppError(c, err, "failed to fetch store products")
		return
	}

	responses := make([]dto.ProductResponse, len(products))
	for i, p := range products {
		responses[i] = dto.ToProductResponse(*p)
	}

	utils.SuccessResponse(c, constants.MsgFetchSuccess, gin.H{
		"products": responses,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
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
// @Router       /stores [post]
func (ctrl *StoreController) CreateStore(c *gin.Context) {
	var req dto.CreateStoreRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	store, err := ctrl.storeService.Create(req)
	if err != nil {
		utils.HandleAppError(c, err, "failed to create store")
		return
	}
	utils.CreatedResponse(c, constants.MsgCreateSuccess, dto.ToStoreResponse(store))
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
// @Router       /stores/{id} [put]
func (ctrl *StoreController) UpdateStore(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid store id")
		return
	}
	var req dto.UpdateStoreRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	store, err := ctrl.storeService.Update(uint(id), req)
	if err != nil {
		utils.HandleAppError(c, err, "failed to update store")
		return
	}
	utils.SuccessResponse(c, constants.MsgUpdateSuccess, dto.ToStoreResponse(store))
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
// @Router       /stores/{id} [delete]
func (ctrl *StoreController) DeleteStore(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid store id")
		return
	}
	if err := ctrl.storeService.Delete(uint(id)); err != nil {
		utils.HandleAppError(c, err, "failed to delete store")
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
func (ctrl *StoreController) FollowStore(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	slug := c.Param("slug")
	store, err := ctrl.storeService.GetBySlug(slug)
	if err != nil {
		utils.HandleAppError(c, err, "store not found")
		return
	}
	if err := ctrl.storeService.FollowStore(userID, store.ID); err != nil {
		utils.HandleAppError(c, err, "failed to follow store")
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
func (ctrl *StoreController) UnfollowStore(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	slug := c.Param("slug")
	store, err := ctrl.storeService.GetBySlug(slug)
	if err != nil {
		utils.HandleAppError(c, err, "store not found")
		return
	}
	if err := ctrl.storeService.UnfollowStore(userID, store.ID); err != nil {
		utils.HandleAppError(c, err, "failed to unfollow store")
		return
	}
	utils.SuccessResponse(c, "unfollowed store successfully", nil)
}

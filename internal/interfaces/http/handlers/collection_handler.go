package handlers

import (
	"net/http"

	"github.com/alireza-akbarzadeh/luxe/internal/application/collection"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// CollectionHandler handles collection HTTP requests.
type CollectionHandler struct {
	collectionService *collection.Service
	validate          *validator.Validate
}

// NewCollectionHandler creates a new CollectionHandler.
func NewCollectionHandler(collectionService *collection.Service) *CollectionHandler {
	return &CollectionHandler{
		collectionService: collectionService,
		validate:          validator.New(),
	}
}

// CreateCollection godoc
// @Summary      Create a new collection
// @Description  Creates a curated collection for the storefront collections page.
// @Tags         collections
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateCollectionRequest true "Collection payload"
// @Success      201  {object}  utils.Response{data=dto.CollectionResponse}
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /collections [post]
func (ctrl *CollectionHandler) CreateCollection(c *gin.Context) {
	var req dto.CreateCollectionRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	collection, err := ctrl.collectionService.Create(c.Request.Context(), &req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to create collection")
		return
	}

	utils.CreatedResponse(c, "collection created successfully", collection)
}

// GetCollection godoc
// @Summary      Get a collection by ID
// @Tags         collections
// @Produce      json
// @Param        id   path      int  true  "Collection ID"
// @Success      200  {object}  utils.Response{data=dto.CollectionResponse}
// @Failure      404  {object}  utils.Response
// @Router       /collections/{id} [get]
func (ctrl *CollectionHandler) GetCollection(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	collection, err := ctrl.collectionService.GetByID(c.Request.Context(), id)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to retrieve collection")
		return
	}

	utils.SuccessResponse(c, "collection retrieved", collection)
}

// ListCollections godoc
// @Summary      List collections
// @Description  Returns a paginated list of collections with optional search and status filtering.
// @Tags         collections
// @Produce      json
// @Param        page    query     int     false  "Page number"
// @Param        limit   query     int     false  "Items per page"
// @Param        search  query     string  false  "Search by title, slug, or eyebrow"
// @Param        status  query     string  false  "Filter by status"
// @Param        theme   query     string  false  "Filter by theme (e.g. lifestyle)"
// @Success      200     {object}  dto.CollectionListResponse
// @Failure      500     {object}  utils.Response
// @Router       /collections [get]
func (ctrl *CollectionHandler) ListCollections(c *gin.Context) {
	var req dto.ListCollectionsRequest
	if !utils.BindAndValidateQuery(c, &req, ctrl.validate) {
		return
	}

	collections, total, err := ctrl.collectionService.List(c.Request.Context(), &req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list collections")
		return
	}

	resp := dto.CollectionListResponse{
		BaseResponse: dto.BaseResponse{
			Success: true,
			Message: constants.MsgFetchSuccess,
			Code:    http.StatusOK,
		},
		Data: dto.CollectionListData{
			Collections: collections,
			Total:       total,
			Page:        req.Page,
			Limit:       req.Limit,
		},
	}
	c.JSON(http.StatusOK, resp)
}

// UpdateCollection godoc
// @Summary      Update a collection
// @Tags         collections
// @Accept       json
// @Produce      json
// @Param        id       path      int                        true  "Collection ID"
// @Param        request  body      dto.UpdateCollectionRequest true "Update payload"
// @Success      200      {object}  utils.Response{data=dto.CollectionResponse}
// @Failure      404      {object}  utils.Response
// @Router       /collections/{id} [put]
func (ctrl *CollectionHandler) UpdateCollection(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateCollectionRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	collection, err := ctrl.collectionService.Update(c.Request.Context(), id, &req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to update collection")
		return
	}

	utils.SuccessResponse(c, "collection updated", collection)
}

// DeleteCollection godoc
// @Summary      Delete a collection
// @Tags         collections
// @Produce      json
// @Param        id   path      int  true  "Collection ID"
// @Success      200  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Router       /collections/{id} [delete]
func (ctrl *CollectionHandler) DeleteCollection(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	if err := ctrl.collectionService.Delete(c.Request.Context(), id); err != nil {
		utils.HandleServiceError(c, err, "failed to delete collection")
		return
	}

	utils.SuccessResponse(c, "collection deleted", nil)
}

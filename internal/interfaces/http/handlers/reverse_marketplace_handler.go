package handlers

import (
	appreversemarketplace "github.com/alireza-akbarzadeh/luxe/internal/application/reversemarketplace"
	appstore "github.com/alireza-akbarzadeh/luxe/internal/application/store"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ReverseMarketplaceHandler serves buyer and vendor reverse marketplace endpoints.
type ReverseMarketplaceHandler struct {
	svc           *appreversemarketplace.Service
	storeQueries  *appstore.Queries
	validate      *validator.Validate
}

// NewReverseMarketplaceHandler creates a reverse marketplace handler.
func NewReverseMarketplaceHandler(svc *appreversemarketplace.Service, storeQueries *appstore.Queries) *ReverseMarketplaceHandler {
	return &ReverseMarketplaceHandler{svc: svc, storeQueries: storeQueries, validate: validator.New()}
}

// ListReverseMarketplaceRequests returns open buyer wanted listings.
// @Summary      List reverse marketplace requests
// @Description  Returns open buyer wanted listings for storefront discovery
// @Tags         reverse-marketplace
// @Produce      json
// @Param        limit  query int false "Items per page"
// @Param        offset query int false "Offset"
// @Success      200 {object} utils.Response{data=dto.ReverseMarketplaceRequestListResponse}
// @Router       /reverse-marketplace/requests [get]
func (h *ReverseMarketplaceHandler) ListReverseMarketplaceRequests(c *gin.Context) {
	limit, offset := paginationParams(c, constants.DefaultLimit)
	data, err := h.svc.ListOpen(c.Request.Context(), limit, offset)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load reverse marketplace requests")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, data)
}

// GetReverseMarketplaceRequest returns a request with vendor offers.
// @Summary      Get reverse marketplace request
// @Description  Returns a buyer wanted listing with vendor offers
// @Tags         reverse-marketplace
// @Produce      json
// @Param        id path int true "Request ID"
// @Success      200 {object} utils.Response{data=dto.ReverseMarketplaceRequestResponse}
// @Failure      404 {object} utils.Response
// @Router       /reverse-marketplace/requests/{id} [get]
func (h *ReverseMarketplaceHandler) GetReverseMarketplaceRequest(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	data, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load reverse marketplace request")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, data)
}

// CreateReverseMarketplaceRequest posts a new wanted listing.
// @Summary      Create reverse marketplace request
// @Description  Buyer posts a wanted listing for vendors to respond
// @Tags         reverse-marketplace
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CreateReverseMarketplaceRequest true "Wanted listing"
// @Success      201 {object} utils.Response{data=dto.ReverseMarketplaceRequestResponse}
// @Router       /reverse-marketplace/requests [post]
func (h *ReverseMarketplaceHandler) CreateReverseMarketplaceRequest(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	var req dto.CreateReverseMarketplaceRequest
	if !utils.BindAndValidate(c, &req, h.validate) {
		return
	}

	data, err := h.svc.CreateRequest(c.Request.Context(), userID, req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to create reverse marketplace request")
		return
	}
	utils.CreatedResponse(c, "request created", data)
}

// ListMyReverseMarketplaceRequests returns the authenticated buyer's requests.
// @Summary      List my reverse marketplace requests
// @Description  Returns wanted listings created by the authenticated buyer
// @Tags         reverse-marketplace
// @Produce      json
// @Security     BearerAuth
// @Param        limit  query int false "Items per page"
// @Param        offset query int false "Offset"
// @Success      200 {object} utils.Response{data=dto.ReverseMarketplaceRequestListResponse}
// @Router       /reverse-marketplace/my-requests [get]
func (h *ReverseMarketplaceHandler) ListMyReverseMarketplaceRequests(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	limit, offset := paginationParams(c, constants.DefaultLimit)
	data, err := h.svc.ListForUser(c.Request.Context(), userID, limit, offset)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load your reverse marketplace requests")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, data)
}

// ListVendorReverseMarketplaceRequests lets vendors browse open buyer requests.
// @Summary      Vendor browse reverse marketplace requests
// @Description  Returns open buyer wanted listings for vendor offer submission
// @Tags         Vendor
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Store ID"
// @Param        limit  query int false "Items per page"
// @Param        offset query int false "Offset"
// @Success      200 {object} utils.Response{data=dto.ReverseMarketplaceRequestListResponse}
// @Router       /vendor/stores/{id}/reverse-marketplace/requests [get]
func (h *ReverseMarketplaceHandler) ListVendorReverseMarketplaceRequests(c *gin.Context) {
	if _, _, _, ok := authorizeVendorStore(c, h.storeQueries); !ok {
		return
	}

	limit, offset := paginationParams(c, constants.DefaultLimit)
	data, err := h.svc.ListOpen(c.Request.Context(), limit, offset)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load reverse marketplace requests")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, data)
}

// CreateVendorReverseMarketplaceOffer submits a vendor offer on a buyer request.
// @Summary      Submit reverse marketplace offer
// @Description  Vendor store responds to a buyer wanted listing with price and message
// @Tags         Vendor
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Store ID"
// @Param        requestId path int true "Request ID"
// @Param        request body dto.CreateReverseMarketplaceOfferRequest true "Offer details"
// @Success      201 {object} utils.Response{data=dto.ReverseMarketplaceOfferResponse}
// @Router       /vendor/stores/{id}/reverse-marketplace/requests/{requestId}/offers [post]
func (h *ReverseMarketplaceHandler) CreateVendorReverseMarketplaceOffer(c *gin.Context) {
	storeID, userID, _, ok := authorizeVendorStore(c, h.storeQueries)
	if !ok {
		return
	}

	requestID, ok := parseUintParam(c, "requestId")
	if !ok {
		return
	}

	var req dto.CreateReverseMarketplaceOfferRequest
	if !utils.BindAndValidate(c, &req, h.validate) {
		return
	}

	data, err := h.svc.SubmitOffer(c.Request.Context(), storeID, userID, requestID, req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to submit reverse marketplace offer")
		return
	}
	utils.CreatedResponse(c, "offer submitted", data)
}

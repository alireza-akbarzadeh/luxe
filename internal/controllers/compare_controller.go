package controllers

import (
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/middleware"
	"github.com/alireza-akbarzadeh/luxe/internal/services"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type CompareController struct {
	compareService services.CompareServiceInterface
	validate       *validator.Validate
}

func NewCompareController(ps services.CompareServiceInterface) *CompareController {
	return &CompareController{
		compareService: ps,
		validate:       validator.New(),
	}
}

// CompareProducts godoc
// @Summary      Compare products
// @Description  Get product details for comparison (2–4 products)
// @Tags         Compare
// @Accept       json
// @Produce      json
// @Param        request body dto.CompareRequest true "Product IDs"
// @Success      200 {object} utils.Response{data=[]dto.CompareProductResponse}
// @Failure      400 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Router       /compare [post]
func (ctrl *CompareController) CompareProducts(c *gin.Context) {
	var req dto.CompareRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	products, err := ctrl.compareService.GetForCompare(c.Request.Context(), req.ProductIDs)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch products for comparison")
		return
	}
	utils.SuccessResponse(c, "products fetched", products)
}

// GetCompareList godoc
// @Summary      Get user's compare list
// @Description  Returns the list of product IDs in the compare list (authenticated or guest via session_id)
// @Tags         Compare
// @Accept       json
// @Produce      json
// @Security     BearerAuth (optional)
// @Param        session_id query string false "Session ID for guest users"
// @Success      200 {object} utils.Response{data=dto.CompareListResponse}
// @Failure      401 {object} utils.Response
// @Router       /compare [get]
func (ctrl *CompareController) GetCompareList(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	productIDs, err := ctrl.compareService.GetCompareList(userID)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to get compare list")
		return
	}
	utils.SuccessResponse(c, "compare list fetched", dto.CompareListResponse{ProductIDs: productIDs})
}

// SyncCompareList godoc
// @Summary      Sync compare list
// @Description  Save the user's compare list (replaces existing)
// @Tags         Compare
// @Accept       json
// @Produce      json
// @Security     BearerAuth (optional)
// @Param        session_id query string false "Session ID for guest users"
// @Param        request body dto.SyncCompareRequest true "Product IDs"
// @Success      200 {object} utils.Response
// @Failure      400 {object} utils.Response
// @Router       /compare [put]
func (ctrl *CompareController) SyncCompareList(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	var req dto.SyncCompareRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	if err := ctrl.compareService.SyncCompareList(userID, req.ProductIDs); err != nil {
		utils.HandleServiceError(c, err, "failed to sync compare list")
		return
	}
	utils.SuccessResponse(c, "compare list synced", nil)
}

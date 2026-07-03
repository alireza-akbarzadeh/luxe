package handlers

import (
	"net/http"
	"strconv"

	"github.com/alireza-akbarzadeh/luxe/internal/application/bundle"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// BundleHandler serves smart bundle suggestion endpoints.
type BundleHandler struct {
	svc      *bundle.Service
	validate *validator.Validate
}

// NewBundleHandler creates a bundle handler.
func NewBundleHandler(svc *bundle.Service) *BundleHandler {
	return &BundleHandler{svc: svc, validate: validator.New()}
}

// GetProductSmartBundles returns compatibility-ranked bundles for a product.
// @Summary      Smart bundles for a product
// @Description  Returns dynamic product bundles based on compatibility and shopper intent
// @Tags         Bundles
// @Produce      json
// @Param        id     path   int    true  "Product ID"
// @Param        intent query  string false "Shopper intent (everyday, workspace, travel, gift)"
// @Param        limit  query  int    false "Max bundles (default 3)"
// @Success      200    {object} utils.Response{data=dto.SmartBundlesResponse}
// @Failure      404    {object} utils.Response
// @Router       /products/{id}/smart-bundles [get]
func (h *BundleHandler) GetProductSmartBundles(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || productID == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid product id")
		return
	}

	limit := 3
	if raw := c.Query("limit"); raw != "" {
		if parsed, parseErr := strconv.Atoi(raw); parseErr == nil {
			limit = parsed
		}
	}

	result, err := h.svc.SuggestForProduct(c.Request.Context(), uint(productID), c.Query("intent"), limit)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to suggest bundles")
		return
	}

	utils.SuccessResponse(c, "smart bundles", result)
}

// SuggestSmartBundles generates bundles for cart or multi-product contexts.
// @Summary      Suggest smart bundles
// @Description  Returns bundles anchored on one or more product IDs (e.g. cart items)
// @Tags         Bundles
// @Accept       json
// @Produce      json
// @Param        request body dto.SuggestSmartBundlesRequest true "Anchor product IDs"
// @Success      200 {object} utils.Response{data=dto.SmartBundlesResponse}
// @Failure      400 {object} utils.Response
// @Router       /bundles/suggest [post]
func (h *BundleHandler) SuggestSmartBundles(c *gin.Context) {
	var req dto.SuggestSmartBundlesRequest
	if !utils.BindAndValidate(c, &req, h.validate) {
		return
	}

	result, err := h.svc.SuggestForCart(c.Request.Context(), req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to suggest bundles")
		return
	}

	utils.SuccessResponse(c, "smart bundles", result)
}

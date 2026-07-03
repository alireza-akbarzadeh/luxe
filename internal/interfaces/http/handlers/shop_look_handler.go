package handlers

import (
	"net/http"
	"strconv"

	appshoplook "github.com/alireza-akbarzadeh/luxe/internal/application/shoplook"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
)

// ShopLookHandler serves storefront shop-the-look endpoints.
type ShopLookHandler struct {
	svc *appshoplook.Service
}

// NewShopLookHandler creates a shop look handler.
func NewShopLookHandler(svc *appshoplook.Service) *ShopLookHandler {
	return &ShopLookHandler{svc: svc}
}

// ListShopLooks returns active shoppable scenes.
// @Summary      List shop looks
// @Description  Returns active shop-the-look scenes for homepage and listing pages
// @Tags         shop-looks
// @Produce      json
// @Param        limit query int false "Max items (default 12, max 24)"
// @Success      200 {object} utils.Response{data=dto.ShopLookListResponse}
// @Router       /shop-looks [get]
func (h *ShopLookHandler) ListShopLooks(c *gin.Context) {
	limit := 12
	if raw := c.Query("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}

	data, err := h.svc.List(c.Request.Context(), limit)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load shop looks")
		return
	}

	utils.SuccessResponse(c, "shop looks", data)
}

// GetShopLook returns a shop look by slug with tagged products.
// @Summary      Get shop look
// @Description  Returns a shoppable scene with hotspot coordinates and product cards
// @Tags         shop-looks
// @Produce      json
// @Param        slug path string true "Shop look slug"
// @Success      200 {object} utils.Response{data=dto.ShopLookResponse}
// @Failure      404 {object} utils.Response
// @Router       /shop-looks/{slug} [get]
func (h *ShopLookHandler) GetShopLook(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "slug is required")
		return
	}

	data, err := h.svc.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load shop look")
		return
	}

	utils.SuccessResponse(c, "shop look", data)
}

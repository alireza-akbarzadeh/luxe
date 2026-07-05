package handlers

import (
	"net/http"
	"strconv"

	appcreator "github.com/alireza-akbarzadeh/luxe/internal/application/creatorstorefront"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
)

// CreatorStorefrontHandler serves public creator storefront endpoints.
type CreatorStorefrontHandler struct {
	svc *appcreator.Service
}

// NewCreatorStorefrontHandler creates a creator storefront handler.
func NewCreatorStorefrontHandler(svc *appcreator.Service) *CreatorStorefrontHandler {
	return &CreatorStorefrontHandler{svc: svc}
}

// ListCreators returns active creator storefront profiles.
// @Summary      List creator storefronts
// @Description  Returns active creator profiles for community discovery pages
// @Tags         creators
// @Produce      json
// @Param        limit query int false "Max items (default 12, max 24)"
// @Success      200 {object} utils.Response{data=dto.CreatorStorefrontListResponse}
// @Router       /creators [get]
func (h *CreatorStorefrontHandler) ListCreators(c *gin.Context) {
	limit := 12
	if raw := c.Query("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}

	data, err := h.svc.List(c.Request.Context(), limit)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load creators")
		return
	}

	utils.SuccessResponse(c, "creators", data)
}

// GetCreator returns a creator storefront by slug with curated picks.
// @Summary      Get creator storefront
// @Description  Returns a creator profile with curated product picks
// @Tags         creators
// @Produce      json
// @Param        slug path string true "Creator slug"
// @Success      200 {object} utils.Response{data=dto.CreatorStorefrontResponse}
// @Failure      404 {object} utils.Response
// @Router       /creators/{slug} [get]
func (h *CreatorStorefrontHandler) GetCreator(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "slug is required")
		return
	}

	data, err := h.svc.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load creator")
		return
	}

	utils.SuccessResponse(c, "creator", data)
}

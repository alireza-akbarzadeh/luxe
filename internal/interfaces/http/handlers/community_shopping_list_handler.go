package handlers

import (
	"net/http"
	"strconv"

	appcommunitylist "github.com/alireza-akbarzadeh/luxe/internal/application/communityshoppinglist"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
)

// CommunityShoppingListHandler serves public community shopping list endpoints.
type CommunityShoppingListHandler struct {
	svc *appcommunitylist.Service
}

// NewCommunityShoppingListHandler creates a community shopping list handler.
func NewCommunityShoppingListHandler(svc *appcommunitylist.Service) *CommunityShoppingListHandler {
	return &CommunityShoppingListHandler{svc: svc}
}

// ListCommunityLists returns active community shopping lists.
// @Summary      List community shopping lists
// @Description  Returns active community-curated shopping lists for discovery pages
// @Tags         community-lists
// @Produce      json
// @Param        limit query int false "Max items (default 12, max 24)"
// @Success      200 {object} utils.Response{data=dto.CommunityShoppingListListResponse}
// @Router       /community-lists [get]
func (h *CommunityShoppingListHandler) ListCommunityLists(c *gin.Context) {
	limit := 12
	if raw := c.Query("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}

	data, err := h.svc.List(c.Request.Context(), limit)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load community lists")
		return
	}

	utils.SuccessResponse(c, "community lists", data)
}

// GetCommunityList returns a community shopping list by slug with items.
// @Summary      Get community shopping list
// @Description  Returns a community shopping list with shoppable product items
// @Tags         community-lists
// @Produce      json
// @Param        slug path string true "List slug"
// @Success      200 {object} utils.Response{data=dto.CommunityShoppingListResponse}
// @Failure      404 {object} utils.Response
// @Router       /community-lists/{slug} [get]
func (h *CommunityShoppingListHandler) GetCommunityList(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "slug is required")
		return
	}

	data, err := h.svc.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load community list")
		return
	}

	utils.SuccessResponse(c, "community list", data)
}

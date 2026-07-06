package handlers

import (
	"net/http"
	"strconv"

	apppubliccollection "github.com/alireza-akbarzadeh/luxe/internal/application/publiccollection"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
)

// PublicCollectionHandler serves public community collection endpoints.
type PublicCollectionHandler struct {
	svc *apppubliccollection.Service
}

// NewPublicCollectionHandler creates a public collection handler.
func NewPublicCollectionHandler(svc *apppubliccollection.Service) *PublicCollectionHandler {
	return &PublicCollectionHandler{svc: svc}
}

// ListPublicCollections returns active community public collections.
// @Summary      List public collections
// @Description  Returns active community-published product collections for discovery pages
// @Tags         public-collections
// @Produce      json
// @Param        limit query int false "Max items (default 12, max 24)"
// @Success      200 {object} utils.Response{data=dto.PublicCollectionListResponse}
// @Router       /public-collections [get]
func (h *PublicCollectionHandler) ListPublicCollections(c *gin.Context) {
	limit := 12
	if raw := c.Query("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}

	data, err := h.svc.List(c.Request.Context(), limit)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load public collections")
		return
	}

	utils.SuccessResponse(c, "public collections", data)
}

// GetPublicCollection returns a public collection by slug with items.
// @Summary      Get public collection
// @Description  Returns a community public collection with shoppable product items
// @Tags         public-collections
// @Produce      json
// @Param        slug path string true "Collection slug"
// @Success      200 {object} utils.Response{data=dto.PublicCollectionResponse}
// @Failure      404 {object} utils.Response
// @Router       /public-collections/{slug} [get]
func (h *PublicCollectionHandler) GetPublicCollection(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "slug is required")
		return
	}

	data, err := h.svc.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load public collection")
		return
	}

	utils.SuccessResponse(c, "public collection", data)
}

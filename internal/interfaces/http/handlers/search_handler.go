package handlers

import (
	"net/http"
	"strconv"

	appsearch "github.com/alireza-akbarzadeh/luxe/internal/application/search"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	commands *appsearch.Commands
	queries  *appsearch.Queries
}

func NewSearchHandler(commands *appsearch.Commands, queries *appsearch.Queries) *SearchHandler {
	return &SearchHandler{commands: commands, queries: queries}
}

// GlobalSearch godoc
// @Summary      Global search
// @Description  Search products, stores, and categories with advanced filters
// @Tags         Search
// @Accept       json
// @Produce      json
// @Param        q             query string false "Search query (optional for filter-only browse)"
// @Param        limit         query int    false "Items per page (default 10, max 50)"
// @Param        offset        query int    false "Offset for pagination"
// @Param        category_id   query int    false "Filter by category ID"
// @Param        category_slug query string false "Filter by category slug"
// @Param        store_id      query int    false "Filter by store ID"
// @Param        min_price     query number false "Minimum price"
// @Param        max_price     query number false "Maximum price"
// @Param        min_rating    query number false "Minimum rating"
// @Param        is_digital    query bool   false "Digital products only"
// @Param        is_new        query bool   false "New arrivals only"
// @Param        in_stock      query bool   false "In-stock products only"
// @Param        on_sale       query bool   false "On-sale products only"
// @Param        sort          query string false "Sort order (price_asc,price_desc,rating_desc,newest,popular)"
// @Success      200 {object} utils.Response{data=dto.SearchResponse}
// @Failure      400 {object} utils.Response
// @Router       /search [get]
func (ctrl *SearchHandler) GlobalSearch(c *gin.Context) {
	var req dto.SearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid query parameters")
		return
	}
	// Enforce pagination limits
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 50 {
		req.Limit = 50
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	// Log the search (async)
	var userIDPtr *uint
	if userID, ok := middleware.GetUserID(c); ok {
		userIDPtr = &userID
	}
	if req.Query != "" {
		go ctrl.commands.LogSearch(req.Query, userIDPtr)
	}

	result, err := ctrl.queries.GlobalSearch(c.Request.Context(), req)
	if err != nil {
		utils.HandleServiceError(c, err, "search failed")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, result)
}

// Suggestions godoc
// @Summary      Search suggestions
// @Description  Get autocomplete suggestions (products, stores, categories)
// @Tags         Search
// @Accept       json
// @Produce      json
// @Param        q     query string true "Partial query"
// @Param        limit query int    false "Max suggestions (default 8, max 20)"
// @Success      200 {object} utils.Response{data=dto.SuggestionsResponse}
// @Failure      400 {object} utils.Response
// @Router       /search/suggestions [get]
func (ctrl *SearchHandler) Suggestions(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "query parameter 'q' is required")
		return
	}
	limit := 8
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "8")); err == nil && l > 0 {
		if l > 20 {
			limit = 20
		} else {
			limit = l
		}
	}
	suggestions, err := ctrl.queries.Suggestions(c.Request.Context(), q, limit)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to get suggestions")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, suggestions)
}

// Trending godoc
// @Summary      Trending searches
// @Description  Get most popular search queries from the last 7 days
// @Tags         Search
// @Accept       json
// @Produce      json
// @Param        limit query int false "Number of trending terms (default 10, max 20)"
// @Success      200 {object} utils.Response{data=dto.TrendingResponse}
// @Router       /search/trending [get]
func (ctrl *SearchHandler) Trending(c *gin.Context) {
	limit := 10
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "10")); err == nil && l > 0 {
		if l > 20 {
			limit = 20
		} else {
			limit = l
		}
	}
	trending, err := ctrl.queries.Trending(limit)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to get trending searches")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, gin.H{"trending": trending})
}

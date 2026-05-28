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
)

type SearchController struct {
	searchService services.SearchServiceInterface
}

func NewSearchController(ss services.SearchServiceInterface) *SearchController {
	return &SearchController{searchService: ss}
}

// GlobalSearch godoc
// @Summary      Global search
// @Description  Search products, stores, and categories with advanced filters
// @Tags         Search
// @Accept       json
// @Produce      json
// @Param        q            query string true  "Search query"
// @Param        limit        query int    false "Items per page (default 10, max 50)"
// @Param        offset       query int    false "Offset for pagination"
// @Param        category_id  query int    false "Filter by category ID"
// @Param        category_slug query string false "Filter by category slug"
// @Param        min_price    query number false "Minimum price"
// @Param        max_price    query number false "Maximum price"
// @Param        min_rating   query number false "Minimum rating"
// @Param        is_digital   query bool   false "Digital products only"
// @Param        is_new       query bool   false "New arrivals only"
// @Param        sort         query string false "Sort order (price_asc,price_desc,rating_desc,newest)"
// @Success      200 {object} utils.Response{data=dto.SearchResponse}
// @Failure      400 {object} utils.Response
// @Router       /search [get]
func (ctrl *SearchController) GlobalSearch(c *gin.Context) {
	var req dto.SearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid query parameters")
		return
	}
	if req.Query == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "query parameter 'q' is required")
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
	go ctrl.searchService.LogSearch(req.Query, userIDPtr)

	result, err := ctrl.searchService.GlobalSearch(req)
	if err != nil {
		utils.HandleAppError(c, err, "search failed")
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
func (ctrl *SearchController) Suggestions(c *gin.Context) {
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
	suggestions, err := ctrl.searchService.Suggestions(q, limit)
	if err != nil {
		utils.HandleAppError(c, err, "failed to get suggestions")
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
func (ctrl *SearchController) Trending(c *gin.Context) {
	limit := 10
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "10")); err == nil && l > 0 {
		if l > 20 {
			limit = 20
		} else {
			limit = l
		}
	}
	trending, err := ctrl.searchService.Trending(limit)
	if err != nil {
		utils.HandleAppError(c, err, "failed to get trending searches")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, gin.H{"trending": trending})
}

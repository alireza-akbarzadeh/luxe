package handlers

import (
	"net/http"
	"strconv"

	"github.com/alireza-akbarzadeh/luxe/internal/application/home"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// HomeHandler serves storefront homepage aggregation endpoints.
type HomeHandler struct {
	svc      *home.Service
	validate *validator.Validate
}

// NewHomeHandler creates a homepage handler.
func NewHomeHandler(svc *home.Service) *HomeHandler {
	return &HomeHandler{svc: svc, validate: validator.New()}
}

func parseHomeLimit(c *gin.Context) int {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "12"))
	return limit
}

func optionalUserID(c *gin.Context) *uint {
	id, ok := middleware.GetUserID(c)
	if !ok {
		return nil
	}
	return &id
}

// Manifest returns available homepage section keys.
// @Summary      Homepage manifest
// @Description  Lightweight list of homepage section keys and cache TTL
// @Tags         Home
// @Produce      json
// @Success      200 {object} utils.Response{data=dto.HomeManifestResponse}
// @Router       /home [get]
func (h *HomeHandler) Manifest(c *gin.Context) {
	utils.SuccessResponse(c, "homepage manifest", h.svc.Manifest())
}

// GetCategories returns popular, featured, and personalized categories.
// @Summary      Homepage categories
// @Tags         Home
// @Produce      json
// @Param        limit query int false "Max items"
// @Success      200 {object} utils.Response{data=dto.HomeCategoriesResponse}
// @Router       /home/categories [get]
func (h *HomeHandler) GetCategories(c *gin.Context) {
	data, err := h.svc.GetCategories(c.Request.Context(), optionalUserID(c), parseHomeLimit(c))
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load homepage categories")
		return
	}
	utils.SuccessResponse(c, "homepage categories", data)
}

// GetTopBrands returns sales-ranked brands.
// @Summary      Top brands
// @Tags         Home
// @Produce      json
// @Param        limit query int false "Max items"
// @Success      200 {object} utils.Response{data=dto.HomeBrandsResponse}
// @Router       /home/top-brands [get]
func (h *HomeHandler) GetTopBrands(c *gin.Context) {
	brands, err := h.svc.GetTopBrands(c.Request.Context(), parseHomeLimit(c))
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load top brands")
		return
	}
	utils.SuccessResponse(c, "top brands", dto.HomeBrandsResponse{Brands: brands})
}

// GetTopProducts returns best-selling products.
// @Summary      Top products
// @Tags         Home
// @Produce      json
// @Param        limit query int false "Max items"
// @Success      200 {object} utils.Response{data=dto.HomeProductsResponse}
// @Router       /home/top-products [get]
func (h *HomeHandler) GetTopProducts(c *gin.Context) {
	products, err := h.svc.GetTopProducts(c.Request.Context(), parseHomeLimit(c))
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load top products")
		return
	}
	utils.SuccessResponse(c, "top products", dto.HomeProductsResponse{Products: products})
}

// GetTrendingProducts returns trending products.
// @Summary      Trending products
// @Tags         Home
// @Produce      json
// @Param        limit query int false "Max items"
// @Success      200 {object} utils.Response{data=dto.HomeProductsResponse}
// @Router       /home/trending-products [get]
func (h *HomeHandler) GetTrendingProducts(c *gin.Context) {
	products, err := h.svc.GetTrendingProducts(c.Request.Context(), parseHomeLimit(c))
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load trending products")
		return
	}
	utils.SuccessResponse(c, "trending products", dto.HomeProductsResponse{Products: products})
}

// GetNewArrivals returns newest products.
// @Summary      New arrivals
// @Tags         Home
// @Produce      json
// @Param        limit query int false "Max items"
// @Success      200 {object} utils.Response{data=dto.HomeProductsResponse}
// @Router       /home/new-arrivals [get]
func (h *HomeHandler) GetNewArrivals(c *gin.Context) {
	products, err := h.svc.GetNewArrivals(c.Request.Context(), parseHomeLimit(c))
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load new arrivals")
		return
	}
	utils.SuccessResponse(c, "new arrivals", dto.HomeProductsResponse{Products: products})
}

// GetFlashDeals returns active flash deals.
// @Summary      Flash deals
// @Tags         Home
// @Produce      json
// @Param        limit query int false "Max items"
// @Success      200 {object} utils.Response{data=dto.HomeFlashDealsResponse}
// @Router       /home/flash-deals [get]
func (h *HomeHandler) GetFlashDeals(c *gin.Context) {
	deals, err := h.svc.GetFlashDeals(c.Request.Context(), parseHomeLimit(c))
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load flash deals")
		return
	}
	utils.SuccessResponse(c, "flash deals", dto.HomeFlashDealsResponse{Deals: deals})
}

// GetRecommended returns personalized recommendations for the logged-in user.
// @Summary      Recommended products
// @Tags         Home
// @Produce      json
// @Security     BearerAuth
// @Param        limit query int false "Max items"
// @Success      200 {object} utils.Response{data=dto.HomeProductsResponse}
// @Router       /home/recommended [get]
func (h *HomeHandler) GetRecommended(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, "authentication required")
		return
	}
	products, err := h.svc.GetRecommended(c.Request.Context(), userID, parseHomeLimit(c))
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load recommendations")
		return
	}
	utils.SuccessResponse(c, "recommended products", dto.HomeProductsResponse{Products: products})
}

// GetRecentlyViewed returns recently viewed products for the logged-in user.
// @Summary      Recently viewed products
// @Tags         Home
// @Produce      json
// @Security     BearerAuth
// @Param        limit query int false "Max items"
// @Success      200 {object} utils.Response{data=dto.HomeProductsResponse}
// @Router       /home/recently-viewed [get]
func (h *HomeHandler) GetRecentlyViewed(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, "authentication required")
		return
	}
	products, err := h.svc.GetRecentlyViewed(c.Request.Context(), userID, parseHomeLimit(c))
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load recently viewed")
		return
	}
	utils.SuccessResponse(c, "recently viewed", dto.HomeProductsResponse{Products: products})
}

// GetMostWishlisted returns most wishlisted products.
// @Summary      Most wishlisted products
// @Tags         Home
// @Produce      json
// @Param        limit query int false "Max items"
// @Success      200 {object} utils.Response{data=dto.HomeProductsResponse}
// @Router       /home/most-wishlisted [get]
func (h *HomeHandler) GetMostWishlisted(c *gin.Context) {
	products, err := h.svc.GetMostWishlisted(c.Request.Context(), parseHomeLimit(c))
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load wishlisted products")
		return
	}
	utils.SuccessResponse(c, "most wishlisted", dto.HomeProductsResponse{Products: products})
}

// GetCustomerFavorites returns highest-rated products.
// @Summary      Customer favorites
// @Tags         Home
// @Produce      json
// @Param        limit query int false "Max items"
// @Success      200 {object} utils.Response{data=dto.HomeProductsResponse}
// @Router       /home/customer-favorites [get]
func (h *HomeHandler) GetCustomerFavorites(c *gin.Context) {
	products, err := h.svc.GetCustomerFavorites(c.Request.Context(), parseHomeLimit(c))
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load customer favorites")
		return
	}
	utils.SuccessResponse(c, "customer favorites", dto.HomeProductsResponse{Products: products})
}

// GetPopularCollections returns curated collections.
// @Summary      Popular collections
// @Tags         Home
// @Produce      json
// @Param        limit query int false "Max items"
// @Success      200 {object} utils.Response{data=dto.HomeCollectionsResponse}
// @Router       /home/popular-collections [get]
func (h *HomeHandler) GetPopularCollections(c *gin.Context) {
	collections, err := h.svc.GetPopularCollections(c.Request.Context(), parseHomeLimit(c))
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load collections")
		return
	}
	utils.SuccessResponse(c, "popular collections", dto.HomeCollectionsResponse{Collections: collections})
}

// GetFeaturedStores returns top stores.
// @Summary      Featured stores
// @Tags         Home
// @Produce      json
// @Param        limit query int false "Max items"
// @Success      200 {object} utils.Response{data=dto.HomeStoresResponse}
// @Router       /home/featured-stores [get]
func (h *HomeHandler) GetFeaturedStores(c *gin.Context) {
	stores, err := h.svc.GetFeaturedStores(c.Request.Context(), parseHomeLimit(c))
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load featured stores")
		return
	}
	utils.SuccessResponse(c, "featured stores", dto.HomeStoresResponse{Stores: stores})
}

// GetShopByPrice returns shop-by-price buckets.
// @Summary      Shop by price
// @Tags         Home
// @Produce      json
// @Success      200 {object} utils.Response{data=dto.HomeShopByPriceResponse}
// @Router       /home/shop-by-price [get]
func (h *HomeHandler) GetShopByPrice(c *gin.Context) {
	buckets, err := h.svc.GetShopByPrice(c.Request.Context())
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load price buckets")
		return
	}
	utils.SuccessResponse(c, "shop by price", dto.HomeShopByPriceResponse{Buckets: buckets})
}

// GetSeasonalPicks returns seasonal homepage sections.
// @Summary      Seasonal picks
// @Tags         Home
// @Produce      json
// @Param        limit query int false "Max items"
// @Success      200 {object} utils.Response{data=dto.HomeSectionsResponse}
// @Router       /home/seasonal-picks [get]
func (h *HomeHandler) GetSeasonalPicks(c *gin.Context) {
	sections, err := h.svc.GetSeasonalPicks(c.Request.Context(), parseHomeLimit(c))
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load seasonal picks")
		return
	}
	utils.SuccessResponse(c, "seasonal picks", dto.HomeSectionsResponse{Sections: sections})
}

// GetRecentlyRestocked returns recently restocked products.
// @Summary      Recently restocked
// @Tags         Home
// @Produce      json
// @Param        limit query int false "Max items"
// @Success      200 {object} utils.Response{data=dto.HomeProductsResponse}
// @Router       /home/recently-restocked [get]
func (h *HomeHandler) GetRecentlyRestocked(c *gin.Context) {
	products, err := h.svc.GetRecentlyRestocked(c.Request.Context(), parseHomeLimit(c))
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load restocked products")
		return
	}
	utils.SuccessResponse(c, "recently restocked", dto.HomeProductsResponse{Products: products})
}

// GetFavoriteCategories returns the user's favorite categories.
// @Summary      Get favorite categories
// @Tags         Home
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} utils.Response{data=dto.FavoriteCategoriesResponse}
// @Router       /users/me/favorite-categories [get]
func (h *HomeHandler) GetFavoriteCategories(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, "authentication required")
		return
	}
	categories, err := h.svc.GetFavoriteCategories(c.Request.Context(), userID)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load favorite categories")
		return
	}
	utils.SuccessResponse(c, "favorite categories", dto.FavoriteCategoriesResponse{Categories: categories})
}

// SetFavoriteCategories replaces the user's favorite categories.
// @Summary      Set favorite categories
// @Tags         Home
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.SetFavoriteCategoriesRequest true "Category IDs"
// @Success      200 {object} utils.Response{data=dto.FavoriteCategoriesResponse}
// @Router       /users/me/favorite-categories [post]
func (h *HomeHandler) SetFavoriteCategories(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, "authentication required")
		return
	}
	var req dto.SetFavoriteCategoriesRequest
	if !utils.BindAndValidate(c, &req, h.validate) {
		return
	}
	if err := h.svc.SetFavoriteCategories(c.Request.Context(), userID, req.CategoryIDs); err != nil {
		utils.HandleServiceError(c, err, "failed to save favorite categories")
		return
	}
	categories, err := h.svc.GetFavoriteCategories(c.Request.Context(), userID)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load favorite categories")
		return
	}
	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "favorite categories updated",
		Code:    http.StatusOK,
		Data:    dto.FavoriteCategoriesResponse{Categories: categories},
	})
}

// RecordProductView records a product view for personalization.
// @Summary      Record product view
// @Tags         Products
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Product ID"
// @Success      204 "No Content"
// @Router       /products/{id}/view [post]
func (h *HomeHandler) RecordProductView(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, "authentication required")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid product id")
		return
	}
	if err := h.svc.RecordProductView(c.Request.Context(), userID, uint(id)); err != nil {
		utils.HandleServiceError(c, err, "failed to record view")
		return
	}
	c.Status(http.StatusNoContent)
}

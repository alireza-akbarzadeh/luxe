package handlers

import (
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	appstorefront "github.com/alireza-akbarzadeh/luxe/internal/application/storefront"
	appuserlike "github.com/alireza-akbarzadeh/luxe/internal/application/userlike"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
)

type StorefrontHandler struct {
	service     *appstorefront.Service
	likeQueries *appuserlike.Queries
}

func NewStorefrontHandler(service *appstorefront.Service, likeQueries *appuserlike.Queries) *StorefrontHandler {
	return &StorefrontHandler{service: service, likeQueries: likeQueries}
}

func (ctrl *StorefrontHandler) likedMap(c *gin.Context) map[uint]bool {
	liked := make(map[uint]bool)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return liked
	}
	likedIDs, err := ctrl.likeQueries.GetUserLikedProductIDs(userID)
	if err != nil {
		return liked
	}
	for _, id := range likedIDs {
		liked[id] = true
	}
	return liked
}

func (ctrl *StorefrontHandler) toProductCards(
	c *gin.Context,
	products []*models.Product,
	unitsSold map[uint]int64,
) []dto.StorefrontProductCard {
	liked := ctrl.likedMap(c)
	items := make([]dto.StorefrontProductCard, len(products))
	for i, product := range products {
		card := dto.StorefrontProductCard{
			ProductWithLike: dto.ProductWithLike{
				ProductResponse: dto.ToProductResponse(c.Request.Context(), *product),
				IsLiked:         liked[product.ID],
			},
		}
		if product.CompareAtPrice != nil {
			card.DiscountPercent = computeDiscountPercent(*product.CompareAtPrice, product.Price)
		}
		if unitsSold != nil {
			card.UnitsSold = unitsSold[product.ID]
		}
		items[i] = card
	}
	return items
}

// GetDiscountedProducts returns products with active compare-at pricing.
// @Summary      List discounted products
// @Description  Landing-page carousel feed of products on sale, sorted by discount depth
// @Tags         Storefront
// @Produce      json
// @Param        limit  query int false "Items per page"
// @Param        offset query int false "Offset"
// @Success      200 {object} utils.Response
// @Router       /storefront/discounted-products [get]
func (ctrl *StorefrontHandler) GetDiscountedProducts(c *gin.Context) {
	limit, offset := paginationParams(c, 12)
	products, total, err := ctrl.service.ListDiscountedProducts(c.Request.Context(), limit, offset)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list discounted products")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, gin.H{
		"products": ctrl.toProductCards(c, products, nil),
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

// GetBestSellers returns top-selling active products.
// @Summary      List best-selling products
// @Description  Landing-page feed of best sellers by units sold in the last six months
// @Tags         Storefront
// @Produce      json
// @Param        limit query int false "Max items" default(8)
// @Success      200 {object} utils.Response
// @Router       /storefront/best-sellers [get]
func (ctrl *StorefrontHandler) GetBestSellers(c *gin.Context) {
	limit, _ := paginationParams(c, 8)
	products, unitsSold, err := ctrl.service.ListBestSellers(c.Request.Context(), limit)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list best sellers")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, gin.H{
		"products": ctrl.toProductCards(c, products, unitsSold),
		"total":    len(products),
		"limit":    limit,
	})
}

// GetPopularBrands returns brands with the largest active catalogs.
// @Summary      List popular brands
// @Description  Landing-page brand strip ranked by active product count
// @Tags         Storefront
// @Produce      json
// @Param        limit query int false "Max items" default(10)
// @Success      200 {object} utils.Response
// @Router       /storefront/popular-brands [get]
func (ctrl *StorefrontHandler) GetPopularBrands(c *gin.Context) {
	limit, _ := paginationParams(c, 10)
	brands, err := ctrl.service.ListPopularBrands(c.Request.Context(), limit)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list popular brands")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, gin.H{
		"brands": brands,
		"total":  len(brands),
		"limit":  limit,
	})
}

// GetHomeCategories returns curated categories for the landing page.
// @Summary      List home categories
// @Description  Active top-level categories with cover images and product counts
// @Tags         Storefront
// @Produce      json
// @Param        limit query int false "Max items" default(8)
// @Success      200 {object} utils.Response
// @Router       /storefront/categories [get]
func (ctrl *StorefrontHandler) GetHomeCategories(c *gin.Context) {
	limit, _ := paginationParams(c, 8)
	categories, err := ctrl.service.ListHomeCategories(c.Request.Context(), limit)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list home categories")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, gin.H{
		"categories": categories,
		"total":      len(categories),
		"limit":      limit,
	})
}

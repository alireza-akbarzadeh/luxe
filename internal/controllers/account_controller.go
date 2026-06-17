package controllers

import (
	"strconv"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/middleware"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/services"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/gin-gonic/gin"
)

type AccountController struct {
	addressService services.AddressServiceInterface
	likeService    services.UsertLikeServiceInterface
	orderService   services.OrderServiceInterface
	userService    services.UserServiceInterface
}

func NewAccountController(
	addressService services.AddressServiceInterface,
	likeService services.UsertLikeServiceInterface,
	orderService services.OrderServiceInterface,
	userService services.UserServiceInterface,
) *AccountController {
	return &AccountController{
		addressService: addressService,
		likeService:    likeService,
		orderService:   orderService,
		userService:    userService,
	}
}

// GetAccountSummary returns combined user dashboard data.
// @Summary      Get user dashboard summary
// @Description  Returns user profile, default addresses, address count, liked products count, and recent orders (max 3)
// @Tags         Account
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success 200 {object} utils.Response{data=dto.DashboardSummaryResponse}
// @Failure      401 {object} utils.Response
// @Router       /account/summary [get]
func (ac *AccountController) GetAccountSummary(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	// 1. Get user profile (from userService – you need to inject it)
	user, err := ac.userService.GetUserByID(userID)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch user")
		return
	}

	// 2. Default addresses
	shippingAddr, _ := ac.addressService.GetDefaultAddress(userID, "shipping")
	billingAddr, _ := ac.addressService.GetDefaultAddress(userID, "billing")

	// 3. Address count
	addresses, _ := ac.addressService.List(userID)
	addressCount := len(addresses)

	// 4. Liked products count
	productIDs, _ := ac.likeService.GetUserLikedProductIDs(userID)
	likedCount := len(productIDs)

	// 5. Recent orders
	recentOrders, _, _ := ac.orderService.GetUserOrders(c.Request.Context(), userID, dto.OrderListFilters{Limit: 3, Offset: 0})
	orderDTOs := make([]dto.OrderResponse, len(recentOrders))
	for i, o := range recentOrders {
		orderDTOs[i] = dto.OrderResponse{
			ID:          o.ID,
			OrderNumber: o.OrderNumber,
			Status:      o.Status,
			TotalAmount: o.TotalAmount,
			CreatedAt:   o.CreatedAt,
		}
	}

	resp := dto.DashboardSummaryResponse{
		ID:                     user.ID,
		Email:                  user.Email,
		FirstName:              user.FirstName,
		LastName:               user.LastName,
		Phone:                  user.Phone,
		Role:                   user.Role,
		IsActive:               user.IsActive,
		EmailVerifiedAt:        user.EmailVerifiedAt,
		CreatedAt:              user.CreatedAt,
		DefaultShippingAddress: toAddressDTO(shippingAddr),
		DefaultBillingAddress:  toAddressDTO(billingAddr),
		AddressCount:           addressCount,
		LikedProductsCount:     likedCount,
		RecentOrders:           orderDTOs,
	}
	utils.SuccessResponse(c, "dashboard summary retrieved", resp)
}

// Helper function to convert models.Address to dto.DefaultAddressDTO
func toAddressDTO(addr *models.Address) *dto.DefaultAddressDTO {
	if addr == nil {
		return nil
	}
	return &dto.DefaultAddressDTO{
		ID:           addr.ID,
		AddressLine1: addr.AddressLine1,
		AddressLine2: addr.AddressLine2,
		City:         addr.City,
		State:        addr.State,
		PostalCode:   addr.PostalCode,
		Country:      addr.Country,
		Phone:        addr.Phone,
	}
}

// GetUserOrderAccount returns user's order history with product images and pagination.
// @Summary      Get user order history
// @Description  Returns paginated list of user orders including items with product details
// @Tags         Account
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        limit  query   int  false  "Items per page (default 10, max 50)"
// @Param        offset query   int  false  "Pagination offset"
// @Success      200    {object} utils.Response{data=dto.OrderListResponseData}
// @Failure      401    {object} utils.Response
// @Router       /account/orders [get]
func (ac *AccountController) GetUserOrderAccount(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit > 50 {
		limit = 50
	}
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	orders, total, err := ac.orderService.GetUserOrders(c.Request.Context(), userID, dto.OrderListFilters{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch orders")
		return
	}

	orderDTOs := make([]dto.OrderDetailDTO, len(orders))
	for i, order := range orders {
		itemsDTO := make([]dto.OrderItemDetailDTO, len(order.Items))
		for j, item := range order.Items {
			product := item.Product
			itemsDTO[j] = dto.OrderItemDetailDTO{
				ProductID:   product.ID,
				ProductName: product.Name,
				Quantity:    item.Quantity,
				Price:       item.Price,
				ImageURL:    product.Images[0],
			}
		}
		orderDTOs[i] = dto.OrderDetailDTO{
			ID:          order.ID,
			OrderNumber: order.OrderNumber,
			CreatedAt:   order.CreatedAt,
			Status:      order.Status,
			TotalAmount: order.TotalAmount,
			Items:       itemsDTO,
		}
	}

	data := dto.OrderListResponseData{
		Orders: orderDTOs,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}
	utils.SuccessResponse(c, "order history retrieved", data)
}

// GetUserWishlist returns paginated list of products liked by the user.
// @Summary      Get user's wishlist
// @Description  Returns products the user has liked, with product details (image, price, name)
// @Tags         Account, Wishlist
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        limit  query   int     false  "Items per page (default 10, max 50)"
// @Param        offset query   int     false  "Pagination offset"
// @Param        sort   query   string  false  "Sort order (name, price-asc, price-desc)"
// @Success      200    {object} utils.Response{data=dto.WishlistResponseData}
// @Failure      401    {object} utils.Response
// @Router       /account/wishlist [get]
func (ac *AccountController) GetUserWishlist(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit > 50 {
		limit = 50
	}
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	sortBy := c.Query("sort") // Reads ?sort=price-asc etc.

	// Pass sortBy parameters straight to the service worker
	products, total, err := ac.likeService.GetUserWishlist(userID, limit, offset, sortBy)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch wishlist")
		return
	}

	items := make([]dto.WishlistItemDTO, len(products))
	for i, p := range products {
		imageURL := ""
		if len(p.Images) > 0 {
			imageURL = p.Images[0]
		}

		var oldPrice *float64
		var discountPercent *int
		if p.CompareAtPrice != nil && *p.CompareAtPrice > p.Price {
			oldPrice = p.CompareAtPrice
			percent := int(((*p.CompareAtPrice - p.Price) / *p.CompareAtPrice) * 100)
			discountPercent = &percent
		}

		// 2. Map cleanly to your struct definition fields
		items[i] = dto.WishlistItemDTO{
			ProductID:       p.ID,
			ProductName:     p.Name,
			Price:           p.Price,
			OldPrice:        oldPrice,
			DiscountPercent: discountPercent,
			IsInStock:       p.Stock > 0,
			StockQuantity:   p.Stock,
			Stock:           p.Stock,
			ImageURL:        imageURL,
			Color:           p.Colors,
			Size:            p.Sizes,
		}
	}

	data := dto.WishlistResponseData{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}
	utils.SuccessResponse(c, "wishlist retrieved", data)
}

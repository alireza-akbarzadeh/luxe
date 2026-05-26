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
	"github.com/go-playground/validator/v10"
)

type CartController struct {
	cartService services.CartServiceInterface
	validate    *validator.Validate
}

func NewCartController(cartService services.CartServiceInterface) *CartController {
	return &CartController{
		cartService: cartService,
		validate:    validator.New(),
	}
}

// AddItem adds a product to the cart.
// @Summary      Add item to cart
// @Description  Add a product to the authenticated user's cart
// @Tags         Cart
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body services.AddItemRequest true "Add item"
// @Success      200 {object} dto.AddItemResponse
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Router       /cart/items [post]
func (ctrl *CartController) AddItem(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	var req services.AddItemRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	item, err := ctrl.cartService.AddItem(userID, req)
	if err != nil {
		utils.HandleAppError(c, err, "failed to add item")
		return
	}
	resp := dto.AddItemResponse{
		BaseResponse: dto.BaseResponse{
			Success: true,
			Message: "item added to cart",
			Code:    http.StatusOK,
		},
		Data: dto.CartItemData{
			ID:        item.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		},
	}
	c.JSON(http.StatusOK, resp)
}

// GetCart returns the current user's cart.
// @Summary      Get cart
// @Description  Retrieve all items in the authenticated user's cart
// @Tags         Cart
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} utils.Response{data=object{id=uint,items=[]dto.CartItemDetail,total=float64}}
// @Failure      401 {object} utils.Response
// @Router       /cart [get]
func (ctrl *CartController) GetCart(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	cart, err := ctrl.cartService.GetCart(userID)
	if err != nil {
		utils.InternalServerErrorResponse(c, err, "failed to fetch cart")
		return
	}

	items := make([]dto.CartItemDetail, len(cart.Items))
	var total float64

	for i, item := range cart.Items {
		itemTotal := float64(item.Quantity) * item.Price
		total += itemTotal

		// Original price (compare at price)
		origPrice := 0.0
		if item.Product.CompareAtPrice != nil {
			origPrice = *item.Product.CompareAtPrice
		}

		// Discount amount for the item
		discount := 0.0
		if item.Product.CompareAtPrice != nil && *item.Product.CompareAtPrice > item.Price {
			discount = (*item.Product.CompareAtPrice - item.Price) * float64(item.Quantity)
		}

		// First image
		image := ""
		if len(item.Product.Images) > 0 {
			image = item.Product.Images[0]
		}
		stock := item.Product.Stock
		inStock := stock > 0
		items[i] = dto.CartItemDetail{
			ID:            item.ID,
			ProductID:     item.ProductID,
			Name:          item.Product.Name,
			Quantity:      item.Quantity,
			Price:         item.Price,
			Total:         itemTotal,
			Image:         image,
			OriginalPrice: origPrice,
			Color:         item.Product.Colors,
			Size:          item.Product.Sizes,
			SelectedColor: item.Color,
			SelectedSize:  item.Size,
			Discount:      discount,
			Stock:         stock,
			IsInStock:     inStock,
		}
	}

	data := gin.H{
		"id":    cart.ID,
		"items": items,
		"total": total,
	}

	utils.SuccessResponse(c, "cart retrieved successfully", data)
}

// UpdateItem updates cart item quantity.
// @Summary      Update cart item quantity
// @Description  Change quantity of a specific cart item
// @Tags         Cart
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Cart item ID"
// @Param        request body services.UpdateCartItemRequest true "Update quantity"
// @Success      200 {object} dto.EmptyResponse
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Router       /cart/items/{id} [put]
func (ctrl *CartController) UpdateItem(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	itemID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, 400, "invalid item id")
		return
	}
	var req services.UpdateCartItemRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	if err := ctrl.cartService.UpdateItemQuantity(userID, uint(itemID), req); err != nil {
		utils.HandleAppError(c, err, "failed to update item")
		return
	}

	resp := dto.EmptyResponse{
		BaseResponse: dto.BaseResponse{
			Success: true,
			Message: "cart item updated",
			Code:    http.StatusOK,
		},
	}
	c.JSON(http.StatusOK, resp)
}

// RemoveItem deletes an item from the cart.
// @Summary      Remove cart item
// @Description  Remove a specific item from the cart
// @Tags         Cart
// @Security     BearerAuth
// @Param        id path int true "Cart item ID"
// @Success      200 {object} dto.EmptyResponse
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Router       /cart/items/{id} [delete]
func (ctrl *CartController) RemoveItem(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	itemID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, 400, "invalid item id")
		return
	}

	if err := ctrl.cartService.RemoveItem(userID, uint(itemID)); err != nil {
		utils.HandleAppError(c, err, "failed to remove item")
		return
	}

	resp := dto.EmptyResponse{
		BaseResponse: dto.BaseResponse{
			Success: true,
			Message: "item removed from cart",
			Code:    http.StatusOK,
		},
	}
	c.JSON(http.StatusOK, resp)
}

// ClearCart removes all items from the authenticated user's cart.
// @Summary      Clear cart
// @Description  Delete all items from the active cart
// @Tags         Cart
// @Security     BearerAuth
// @Success      200 {object} dto.EmptyResponse
// @Failure      401 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /cart/items [delete]
func (ctrl *CartController) ClearCart(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	if err := ctrl.cartService.ClearCart(userID); err != nil {
		utils.HandleAppError(c, err, "failed to clear cart")
		return
	}
	utils.SuccessResponse(c, "cart cleared successfully", nil)
}

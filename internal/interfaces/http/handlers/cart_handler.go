package handlers

import (
	"net/http"
	"strconv"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	appcart "github.com/alireza-akbarzadeh/luxe/internal/application/cart"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type CartHandler struct {
	commands *appcart.Commands
	queries  *appcart.Queries
	validate    *validator.Validate
}

func NewCartHandler(commands *appcart.Commands, queries *appcart.Queries) *CartHandler {
	return &CartHandler{
		commands: commands,
		queries:  queries,
		validate: validator.New(),
	}
}

// AddItem adds a product to the cart.
// @Summary      Add item to cart
// @Description  Add a product to the authenticated user's cart
// @Tags         Cart
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body appcart.AddItemRequest true "Add item"
// @Success      200 {object} dto.AddItemResponse
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Router       /cart/items [post]
func (ctrl *CartHandler) AddItem(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	var req appcart.AddItemRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	item, err := ctrl.commands.AddItem(c.Request.Context(), userID, appcart.AddItemInput{ProductID: req.ProductID, Quantity: req.Quantity})
	if err != nil {
		RespondServiceError(c, err, "failed to add item")
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
func (ctrl *CartHandler) GetCart(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	cart, err := ctrl.queries.Get(c.Request.Context(), userID)
	if err != nil {
		RespondServiceError(c, err, "failed to fetch cart")
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
			ProductName:   item.Product.Name,
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
// @Param        request body appcart.UpdateCartItemRequest true "Update quantity"
// @Success      200 {object} dto.EmptyResponse
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Router       /cart/items/{id} [put]
func (ctrl *CartHandler) UpdateItem(c *gin.Context) {
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
	var req appcart.UpdateCartItemRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	if err := ctrl.commands.UpdateItem(c.Request.Context(), userID, uint(itemID), appcart.UpdateItemInput{Quantity: req.Quantity, Color: req.Color, Size: req.Size}); err != nil {
		RespondServiceError(c, err, "failed to update item")
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
func (ctrl *CartHandler) RemoveItem(c *gin.Context) {
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

	if err := ctrl.commands.RemoveItem(c.Request.Context(), userID, uint(itemID)); err != nil {
		RespondServiceError(c, err, "failed to remove item")
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
func (ctrl *CartHandler) ClearCart(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	if err := ctrl.commands.Clear(c.Request.Context(), userID); err != nil {
		RespondServiceError(c, err, "failed to clear cart")
		return
	}
	utils.SuccessResponse(c, "cart cleared successfully", nil)
}

package handlers

import (
	"strconv"

	appcatalog "github.com/alireza-akbarzadeh/luxe/internal/application/catalog"
	appuserlike "github.com/alireza-akbarzadeh/luxe/internal/application/userlike"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type UserLikeHandler struct {
	commands       *appuserlike.Commands
	queries        *appuserlike.Queries
	productService *appcatalog.Service
	validate       *validator.Validate
}

func NewUserLikeHandler(commands *appuserlike.Commands, queries *appuserlike.Queries, productServices *appcatalog.Service) *UserLikeHandler {
	return &UserLikeHandler{
		commands:       commands,
		queries:        queries,
		productService: productServices,
		validate:       validator.New(),
	}
}

// ToggleLike toggles a like on a product for the authenticated user.
// @Summary      Toggle product like
// @Description  Like or unlike a product. Send `{"like": true}` to like, `{"like": false}` to unlike.
// @Tags         Product Likes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path      int     true  "Product ID"
// @Param request body dto.ToggleLikeRequest true "Toggle action"
// @Success 200 {object} utils.Response{data=dto.ToggleLikeResponse}
// @Failure      400     {object}  utils.Response
// @Failure      401     {object}  utils.Response
// @Failure      404     {object}  utils.Response
// @Router       /products/{id}/like [post]
func (ctrl *UserLikeHandler) ToggleLike(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	productID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, 400, "invalid product id")
		return
	}

	var req struct {
		Like *bool `json:"like" validate:"required"`
	}
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	var liked bool
	if req.Like != nil && *req.Like {
		err = ctrl.commands.Like(userID, uint(productID))
		liked = true
	} else if req.Like != nil && !*req.Like {
		err = ctrl.commands.Unlike(userID, uint(productID))
		liked = false
	} else {
		utils.ErrorResponse(c, 400, "like field is required")
		return
	}
	if err != nil {
		utils.HandleServiceError(c, err, "failed to toggle like")
		return
	}

	message := "product unliked successfully"
	if liked {
		message = "product liked successfully"
	}
	utils.SuccessResponse(c, message, dto.ToggleLikeResponse{
		Liked: liked,
	})
}

// IsLikedByUser checks if the current user has liked a specific product.
// @Summary      Check if product is liked
// @Description  Returns whether the authenticated user has liked the given product.
// @Tags         Product Likes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Product ID"
// @Success      200  {object}  utils.Response{data=object{liked=bool}}
// @Failure      400  {object}  utils.Response
// @Failure      401  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Router       /products/{id}/liked [get]
func (ctrl *UserLikeHandler) IsLikedByUser(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	productID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, 400, "invalid product id")
		return
	}

	liked, err := ctrl.queries.IsLikedByUser(userID, uint(productID))
	if err != nil {
		utils.HandleServiceError(c, err, "failed to check like status")
		return
	}

	utils.SuccessResponse(c, "success", gin.H{"liked": liked})
}

// GetUserLikedProductIDs returns all product IDs liked by the current user.
// @Summary      Get user's liked product IDs
// @Description  Returns a list of product IDs that the authenticated user has liked.
// @Tags         Product Likes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  utils.Response{data=object{product_ids=[]int}}
// @Failure      401  {object}  utils.Response
// @Router       /users/me/liked-products [get]
func (ctrl *UserLikeHandler) GetUserLikedProductIDs(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	ids, err := ctrl.queries.GetUserLikedProductIDs(userID)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch liked products")
		return
	}

	utils.SuccessResponse(c, "success", gin.H{"product_ids": ids})
}

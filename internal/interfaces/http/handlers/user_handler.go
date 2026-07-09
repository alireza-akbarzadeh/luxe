package handlers

import (
	"net/http"
	"strconv"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	appaddress "github.com/alireza-akbarzadeh/luxe/internal/application/address"
	appuser "github.com/alireza-akbarzadeh/luxe/internal/application/user"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type UserHandler struct {
	userCommands *appuser.Commands
	userQueries  *appuser.Queries
	addressCommands *appaddress.Commands
	addressQueries  *appaddress.Queries
	validate       *validator.Validate
}

func NewUserHandler(userCommands *appuser.Commands, userQueries *appuser.Queries, addressCommands *appaddress.Commands, addressQueries *appaddress.Queries) *UserHandler {
	return &UserHandler{
		userCommands:    userCommands,
		userQueries:     userQueries,
		addressCommands: addressCommands,
		addressQueries:  addressQueries,
		validate:       validator.New(),
	}
}

// GetProfile returns the authenticated user's profile.
// @Summary      Get user profile
// @Description  Returns the profile of the currently authenticated user including default addresses
// @Tags         User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} utils.Response{data=object{id=uint,email=string,first_name=string,last_name=string,phone=string,role=string,is_active=bool,created_at=string,default_shipping_address=object,default_billing_address=object}}
// @Failure      401 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /profile [get]
func (pc *UserHandler) GetProfile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	user, err := pc.userQueries.GetByID(c.Request.Context(), userID)
	if err != nil {
		utils.HandleServiceError(c, err, constants.ErrInternalServer.Error())
		return
	}

	// Fetch default addresses
	var defaultShipping, defaultBilling interface{}

	shippingAddr, err := pc.addressQueries.GetDefaultAddress(userID, "shipping")
	if err == nil && shippingAddr != nil {
		defaultShipping = gin.H{
			"id":            shippingAddr.ID,
			"address_line1": shippingAddr.AddressLine1,
			"address_line2": shippingAddr.AddressLine2,
			"city":          shippingAddr.City,
			"state":         shippingAddr.State,
			"postal_code":   shippingAddr.PostalCode,
			"country":       shippingAddr.Country,
			"phone":         shippingAddr.Phone,
		}
	}

	billingAddr, err := pc.addressQueries.GetDefaultAddress(userID, "billing")
	if err == nil && billingAddr != nil {
		defaultBilling = gin.H{
			"id":            billingAddr.ID,
			"address_line1": billingAddr.AddressLine1,
			"address_line2": billingAddr.AddressLine2,
			"city":          billingAddr.City,
			"state":         billingAddr.State,
			"postal_code":   billingAddr.PostalCode,
			"country":       billingAddr.Country,
			"phone":         billingAddr.Phone,
		}
	}

	data := gin.H{
		"id":                       user.ID,
		"email":                    user.Email,
		"first_name":               user.FirstName,
		"last_name":                user.LastName,
		"phone":                    user.Phone,
		"avatar_url":               user.AvatarURL,
		"role":                     user.Role,
		"is_active":                user.IsActive,
		"created_at":               user.CreatedAt,
		"default_shipping_address": defaultShipping,
		"default_billing_address":  defaultBilling,
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, data)
}

// UpdateProfile updates the authenticated user's profile.
// @Summary      Update user profile
// @Description  Updates the first name, last name, and phone number of the authenticated user
// @Tags         User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body object true "Profile update data" SchemaExample({"first_name":"John","last_name":"Doe","phone":"+1234567890","avatar_url":"https://cdn.example.com/avatar.jpg"})
// @Success      200 {object} utils.Response{data=object{id=uint,email=string,first_name=string,last_name=string,phone=string}}
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /profile [put]
func (pc *UserHandler) UpdateProfile(c *gin.Context) {
	var req appuser.UpdateProfileRequest
	if !utils.BindAndValidate(c, &req, pc.validate) {
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	user, err := pc.userCommands.UpdateProfile(c.Request.Context(), userID, appuser.UpdateProfileInput{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		AvatarURL: req.AvatarURL,
		Role:      req.Role,
	})
	if err != nil {
		utils.HandleServiceError(c, err, constants.ErrInternalServer.Error())
		return
	}

	data := gin.H{
		"id":         user.ID,
		"email":      user.Email,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"phone":      user.Phone,
		"avatar_url": user.AvatarURL,
		"role":       user.Role,
	}
	utils.SuccessResponse(c, constants.MsgUpdateSuccess, data)
}

// GetAllUsers returns a paginated list of users (admin only).
// @Summary      Get all users
// @Description  Returns a paginated list of all users. Supports advanced filtering.
// @Tags         User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        limit       query  int     false  "Items per page"                       default(20)  minimum(1)  maximum(100)
// @Param        offset      query  int     false  "Offset (skip number of items)"        default(0)   minimum(0)
// @Param        is_active   query  bool    false  "Filter by active status"
// @Param        email       query  string  false  "Partial match on email"
// @Param        phone       query  string  false  "Partial match on phone"
// @Param        first_name  query  string  false  "Partial match on first name"
// @Param        last_name   query  string  false  "Partial match on last name"
// @Param        role        query  string  false  "Exact match on role"                  Enums(user, admin, moderator)
// @Success      200 {object} utils.Response{data=object{users=[]object{id=uint,email=string,first_name=string,last_name=string,phone=string,role=string,is_active=bool,created_at=string},limit=int,offset=int,total=int64}}
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /users [get]
func (pc *UserHandler) GetAllUsers(c *gin.Context) {
	var filter appuser.UserFilter
	if !utils.BindAndValidateQuery(c, &filter, pc.validate) {
		return
	}

	users, total, err := pc.userQueries.List(c.Request.Context(), filter)
	if err != nil {
		utils.HandleServiceError(c, err, constants.ErrUserNotFound)
		return
	}

	// Map to safe response objects (exclude sensitive fields)
	safeUsers := make([]gin.H, len(users))
	for i, u := range users {
		safeUsers[i] = gin.H{
			"id":         u.ID,
			"email":      u.Email,
			"first_name": u.FirstName,
			"last_name":  u.LastName,
			"phone":      u.Phone,
			"role":       u.Role,
			"is_active":  u.IsActive,
			"created_at": u.CreatedAt,
		}
	}

	data := gin.H{
		"users":  safeUsers,
		"limit":  filter.Limit,
		"offset": filter.Offset,
		"total":  total, // total records matching filters (for pagination UI)
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, data)
}

// DeleteUser deletes a user by ID (admin only).
// @Summary      Delete a user
// @Description  Soft‑deletes a user by ID. Only accessible by users with the "admin" role.
// @Tags         User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      401  {object}  utils.Response
// @Failure      403  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /users/{id} [delete]
func (pc *UserHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid user id")
		return
	}

	err = pc.userCommands.Delete(c.Request.Context(), uint(id))
	if err != nil {
		utils.HandleServiceError(c, err, "failed to delete user")
		return
	}

	utils.SuccessResponse(c, constants.MsgDeleteSuccess, nil)
}

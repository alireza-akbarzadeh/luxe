package controllers

import (
	"net/http"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/middleware"
	"github.com/alireza-akbarzadeh/luxe/internal/services"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AddressController struct {
	addressService services.AddressServiceInterface
	validate       *validator.Validate
}

func NewAddressController(svc services.AddressServiceInterface) *AddressController {
	return &AddressController{
		addressService: svc,
		validate:       validator.New(),
	}
}

// Create a new address for the authenticated user.
// @Summary      Create address
// @Description  Adds a new shipping, billing, or both type address for the current user.
// @Tags         Addresses
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CreateAddressRequest true "Address data including address_type (shipping, billing, both)"
// @Success      201 {object} dto.AddressSingleResponse
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /addresses [post]
func (ac *AddressController) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	var req dto.CreateAddressRequest
	if !utils.BindAndValidate(c, &req, ac.validate) {
		return
	}
	address, err := ac.addressService.Create(userID, req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to create address")
		return
	}
	resp := dto.AddressSingleResponse{
		Success: true,
		Message: "address created",
		Code:    http.StatusCreated,
		Address: *address,
	}
	c.JSON(http.StatusCreated, resp)
}

// Update an existing address.
// @Summary      Update address
// @Description  Updates an address by its ID. Only owner can update. Supports partial updates.
// @Tags         Addresses
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path      int                       true  "Address ID"
// @Param        request body      dto.UpdateAddressRequest true  "Fields to update (optional) including address_type (shipping, billing, both)"
// @Success      200     {object}  dto.AddressSingleResponse
// @Failure      400     {object}  utils.Response
// @Failure      401     {object}  utils.Response
// @Failure      404     {object}  utils.Response
// @Failure      500     {object}  utils.Response
// @Router       /addresses/{id} [put]
func (ac *AddressController) Update(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateAddressRequest
	if !utils.BindAndValidate(c, &req, ac.validate) {
		return
	}
	address, err := ac.addressService.Update(id, userID, req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to update address")
		return
	}
	resp := dto.AddressSingleResponse{
		Success: true,
		Message: "address updated",
		Code:    http.StatusOK,
		Address: *address,
	}
	c.JSON(http.StatusOK, resp)
}

// Delete an address.
// @Summary      Delete address
// @Description  Soft deletes an address by ID. Only owner can delete.
// @Tags         Addresses
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Address ID"
// @Success      200  {object}  dto.EmptyResponse
// @Failure      400  {object}  utils.Response
// @Failure      401  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /addresses/{id} [delete]
func (ac *AddressController) Delete(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	err := ac.addressService.Delete(id, userID)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to delete address")
		return
	}
	resp := dto.EmptyResponse{
		BaseResponse: dto.BaseResponse{
			Success: true,
			Message: "address deleted",
			Code:    http.StatusOK,
		},
	}
	c.JSON(http.StatusOK, resp)
}

// List returns all addresses for the authenticated user.
// @Summary      List user addresses
// @Description  Returns all saved addresses for the current user.
// @Tags         Addresses
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  utils.Response{data=object{addresses=[]models.Address}}
// @Failure      401  {object}  utils.Response
// @Router       /addresses [get]
func (ac *AddressController) List(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	addresses, err := ac.addressService.List(userID)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch addresses")
		return
	}
	// Use SuccessResponse for consistency
	utils.SuccessResponse(c, constants.MsgFetchSuccess, gin.H{
		"addresses": addresses,
	})
}

// SetDefault sets an address as the default for its type.
// @Summary      Set default address
// @Description  Marks a specific address as the default (for its address_type, e.g., shipping or billing). Only one default per type.
// @Tags         Addresses
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Address ID"
// @Success      200  {object}  dto.EmptyResponse
// @Failure      400  {object}  utils.Response
// @Failure      401  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /addresses/{id}/default [patch]
func (ac *AddressController) SetDefault(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	err := ac.addressService.SetDefault(id, userID)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to set default address")
		return
	}
	resp := dto.EmptyResponse{
		BaseResponse: dto.BaseResponse{
			Success: true,
			Message: "default address updated",
			Code:    http.StatusOK,
		},
	}
	c.JSON(http.StatusOK, resp)
}

// GetDefault returns the default address of a given type for the user.
// @Summary      Get default address
// @Description  Returns the default shipping or billing address for the current user.
// @Tags         Addresses
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        type query string false "Address type (shipping/billing)" default(shipping) Enums(shipping, billing)
// @Success      200 {object} dto.AddressSingleResponse
// @Failure      401 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Router       /addresses/default [get]
func (ac *AddressController) GetDefault(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	addressType := c.DefaultQuery("type", "shipping")
	if addressType != "shipping" && addressType != "billing" {
		utils.ErrorResponse(c, 400, "type must be shipping or billing")
		return
	}
	addr, err := ac.addressService.GetDefaultAddress(userID, addressType)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to get default address")
		return
	}
	if addr == nil {
		utils.NotFoundResponse(c, "no default address found")
		return
	}
	resp := dto.AddressSingleResponse{
		Success: true,
		Message: constants.MsgFetchSuccess,
		Code:    http.StatusOK,
		Address: *addr,
	}
	c.JSON(http.StatusOK, resp)
}

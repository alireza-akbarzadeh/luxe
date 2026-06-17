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

type CouponController struct {
	couponService services.CouponServiceInterface
	validate      *validator.Validate
}

func NewCouponController(couponService services.CouponServiceInterface) *CouponController {
	return &CouponController{
		couponService: couponService,
		validate:      validator.New(),
	}
}

// Create a new coupon (admin only).
// @Summary      Create coupon
// @Description  Creates a new discount coupon. Only accessible by users with the "admin" role.
// @Tags         Coupons
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CreateCouponRequest true "Coupon creation data"
// @Success      201 {object} dto.CouponSingleResponse
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Failure      409 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /coupons [post]
func (cc *CouponController) Create(c *gin.Context) {
	var req dto.CreateCouponRequest

	if !utils.BindAndValidate(c, &req, cc.validate) {
		return
	}
	coupon, err := cc.couponService.Create(req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to create coupon")
		return
	}
	resp := dto.CouponSingleResponse{
		BaseResponse: dto.BaseResponse{
			Success: true,
			Message: constants.MsgCreateSuccess,
			Code:    http.StatusCreated,
		},
		Data: dto.CouponData{Coupon: *coupon},
	}
	c.JSON(http.StatusCreated, resp)
}

// Update an existing coupon (admin only).
// @Summary      Update coupon
// @Description  Updates a coupon by ID. Only accessible by users with the "admin" role.
// @Tags         Coupons
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path      int                       true  "Coupon ID"
// @Param        request body      dto.UpdateCouponRequest   true  "Coupon update data"
// @Success      200     {object}  dto.CouponSingleResponse
// @Failure      400     {object}  utils.Response
// @Failure      401     {object}  utils.Response
// @Failure      403     {object}  utils.Response
// @Failure      404     {object}  utils.Response
// @Failure      409     {object}  utils.Response
// @Failure      500     {object}  utils.Response
// @Router       /coupons/{id} [put]
func (cc *CouponController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, 400, "invalid coupon id")
		return
	}
	var req dto.UpdateCouponRequest
	if !utils.BindAndValidate(c, &req, cc.validate) {
		return
	}
	coupon, err := cc.couponService.Update(uint(id), req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to update coupon")
		return
	}
	resp := dto.CouponSingleResponse{
		BaseResponse: dto.BaseResponse{
			Success: true,
			Message: constants.MsgUpdateSuccess,
			Code:    http.StatusOK,
		},
		Data: dto.CouponData{Coupon: *coupon},
	}
	c.JSON(http.StatusOK, resp)
}

// Delete a coupon (admin only).
// @Summary      Delete coupon
// @Description  Soft-deletes a coupon by ID. Only accessible by users with the "admin" role.
// @Tags         Coupons
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Coupon ID"
// @Success      200  {object}  dto.EmptyResponse
// @Failure      400  {object}  utils.Response
// @Failure      401  {object}  utils.Response
// @Failure      403  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /coupons/{id} [delete]
func (cc *CouponController) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, 400, "invalid coupon id")
		return
	}
	err = cc.couponService.Delete(uint(id))
	if err != nil {
		utils.HandleServiceError(c, err, "failed to delete coupon")
		return
	}
	resp := dto.EmptyResponse{
		BaseResponse: dto.BaseResponse{
			Success: true,
			Message: constants.MsgDeleteSuccess,
			Code:    http.StatusOK,
		},
	}
	c.JSON(http.StatusOK, resp)
}

// Validate checks if a coupon code is applicable to the user's cart.
// @Summary      Validate coupon
// @Description  Validates a coupon code for the authenticated user's order total. Returns discount amount and final total if valid.
// @Tags         Coupons
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.ValidateRequest true "Coupon validation request"
// @Success      200 {object} dto.CouponValidateResponse
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /coupons/validate [post]
func (cc *CouponController) Validate(c *gin.Context) {
	var req dto.ValidateRequest
	if !utils.BindAndValidate(c, &req, cc.validate) {
		return
	}
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	coupon, discount, err := cc.couponService.ValidateCoupon(req.Code, userID, req.OrderTotal)
	if err != nil {
		utils.HandleServiceError(c, err, "coupon validation failed")
		return
	}
	resp := dto.CouponValidateResponse{
		BaseResponse: dto.BaseResponse{
			Success: true,
			Message: "coupon valid",
			Code:    http.StatusOK,
		},
		Data: dto.CouponValidateData{
			Coupon:         *coupon,
			DiscountAmount: discount,
			FinalTotal:     req.OrderTotal - discount,
		},
	}
	c.JSON(http.StatusOK, resp)
}

// List returns a paginated list of coupons with optional filters (admin only).
// @Summary      List coupons
// @Description  Returns a paginated list of coupons with optional filters (admin only)
// @Tags         Coupons
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        limit         query     int     false  "Items per page"  default(20)  minimum(1)  maximum(100)
// @Param        offset        query     int     false  "Offset (skip number of items)"  default(0)  minimum(0)
// @Param        code          query     string  false  "Filter by coupon code (partial match)"
// @Param        is_active     query     bool    false  "Filter by active status"
// @Param        discount_type query     string  false  "Filter by discount type (percentage/fixed)"
// @Param        start_date    query     string  false  "Filter by start date (ISO 8601)"
// @Param        end_date      query     string  false  "Filter by end date (ISO 8601)"
// @Success      200           {object}  dto.CouponListResponse
// @Failure      400           {object}  utils.Response
// @Failure      401           {object}  utils.Response
// @Failure      403           {object}  utils.Response
// @Failure      500           {object}  utils.Response
// @Router       /coupons [get]
func (cc *CouponController) List(c *gin.Context) {
	var filters dto.CouponListFilters
	if !utils.BindAndValidateQuery(c, &filters, cc.validate) {
		return
	}

	coupons, total, err := cc.couponService.List(filters)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list coupons")
		return
	}
	resp := dto.CouponListResponse{
		BaseResponse: dto.BaseResponse{
			Success: true,
			Message: constants.MsgFetchSuccess,
			Code:    http.StatusOK,
		},
		Data: dto.CouponListData{
			Coupons: coupons,
			Total:   total,
			Limit:   filters.Limit,
			Offset:  filters.Offset,
		},
	}
	c.JSON(http.StatusOK, resp)
}

// GetMyCoupons returns available coupons for the authenticated user
// @Summary      Get my available coupons
// @Description  Retrieve all coupons that are valid and not yet used by the authenticated user
// @Tags         Coupons
// @Produce      json
// @Security     BearerAuth
// @Param        order_total query number false "Order total amount to filter by minimum order requirement"
// @Success      200 {object} utils.Response{data=[]models.Coupon}
// @Failure      401 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /coupons/my [get]
func (cc *CouponController) GetMyCoupons(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	orderTotal := 0.0
	if orderTotalStr := c.Query("order_total"); orderTotalStr != "" {
		parsed, err := strconv.ParseFloat(orderTotalStr, 64)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "invalid order_total parameter")
			return
		}
		orderTotal = parsed
	}
	coupons, err := cc.couponService.GetAvailableCouponsForUser(userID, orderTotal)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch available coupons")
		return
	}

	utils.SuccessResponse(c, "available coupons retrieved successfully", coupons)
}

// GetCouponByID retrieves a single coupon by ID (admin only).
// @Summary      Get a coupon by ID
// @Description  Returns a single coupon by its numeric ID. Only accessible by users with the "admin" role.
// @Tags         Coupons
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Coupon ID"
// @Success      200  {object}  dto.CouponSingleResponse
// @Failure      400  {object}  utils.Response
// @Failure      401  {object}  utils.Response
// @Failure      403  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /coupons/{id} [get]
func (cc *CouponController) GetCouponByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, 400, "invalid coupon ID")
		return
	}

	coupon, err := cc.couponService.GetByID(uint(id))
	if err != nil {
		utils.HandleServiceError(c, err, "failed to find coupon")
		return
	}

	resp := dto.CouponSingleResponse{
		BaseResponse: dto.BaseResponse{
			Success: true,
			Message: constants.MsgFetchSuccess,
			Code:    http.StatusOK,
		},
		Data: dto.CouponData{Coupon: *coupon},
	}
	c.JSON(http.StatusOK, resp)
}
